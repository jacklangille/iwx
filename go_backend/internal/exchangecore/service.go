package exchangecore

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/nats-io/nuid"

	"iwx/go_backend/internal/commands"
	"iwx/go_backend/internal/domain"
	"iwx/go_backend/internal/projectionchange"
	"iwx/go_backend/internal/requestctx"
	"iwx/go_backend/internal/store"
)

type ValidationError struct {
	Errors map[string][]string
}

func (e *ValidationError) Error() string {
	return "validation failed"
}

type Service struct {
	repo           store.ExchangeCoreRepository
	stationCatalog store.StationCatalog
	emitter        *projectionchange.Emitter
}

func NewService(repo store.ExchangeCoreRepository, stationCatalog store.StationCatalog, emitter *projectionchange.Emitter) *Service {
	return &Service{repo: repo, stationCatalog: stationCatalog, emitter: emitter}
}

func (s *Service) SubmitCreateContract(ctx context.Context, command commands.CreateContract) (commands.CreateContractAccepted, error) {
	command.Status = string(domain.ContractStateDraft)
	if err := ValidateCreateContract(command); err != nil {
		return commands.CreateContractAccepted{}, err
	}
	station, err := s.validateAndPopulateStation(ctx, command)
	if err != nil {
		return commands.CreateContractAccepted{}, err
	}
	command.Region = station.Region
	command.DataProviderName = station.ProviderName
	if strings.TrimSpace(command.DataProviderStationMode) == "" {
		command.DataProviderStationMode = "single_station"
	}
	if err := s.rejectDuplicateMarket(ctx, command); err != nil {
		return commands.CreateContractAccepted{}, err
	}

	enqueuedAt := time.Now().UTC().Truncate(time.Millisecond)
	envelope := commands.CreateContractEnvelope{
		CommandID:  nuid.Next(),
		TraceID:    requestctx.TraceID(ctx),
		EnqueuedAt: enqueuedAt,
		Command:    command,
	}

	result, err := s.repo.ProcessCreateContract(ctx, envelope)
	if err != nil {
		return commands.CreateContractAccepted{}, err
	}
	if result.Contract != nil && s.emitter != nil {
		if err := s.emitter.EmitContractChanged(ctx, result.Contract.ID, result.Contract.UpdatedAt); err != nil {
			return commands.CreateContractAccepted{}, err
		}
	}

	_ = result
	return commands.CreateContractAccepted{
		CommandID:  envelope.CommandID,
		Partition:  0,
		Status:     "succeeded",
		EnqueuedAt: enqueuedAt,
	}, nil
}

func (s *Service) GetContractCommand(ctx context.Context, commandID string) (*commands.ContractCommand, error) {
	return s.repo.GetContractCommand(ctx, commandID)
}

func (s *Service) GetContractByID(ctx context.Context, contractID int64) (*domain.Contract, error) {
	return s.repo.GetContract(ctx, contractID)
}

func (s *Service) GetContractRuleByContractID(ctx context.Context, contractID int64) (*domain.ContractRule, error) {
	return s.repo.GetContractRule(ctx, contractID)
}

func ValidateCreateContract(command commands.CreateContract) error {
	errors := map[string][]string{}

	if command.CreatorUserID <= 0 {
		errors["creator_user_id"] = append(errors["creator_user_id"], "must be present")
	}

	validateRequired := func(field, value string) {
		if strings.TrimSpace(value) == "" {
			errors[field] = append(errors[field], "can't be blank")
		}
	}

	validateRequired("name", command.Name)
	validateRequired("metric", command.Metric)
	validateRequired("data_provider_name", command.DataProviderName)
	validateRequired("station_id", command.StationID)
	if status := strings.TrimSpace(command.Status); status != "" && status != string(domain.ContractStateDraft) {
		errors["status"] = append(errors["status"], "must be draft on creation")
	}
	if mode := strings.TrimSpace(command.DataProviderStationMode); mode != "" && mode != "single_station" {
		errors["data_provider_station_mode"] = append(errors["data_provider_station_mode"], "must be single_station for station-backed markets")
	}

	if command.Threshold != nil && *command.Threshold <= 0 {
		errors["threshold"] = append(errors["threshold"], "must be greater than 0")
	}
	if command.Multiplier != nil && *command.Multiplier <= 0 {
		errors["multiplier"] = append(errors["multiplier"], "must be greater than 0")
	}

	validateDateField(errors, "trading_period_start", command.TradingPeriodStart)
	validateDateField(errors, "trading_period_end", command.TradingPeriodEnd)
	validateDateField(errors, "measurement_period_start", command.MeasurementPeriodStart)
	validateDateField(errors, "measurement_period_end", command.MeasurementPeriodEnd)

	validateDateOrder(errors, "trading_period_start", "trading_period_end", command.TradingPeriodStart, command.TradingPeriodEnd)
	validateDateOrder(errors, "measurement_period_start", "measurement_period_end", command.MeasurementPeriodStart, command.MeasurementPeriodEnd)

	if len(errors) > 0 {
		return &ValidationError{Errors: errors}
	}

	return nil
}

func (s *Service) rejectDuplicateMarket(ctx context.Context, command commands.CreateContract) error {
	if s.repo == nil {
		return fmt.Errorf("exchange core repository unavailable")
	}

	duplicate, err := s.repo.FindDuplicateContract(ctx, store.FindDuplicateContractInput{
		ProviderName:           command.DataProviderName,
		StationID:              command.StationID,
		Metric:                 command.Metric,
		Threshold:              command.Threshold,
		TradingPeriodStart:     command.TradingPeriodStart,
		TradingPeriodEnd:       command.TradingPeriodEnd,
		MeasurementPeriodStart: command.MeasurementPeriodStart,
		MeasurementPeriodEnd:   command.MeasurementPeriodEnd,
	})
	if err != nil {
		return err
	}
	if duplicate == nil {
		return nil
	}

	return &ValidationError{Errors: map[string][]string{
		"contract": {fmt.Sprintf("duplicate market already exists as contract %d in status %s", duplicate.ID, duplicate.Status)},
	}}
}

func (s *Service) validateAndPopulateStation(ctx context.Context, command commands.CreateContract) (*domain.WeatherStation, error) {
	if s.stationCatalog == nil {
		return nil, fmt.Errorf("station catalog unavailable")
	}

	station, err := s.stationCatalog.FindStation(ctx, command.DataProviderName, command.StationID)
	if err != nil {
		return nil, err
	}
	if station == nil {
		return nil, &ValidationError{Errors: map[string][]string{
			"station_id": {"is not a known station for the specified provider"},
		}}
	}
	if !station.Active {
		return nil, &ValidationError{Errors: map[string][]string{
			"station_id": {"is not active"},
		}}
	}
	if !stationSupportsMetric(station, command.Metric) {
		return nil, &ValidationError{Errors: map[string][]string{
			"metric": {"is not supported by the selected station"},
		}}
	}
	if region := strings.TrimSpace(command.Region); region != "" && !strings.EqualFold(region, strings.TrimSpace(station.Region)) {
		return nil, &ValidationError{Errors: map[string][]string{
			"region": {"must match the selected station region"},
		}}
	}

	return station, nil
}

func stationSupportsMetric(station *domain.WeatherStation, metric string) bool {
	trimmedMetric := strings.TrimSpace(metric)
	for _, supportedMetric := range station.SupportedMetrics {
		if strings.EqualFold(strings.TrimSpace(supportedMetric), trimmedMetric) {
			return true
		}
	}
	return false
}

func validateDateField(errors map[string][]string, field, value string) {
	if strings.TrimSpace(value) == "" {
		return
	}

	if _, err := parseCommandDate(value); err != nil {
		errors[field] = append(errors[field], "must be in YYYY-MM-DD format")
	}
}

func validateDateOrder(errors map[string][]string, startField, endField, startValue, endValue string) {
	if strings.TrimSpace(startValue) == "" || strings.TrimSpace(endValue) == "" {
		return
	}

	start, err := parseCommandDate(startValue)
	if err != nil {
		return
	}
	end, err := parseCommandDate(endValue)
	if err != nil {
		return
	}

	if end.Before(start) {
		errors[endField] = append(errors[endField], "must be on or after "+startField)
	}
}

func parseCommandDate(value string) (time.Time, error) {
	return time.Parse("2006-01-02", strings.TrimSpace(value))
}

func (s *Service) projectUser(ctx context.Context, userID int64) error {
	if s.emitter == nil || userID <= 0 {
		return nil
	}

	return s.emitter.EmitUserStateChanged(ctx, userID, time.Now().UTC())
}
