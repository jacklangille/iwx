package oracle

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"iwx/go_backend/internal/domain"
	"iwx/go_backend/internal/events"
	"iwx/go_backend/internal/readprojection"
	"iwx/go_backend/internal/requestctx"
	"iwx/go_backend/internal/store"
)

type contractRepository interface {
	GetContract(ctx context.Context, contractID int64) (*domain.Contract, error)
	GetContractRule(ctx context.Context, contractID int64) (*domain.ContractRule, error)
	UpdateContractStatus(ctx context.Context, contractID int64, status string) (*domain.Contract, error)
}

type EventPublisher interface {
	PublishContractResolved(ctx context.Context, event events.ContractResolved) error
	Close()
}

type ValidationError struct {
	Errors map[string][]string
}

func (e *ValidationError) Error() string {
	return "validation failed"
}

type Service struct {
	oracleRepo     store.OracleRepository
	contracts      contractRepository
	projector      *readprojection.Projector
	eventPublisher EventPublisher
}

func NewService(
	oracleRepo store.OracleRepository,
	contracts contractRepository,
	projector *readprojection.Projector,
	eventPublisher EventPublisher,
) *Service {
	return &Service{
		oracleRepo:     oracleRepo,
		contracts:      contracts,
		projector:      projector,
		eventPublisher: eventPublisher,
	}
}

func (s *Service) UpsertStation(ctx context.Context, input store.UpsertStationInput) (*domain.WeatherStation, error) {
	if err := validateStationInput(input); err != nil {
		return nil, err
	}

	station, err := s.oracleRepo.UpsertStation(ctx, input)
	if err != nil {
		return nil, err
	}

	if err := s.projectStationCatalog(ctx); err != nil {
		return nil, err
	}

	return station, nil
}

func (s *Service) ListStations(ctx context.Context, activeOnly bool) ([]domain.WeatherStation, error) {
	return s.oracleRepo.ListStations(ctx, activeOnly)
}

func (s *Service) RecordObservation(ctx context.Context, input store.RecordObservationInput) (*domain.OracleObservation, error) {
	if err := validateObservationInput(input); err != nil {
		return nil, err
	}

	contract, err := s.contracts.GetContract(ctx, input.ContractID)
	if err != nil {
		return nil, err
	}
	if contract == nil {
		return nil, fmt.Errorf("contract not found")
	}

	observation, err := s.oracleRepo.RecordObservation(ctx, input)
	if err != nil {
		return nil, err
	}

	return observation, s.projectOracleState(ctx, input.ContractID)
}

func (s *Service) ListObservations(ctx context.Context, contractID int64, limit int) ([]domain.OracleObservation, error) {
	return s.oracleRepo.ListObservations(ctx, contractID, limit)
}

func (s *Service) GetLatestResolution(ctx context.Context, contractID int64) (*domain.ContractResolution, error) {
	return s.oracleRepo.GetLatestResolution(ctx, contractID)
}

func (s *Service) ResolveContract(ctx context.Context, contractID int64) (*domain.ContractResolution, error) {
	contract, err := s.contracts.GetContract(ctx, contractID)
	if err != nil {
		return nil, err
	}
	if contract == nil {
		return nil, fmt.Errorf("contract not found")
	}

	if existing, err := s.oracleRepo.GetLatestResolution(ctx, contractID); err != nil {
		return nil, err
	} else if existing != nil {
		return existing, nil
	}

	rule, err := s.contracts.GetContractRule(ctx, contractID)
	if err != nil {
		return nil, err
	}
	if rule == nil {
		return nil, fmt.Errorf("contract rule not found")
	}
	if rule.Threshold == nil {
		return nil, fmt.Errorf("contract threshold missing")
	}

	observations, err := s.oracleRepo.ListObservations(ctx, contractID, 500)
	if err != nil {
		return nil, err
	}
	selected := latestObservationWithinWindow(observations, contract)
	if selected == nil {
		return nil, fmt.Errorf("no observations available for contract measurement window")
	}

	value, err := strconv.ParseFloat(strings.TrimSpace(selected.NormalizedValue), 64)
	if err != nil {
		return nil, fmt.Errorf("invalid normalized_value: %w", err)
	}

	threshold := float64(*rule.Threshold)
	outcome := resolveOutcome(value, threshold, rule.ResolutionInclusiveSide)

	resolution, err := s.oracleRepo.InsertResolution(ctx, domain.ContractResolution{
		ContractID:             contractID,
		ProviderName:           selected.ProviderName,
		StationID:              selected.StationID,
		ObservedMetric:         selected.ObservedMetric,
		ObservationWindowStart: selected.ObservationWindowStart,
		ObservationWindowEnd:   selected.ObservationWindowEnd,
		RuleVersion:            rule.RuleVersion,
		ResolvedValue:          selected.NormalizedValue,
		Outcome:                outcome,
	})
	if err != nil {
		return nil, err
	}

	if _, err := s.contracts.UpdateContractStatus(ctx, contractID, string(domain.ContractStateResolved)); err != nil {
		return nil, err
	}
	if err := s.projectOracleState(ctx, contractID); err != nil {
		return nil, err
	}
	if err := s.projectContract(ctx, contractID); err != nil {
		return nil, err
	}
	if s.eventPublisher != nil {
		if err := s.eventPublisher.PublishContractResolved(ctx, events.ContractResolved{
			ContractID: contractID,
			Outcome:    string(outcome),
			TraceID:    requestctx.TraceID(ctx),
			ResolvedAt: resolution.ResolvedAt,
		}); err != nil {
			return nil, err
		}
	}

	return resolution, nil
}

func validateObservationInput(input store.RecordObservationInput) error {
	errors := map[string][]string{}
	if input.ContractID <= 0 {
		errors["contract_id"] = append(errors["contract_id"], "must be present")
	}
	if strings.TrimSpace(input.ProviderName) == "" {
		errors["provider_name"] = append(errors["provider_name"], "can't be blank")
	}
	if strings.TrimSpace(input.StationID) == "" {
		errors["station_id"] = append(errors["station_id"], "can't be blank")
	}
	if strings.TrimSpace(input.ObservedMetric) == "" {
		errors["observed_metric"] = append(errors["observed_metric"], "can't be blank")
	}
	if input.ObservationWindowEnd.Before(input.ObservationWindowStart) {
		errors["observation_window_end"] = append(errors["observation_window_end"], "must be on or after observation_window_start")
	}
	if strings.TrimSpace(input.ObservedValue) == "" {
		errors["observed_value"] = append(errors["observed_value"], "can't be blank")
	}
	if strings.TrimSpace(input.NormalizedValue) == "" {
		errors["normalized_value"] = append(errors["normalized_value"], "can't be blank")
	}
	if input.ObservedAt.IsZero() {
		errors["observed_at"] = append(errors["observed_at"], "must be present")
	}
	if len(errors) > 0 {
		return &ValidationError{Errors: errors}
	}
	return nil
}

func validateStationInput(input store.UpsertStationInput) error {
	errors := map[string][]string{}
	if strings.TrimSpace(input.ProviderName) == "" {
		errors["provider_name"] = append(errors["provider_name"], "can't be blank")
	}
	if strings.TrimSpace(input.StationID) == "" {
		errors["station_id"] = append(errors["station_id"], "can't be blank")
	}
	if strings.TrimSpace(input.DisplayName) == "" {
		errors["display_name"] = append(errors["display_name"], "can't be blank")
	}
	if strings.TrimSpace(input.Region) == "" {
		errors["region"] = append(errors["region"], "can't be blank")
	}
	if len(input.SupportedMetrics) == 0 {
		errors["supported_metrics"] = append(errors["supported_metrics"], "must contain at least one metric")
	}
	if len(errors) > 0 {
		return &ValidationError{Errors: errors}
	}
	return nil
}

func latestObservationWithinWindow(observations []domain.OracleObservation, contract *domain.Contract) *domain.OracleObservation {
	if contract == nil || contract.MeasurementPeriodStart == nil || contract.MeasurementPeriodEnd == nil {
		if len(observations) == 0 {
			return nil
		}
		return &observations[0]
	}

	start := contract.MeasurementPeriodStart.UTC()
	end := contract.MeasurementPeriodEnd.UTC().Add(24*time.Hour - time.Nanosecond)

	for _, observation := range observations {
		if observation.ObservedAt.Before(start) || observation.ObservedAt.After(end) {
			continue
		}
		copy := observation
		return &copy
	}

	return nil
}

func resolveOutcome(value, threshold float64, inclusiveSide domain.ClaimSide) domain.ResolutionOutcome {
	if inclusiveSide == domain.ClaimSideAbove {
		if value >= threshold {
			return domain.ResolutionOutcomeAbove
		}
		return domain.ResolutionOutcomeBelow
	}

	if value <= threshold {
		return domain.ResolutionOutcomeBelow
	}
	return domain.ResolutionOutcomeAbove
}

func (s *Service) projectOracleState(ctx context.Context, contractID int64) error {
	if s.projector == nil {
		return nil
	}

	return s.projector.ProjectOracleState(ctx, contractID)
}

func (s *Service) projectContract(ctx context.Context, contractID int64) error {
	if s.projector == nil {
		return nil
	}

	return s.projector.ProjectContract(ctx, contractID)
}

func (s *Service) projectStationCatalog(ctx context.Context) error {
	if s.projector == nil {
		return nil
	}

	return s.projector.ProjectStationCatalog(ctx)
}
