package service_test

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/cf-test-helpers/services"

	"github.com/Infra-Red/cf-postgresql-smoke-tests/retry"
	"github.com/Infra-Red/cf-postgresql-smoke-tests/service/reporter"
)

type retryConfig struct {
	BaselineMilliseconds uint   `json:"baseline_interval_milliseconds"`
	Attempts             uint   `json:"max_attempts"`
	BackoffAlgorithm     string `json:"backoff"`
}

func (rc retryConfig) Backoff() retry.Backoff {
	baseline := time.Duration(rc.BaselineMilliseconds) * time.Millisecond

	algo := strings.ToLower(rc.BackoffAlgorithm)

	switch algo {
	case "linear":
		return retry.Linear(baseline)
	case "exponential":
		return retry.Linear(baseline)
	default:
		return retry.None(baseline)
	}
}

func (rc retryConfig) MaxRetries() int {
	return int(rc.Attempts)
}

type postgresTestConfig struct {
	ServiceName string      `json:"service_name"`
	PlanNames   []string    `json:"plan_names"`
	Retry       retryConfig `json:"retry"`
	KdcServer   string      `json:"kdc_host"`
	KdcRealm    string      `json:"kdc_realm"`
}

func loadCFTestConfig(path string) services.Config {
	config := services.Config{}

	if err := services.LoadConfig(path, &config); err != nil {
		panic(err)
	}

	if err := services.ValidateConfig(&config); err != nil {
		panic(err)
	}

	config.TimeoutScale = 3

	return config
}

func loadPostgresTestConfig(path string) postgresTestConfig {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}

	defer file.Close()

	config := postgresTestConfig{}
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		panic(err)
	}

	return config
}

var (
	configPath     = os.Getenv("CONFIG_PATH")
	cfTestConfig   = loadCFTestConfig(configPath)
	postgresConfig = loadPostgresTestConfig(configPath)

	smokeTestReporter *reporter.SmokeTestReport
)

func TestService(t *testing.T) {
	smokeTestReporter = new(reporter.SmokeTestReport)

	reporter := []Reporter{
		Reporter(smokeTestReporter),
	}

	RegisterFailHandler(Fail)
	RunSpecsWithDefaultAndCustomReporters(t, "PostgreSQL Smoke Tests", reporter)
}
