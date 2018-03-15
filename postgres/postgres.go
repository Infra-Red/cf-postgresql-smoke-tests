package postgres

import (
	"fmt"
	"regexp"
	"time"

	"github.com/Infra-Red/cf-postgresql-smoke-tests/retry"
	"github.com/onsi/gomega/gexec"
	"github.com/pivotal-cf-experimental/cf-test-helpers/runner"
)

type App struct {
	uri          string
	timeout      time.Duration
	retryBackoff retry.Backoff
}

func NewApp(uri string, timeout, retryInterval time.Duration) *App {
	return &App{
		uri:          uri,
		timeout:      timeout,
		retryBackoff: retry.None(retryInterval),
	}
}

func (app *App) IsRunning() func() {
	return func() {
		pingURI := fmt.Sprintf("%s/test", app.uri)

		curlAppFn := func() *gexec.Session {
			fmt.Println("Checking that the app is responding at url: ", pingURI)
			return runner.Curl(pingURI, "-k")
		}

		retry.Session(curlAppFn).WithSessionTimeout(app.timeout).AndBackoff(app.retryBackoff).Until(
			retry.MatchesOutput(regexp.MustCompile("works")),
			`{"FailReason": "Test app deployed but did not respond in time"}`,
		)
	}
}
