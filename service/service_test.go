package service_test

import (
	"fmt"
	"strings"
	"time"

	"github.com/Infra-Red/cf-postgresql-smoke-tests/postgres"
	"github.com/Infra-Red/cf-postgresql-smoke-tests/service/reporter"
	"github.com/pborman/uuid"

	smokeTestCF "github.com/Infra-Red/cf-postgresql-smoke-tests/cf"
	"github.com/pivotal-cf-experimental/cf-test-helpers/services"

	. "github.com/onsi/ginkgo"
)

var _ = Describe("PostgreSQL Service", func() {
	var (
		testCF = smokeTestCF.CF{
			ShortTimeout: time.Minute * 3,
			LongTimeout:  time.Minute * 15,
			RetryBackoff: postgresConfig.Retry.Backoff(),
			MaxRetries:   postgresConfig.Retry.MaxRetries(),
		}

		retryInterval = time.Second

		appPath             = "../assets/app_sinatra_service"
		serviceInstanceName string
		appName             string
		planName            string

		context services.Context
	)

	BeforeSuite(func() {
		context = services.NewContext(cfTestConfig, "postgres-test")

		createQuotaArgs := []string{
			"-m", "10G",
			"-r", "1000",
			"-s", "100",
			"--allow-paid-service-plans",
		}

		regularContext := context.RegularUserContext()

		beforeSuiteSteps := []*reporter.Step{
			reporter.NewStep(
				"Connect to CloudFoundry",
				testCF.API(cfTestConfig.ApiEndpoint, cfTestConfig.SkipSSLValidation),
			),
			reporter.NewStep(
				"Log in as admin",
				testCF.Auth(cfTestConfig.AdminUser, cfTestConfig.AdminPassword),
			),
			reporter.NewStep(
				"Create 'postgres-smoke-tests' quota",
				testCF.CreateQuota("postgres-smoke-test-quota", createQuotaArgs...),
			),
			reporter.NewStep(
				fmt.Sprintf("Create '%s' org", cfTestConfig.OrgName),
				testCF.CreateOrg(cfTestConfig.OrgName, "postgres-smoke-test-quota"),
			),
			reporter.NewStep(
				fmt.Sprintf("Enable service access for '%s' org", cfTestConfig.OrgName),
				testCF.EnableServiceAccess(cfTestConfig.OrgName, postgresConfig.ServiceName),
			),
			reporter.NewStep(
				fmt.Sprintf("Target '%s' org", cfTestConfig.OrgName),
				testCF.TargetOrg(cfTestConfig.OrgName),
			),
			reporter.NewStep(
				fmt.Sprintf("Create '%s' space", cfTestConfig.SpaceName),
				testCF.CreateSpace(cfTestConfig.SpaceName),
			),
			reporter.NewStep(
				fmt.Sprintf("Create user '%s'", regularContext.Username),
				testCF.CreateUser(regularContext.Username, regularContext.Password),
			),
			reporter.NewStep(
				fmt.Sprintf(
					"Assign user '%s' to 'SpaceManager' role for '%s'",
					regularContext.Username,
					cfTestConfig.SpaceName,
				),
				testCF.SetSpaceRole(regularContext.Username, regularContext.Org, cfTestConfig.SpaceName, "SpaceManager"),
			),
			reporter.NewStep(
				fmt.Sprintf(
					"Assign user '%s' to 'SpaceDeveloper' role for '%s'",
					regularContext.Username,
					cfTestConfig.SpaceName,
				),
				testCF.SetSpaceRole(regularContext.Username, regularContext.Org, cfTestConfig.SpaceName, "SpaceDeveloper"),
			),
			reporter.NewStep(
				fmt.Sprintf(
					"Assign user '%s' to 'SpaceAuditor' role for '%s'",
					regularContext.Username,
					cfTestConfig.SpaceName,
				),
				testCF.SetSpaceRole(regularContext.Username, regularContext.Org, cfTestConfig.SpaceName, "SpaceAuditor"),
			),
			reporter.NewStep(
				"Log out",
				testCF.Logout(),
			),
		}

		smokeTestReporter.RegisterBeforeSuiteSteps(beforeSuiteSteps)

		for _, task := range beforeSuiteSteps {
			task.Perform()
		}
	})

	BeforeEach(func() {
		regularContext := context.RegularUserContext()
		appName = randomName()
		serviceInstanceName = randomName()

		pushArgs := []string{
			"-m", "128M",
			"-p", appPath,
			"-s", "cflinuxfs2",
			"--no-start",
		}

		specSteps := []*reporter.Step{
			reporter.NewStep(
				fmt.Sprintf("Log in as %s", regularContext.Username),
				testCF.Auth(regularContext.Username, regularContext.Password),
			),
			reporter.NewStep(
				fmt.Sprintf("Target '%s' org and '%s' space", cfTestConfig.OrgName, cfTestConfig.SpaceName),
				testCF.TargetOrgAndSpace(cfTestConfig.OrgName, cfTestConfig.SpaceName),
			),
			reporter.NewStep(
				"Push the PostgreSQL sample app to Cloud Foundry",
				testCF.Push(appName, pushArgs...),
			),
		}

		smokeTestReporter.ClearSpecSteps()
		smokeTestReporter.RegisterSpecSteps(specSteps)

		for _, task := range specSteps {
			task.Perform()
		}
	})

	AfterEach(func() {
		specSteps := []*reporter.Step{
			reporter.NewStep(
				fmt.Sprintf("Unbind the %q plan instance", planName),
				testCF.UnbindService(appName, serviceInstanceName),
			),
			reporter.NewStep(
				fmt.Sprintf("Delete the %q plan instance", planName),
				testCF.DeleteService(serviceInstanceName),
			),
			reporter.NewStep(
				fmt.Sprintf("Ensure service instance for plan %q has been deleted", planName),
				testCF.EnsureServiceInstanceGone(serviceInstanceName),
			),
			reporter.NewStep(
				"Delete the app",
				testCF.Delete(appName),
			),
			reporter.NewStep(
				"Log out",
				testCF.Logout(),
			),
			reporter.NewStep(
				"Log in as admin",
				testCF.Auth(cfTestConfig.AdminUser, cfTestConfig.AdminPassword),
			),
			reporter.NewStep(
				"Delete security group 'postgres-smoke-tests-sg'",
				testCF.DeleteSecurityGroup("postgres-smoke-tests-sg"),
			),
			reporter.NewStep(
				"Log out",
				testCF.Logout(),
			),
		}

		smokeTestReporter.RegisterSpecSteps(specSteps)

		for _, task := range specSteps {
			task.Perform()
		}
	})

	AfterSuite(func() {
		regularContext := context.RegularUserContext()

		afterSuiteSteps := []*reporter.Step{
			reporter.NewStep(
				"Connect to CloudFoundry",
				testCF.API(cfTestConfig.ApiEndpoint, cfTestConfig.SkipSSLValidation),
			),
			reporter.NewStep(
				"Log in as admin",
				testCF.Auth(cfTestConfig.AdminUser, cfTestConfig.AdminPassword),
			),
			reporter.NewStep(
				fmt.Sprintf("Target '%s' org and '%s' space", cfTestConfig.OrgName, cfTestConfig.SpaceName),
				testCF.TargetOrgAndSpace(cfTestConfig.OrgName, cfTestConfig.SpaceName),
			),
			reporter.NewStep(
				"Ensure no service-instances left",
				testCF.EnsureAllServiceInstancesGone(),
			),
			reporter.NewStep(
				fmt.Sprintf("Delete user '%s'", regularContext.Username),
				testCF.DeleteUser(regularContext.Username),
			),
			reporter.NewStep(
				fmt.Sprintf("Delete org '%s'", cfTestConfig.OrgName),
				testCF.DeleteOrg(cfTestConfig.OrgName),
			),
			reporter.NewStep(
				"Log out",
				testCF.Logout(),
			),
		}

		smokeTestReporter.RegisterAfterSuiteSteps(afterSuiteSteps)

		for _, task := range afterSuiteSteps {
			task.Perform()
		}
	})

	AssertLifeCycleBehavior := func(planName string) {

		It(strings.ToUpper(planName)+": create, bind to, write to, read from, unbind, and destroy a service instance", func() {
			regularContext := context.RegularUserContext()

			var skip bool

			uri := fmt.Sprintf("https://%s.%s", appName, cfTestConfig.AppsDomain)
			app := postgres.NewApp(uri, testCF.ShortTimeout, retryInterval)

			serviceCreateStep := reporter.NewStep(
				fmt.Sprintf("Create a '%s' plan instance of PostgreSQL", planName),
				testCF.CreateService(postgresConfig.ServiceName, planName, serviceInstanceName, &skip),
			)

			smokeTestReporter.RegisterSpecSteps([]*reporter.Step{serviceCreateStep})

			specSteps := []*reporter.Step{
				reporter.NewStep(
					fmt.Sprintf("Bind the PostgreSQL sample app '%s' to the '%s' plan instance '%s' of PostgreSQL", appName, planName, serviceInstanceName),
					testCF.BindService(appName, serviceInstanceName),
				),
				reporter.NewStep(
					"Log in as admin",
					testCF.Auth(cfTestConfig.AdminUser, cfTestConfig.AdminPassword),
				),
				reporter.NewStep(
					fmt.Sprintf("Target '%s' org and '%s' space", cfTestConfig.OrgName, cfTestConfig.SpaceName),
					testCF.TargetOrgAndSpace(cfTestConfig.OrgName, cfTestConfig.SpaceName),
				),
				reporter.NewStep(
					"Create and bind security group for running smoke tests",
					testCF.CreateAndBindSecurityGroup("postgres-smoke-tests-sg", appName, cfTestConfig.OrgName, cfTestConfig.SpaceName),
				),
				reporter.NewStep(
					fmt.Sprintf("Log in as %s", regularContext.Username),
					testCF.Auth(regularContext.Username, regularContext.Password),
				),
				reporter.NewStep(
					fmt.Sprintf("Target '%s' org and '%s' space", cfTestConfig.OrgName, cfTestConfig.SpaceName),
					testCF.TargetOrgAndSpace(cfTestConfig.OrgName, cfTestConfig.SpaceName),
				),
				reporter.NewStep(
					"Start the app",
					testCF.Start(appName),
				),
				reporter.NewStep(
					"Verify that the app is responding",
					app.IsRunning(),
				),
			}

			smokeTestReporter.RegisterSpecSteps(specSteps)

			serviceCreateStep.Perform()
			serviceCreateStep.Description = fmt.Sprintf("Create a '%s' plan instance of PostgreSQL", planName)

			if skip {
				serviceCreateStep.Result = "SKIPPED"
			} else {
				for _, task := range specSteps {
					task.Perform()
				}
			}
		})
	}

	Context("for each plan", func() {
		for _, planName = range postgresConfig.PlanNames {
			AssertLifeCycleBehavior(planName)
		}
	})
})

func randomName() string {
	return uuid.NewRandom().String()
}
