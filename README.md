PostgreSQL Service Broker Smoke Tests
======================================

## Running the tests

### Set up your `go` environment

Set up your golang development environment, [per golang.org](http://golang.org/doc/install).

See [Go CLI](https://github.com/cloudfoundry/cli) for instructions on
installing the go version of `cf`.

Make sure that [curl](http://curl.haxx.se/) is installed on your system.

Make sure that the go version of `cf` is accessible in your `$PATH`.

All `go` dependencies required by the smoke tests are vendored in
`cf-postgresql-smoke-tests/vendor`.

### Test Setup

To run the PostgreSQL Service Broker Smoke Tests, you will need:
- a running CF instance
- an environment variable `$CONFIG_PATH` which points to a `.json` file that
contains the CF settings

Below is an example `integration_config.json`:
```json
{
  "service_name": "a.postgresql",
  "plan_names": ["standard"],
  "retry": {
    "max_attempts": 10,
    "backoff": "linear",
    "baseline_interval_milliseconds": 1000
  },
  "apps_domain": "bosh-lite.com",
  "system_domain": "bosh-lite.com",
  "api": "api.bosh-lite.com",
  "admin_user": "admin",
  "admin_password": "admin",
  "space_name": "postgres-test-space",
  "org_name": "postgres-test-org",
  "skip_ssl_validation": true,
  "create_permissive_security_group": false
}
```

### Test Execution

To execute the tests, run:

```bash
./bin/test
```

Internally the `bin/test` script runs tests using [ginkgo](https://github.com/onsi/ginkgo).