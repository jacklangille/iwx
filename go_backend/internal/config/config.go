package config

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

type Config struct {
	HTTPAddr                       string
	MatcherHTTPAddr                string
	ExchangeCoreHTTPAddr           string
	ExchangeCoreServiceURL         string
	MatcherServiceURL              string
	AuthHTTPAddr                   string
	OracleHTTPAddr                 string
	OracleServiceURL               string
	AuthDatabaseURL                string
	ExchangeCoreDatabaseURL        string
	OracleDatabaseURL              string
	ReadDatabaseURL                string
	MatcherDatabaseURL             string
	AuthBootstrapUsername          string
	AuthBootstrapPassword          string
	AuthJWTSecret                  string
	AuthJWTIssuer                  string
	AuthJWTTTL                     int
	NATSURL                        string
	NATSPlaceOrderSubject          string
	NATSPlaceOrderStream           string
	NATSExecutionCreatedSubject    string
	NATSExecutionCreatedStream     string
	NATSProjectionChangeSubject    string
	NATSProjectionChangeStream     string
	NATSContractResolvedSubject    string
	NATSContractResolvedStream     string
	NATSSettlementCompletedSubject string
	NATSSettlementCompletedStream  string
	NATSPartitionCount             int
	MatcherOwnedPartitions         []int
	MatcherInstanceID              string
	ExchangeCoreInstanceID         string
	ReadAPIInstanceID              string
	ProjectorInstanceID            string
	SettlementInstanceID           string
}

func FromEnv() Config {
	partitionCount := getenvInt("IWX_MATCHER_PARTITION_COUNT", 16)

	return Config{
		HTTPAddr:                       getenv("IWX_HTTP_ADDR", ":8080"),
		MatcherHTTPAddr:                getenv("IWX_MATCHER_HTTP_ADDR", ":8084"),
		ExchangeCoreHTTPAddr:           getenv("IWX_EXCHANGE_CORE_HTTP_ADDR", ":8082"),
		ExchangeCoreServiceURL:         getenv("IWX_EXCHANGE_CORE_SERVICE_URL", "http://127.0.0.1:8082"),
		MatcherServiceURL:              getenv("IWX_MATCHER_SERVICE_URL", "http://127.0.0.1:8084"),
		AuthHTTPAddr:                   getenv("IWX_AUTH_HTTP_ADDR", ":8081"),
		OracleHTTPAddr:                 getenv("IWX_ORACLE_HTTP_ADDR", ":8083"),
		OracleServiceURL:               getenv("IWX_ORACLE_SERVICE_URL", "http://127.0.0.1:8083"),
		AuthDatabaseURL:                getenv("IWX_AUTH_DATABASE_URL", ""),
		ExchangeCoreDatabaseURL:        getenv("IWX_EXCHANGE_CORE_DATABASE_URL", ""),
		OracleDatabaseURL:              getenv("IWX_ORACLE_DATABASE_URL", ""),
		ReadDatabaseURL:                getenv("IWX_READ_DATABASE_URL", ""),
		MatcherDatabaseURL:             getenv("IWX_MATCHER_DATABASE_URL", ""),
		AuthBootstrapUsername:          getenv("IWX_AUTH_BOOTSTRAP_USERNAME", ""),
		AuthBootstrapPassword:          getenv("IWX_AUTH_BOOTSTRAP_PASSWORD", ""),
		AuthJWTSecret:                  getenv("IWX_AUTH_JWT_SECRET", ""),
		AuthJWTIssuer:                  getenv("IWX_AUTH_JWT_ISSUER", "iwx-go-api"),
		AuthJWTTTL:                     getenvInt("IWX_AUTH_JWT_TTL_SECONDS", 3600),
		NATSURL:                        getenv("IWX_NATS_URL", "nats://127.0.0.1:4222"),
		NATSPlaceOrderSubject:          getenv("IWX_NATS_SUBJECT_PLACE_ORDER", "iwx.matcher.place_order"),
		NATSPlaceOrderStream:           getenv("IWX_NATS_STREAM_PLACE_ORDER", "IWX_PLACE_ORDER"),
		NATSExecutionCreatedSubject:    getenv("IWX_NATS_SUBJECT_EXECUTION_CREATED", "iwx.matcher.execution_created"),
		NATSExecutionCreatedStream:     getenv("IWX_NATS_STREAM_EXECUTION_CREATED", "IWX_EXECUTION_CREATED"),
		NATSProjectionChangeSubject:    getenv("IWX_NATS_SUBJECT_PROJECTION_CHANGE", "iwx.read.change"),
		NATSProjectionChangeStream:     getenv("IWX_NATS_STREAM_PROJECTION_CHANGE", "IWX_PROJECTION_CHANGE"),
		NATSContractResolvedSubject:    getenv("IWX_NATS_SUBJECT_CONTRACT_RESOLVED", "iwx.oracle.contract_resolved"),
		NATSContractResolvedStream:     getenv("IWX_NATS_STREAM_CONTRACT_RESOLVED", "IWX_CONTRACT_RESOLVED"),
		NATSSettlementCompletedSubject: getenv("IWX_NATS_SUBJECT_SETTLEMENT_COMPLETED", "iwx.settlement.completed"),
		NATSSettlementCompletedStream:  getenv("IWX_NATS_STREAM_SETTLEMENT_COMPLETED", "IWX_SETTLEMENT_COMPLETED"),
		NATSPartitionCount:             partitionCount,
		MatcherOwnedPartitions:         getenvPartitions("IWX_MATCHER_OWNED_PARTITIONS", partitionCount),
		MatcherInstanceID:              getenv("IWX_MATCHER_INSTANCE_ID", defaultInstanceID()),
		ExchangeCoreInstanceID:         getenv("IWX_EXCHANGE_CORE_INSTANCE_ID", "exchange-core-"+defaultInstanceID()),
		ReadAPIInstanceID:              getenv("IWX_READ_API_INSTANCE_ID", "read-api-"+defaultInstanceID()),
		ProjectorInstanceID:            getenv("IWX_PROJECTOR_INSTANCE_ID", "projector-"+defaultInstanceID()),
		SettlementInstanceID:           getenv("IWX_SETTLEMENT_INSTANCE_ID", "settlement-"+defaultInstanceID()),
	}
}

func (c Config) ValidateForAuth() error {
	return validateRequired(map[string]string{
		"IWX_AUTH_DATABASE_URL": c.AuthDatabaseURL,
		"IWX_AUTH_JWT_SECRET":   c.AuthJWTSecret,
		"IWX_AUTH_JWT_ISSUER":   c.AuthJWTIssuer,
	})
}

func (c Config) ValidateForReadAPI() error {
	return validateRequired(map[string]string{
		"IWX_READ_DATABASE_URL": c.ReadDatabaseURL,
		"IWX_AUTH_JWT_SECRET":   c.AuthJWTSecret,
		"IWX_AUTH_JWT_ISSUER":   c.AuthJWTIssuer,
	})
}

func (c Config) ValidateForExchangeCore() error {
	return validateRequired(map[string]string{
		"IWX_EXCHANGE_CORE_DATABASE_URL": c.ExchangeCoreDatabaseURL,
		"IWX_ORACLE_SERVICE_URL":         c.OracleServiceURL,
		"IWX_AUTH_JWT_SECRET":            c.AuthJWTSecret,
		"IWX_AUTH_JWT_ISSUER":            c.AuthJWTIssuer,
		"IWX_NATS_URL":                   c.NATSURL,
	})
}

func (c Config) ValidateForMatcher() error {
	return validateRequired(map[string]string{
		"IWX_MATCHER_DATABASE_URL": c.MatcherDatabaseURL,
		"IWX_NATS_URL":             c.NATSURL,
	})
}

func (c Config) ValidateForProjector() error {
	return validateRequired(map[string]string{
		"IWX_READ_DATABASE_URL":         c.ReadDatabaseURL,
		"IWX_EXCHANGE_CORE_SERVICE_URL": c.ExchangeCoreServiceURL,
		"IWX_MATCHER_SERVICE_URL":       c.MatcherServiceURL,
		"IWX_ORACLE_SERVICE_URL":        c.OracleServiceURL,
		"IWX_NATS_URL":                  c.NATSURL,
	})
}

func (c Config) ValidateForOracle() error {
	return validateRequired(map[string]string{
		"IWX_ORACLE_DATABASE_URL":       c.OracleDatabaseURL,
		"IWX_EXCHANGE_CORE_SERVICE_URL": c.ExchangeCoreServiceURL,
		"IWX_NATS_URL":                  c.NATSURL,
	})
}

func (c Config) ValidateForSettlement() error {
	return validateRequired(map[string]string{
		"IWX_EXCHANGE_CORE_SERVICE_URL": c.ExchangeCoreServiceURL,
		"IWX_NATS_URL":                  c.NATSURL,
	})
}

func validateRequired(values map[string]string) error {
	missing := make([]string, 0)
	for key, value := range values {
		if strings.TrimSpace(value) == "" {
			missing = append(missing, key)
		}
	}

	sort.Strings(missing)
	if len(missing) == 0 {
		return nil
	}

	return errors.New("missing required configuration: " + strings.Join(missing, ", "))
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}

func getenvInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}

	return parsed
}

func getenvPartitions(key string, partitionCount int) []int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		partitions := make([]int, 0, partitionCount)
		for partition := 0; partition < partitionCount; partition++ {
			partitions = append(partitions, partition)
		}
		return partitions
	}

	parts := strings.Split(value, ",")
	seen := map[int]struct{}{}
	partitions := make([]int, 0, len(parts))

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}

		parsed, err := strconv.Atoi(trimmed)
		if err != nil || parsed < 0 || parsed >= partitionCount {
			continue
		}

		if _, exists := seen[parsed]; exists {
			continue
		}

		seen[parsed] = struct{}{}
		partitions = append(partitions, parsed)
	}

	sort.Ints(partitions)
	return partitions
}

func defaultInstanceID() string {
	hostname, err := os.Hostname()
	if err != nil || strings.TrimSpace(hostname) == "" {
		return fmt.Sprintf("matcher-%d", os.Getpid())
	}

	return fmt.Sprintf("%s-%d", hostname, os.Getpid())
}
