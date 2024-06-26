package hook

import (
	"github.com/cucumber/godog"
)

func InjectHooks(ctx *godog.ScenarioContext) {
	injectHookSetup(ctx)
	injectHookCleanup(ctx)
}

func injectHookSetup(ctx *godog.ScenarioContext) {
	ctx.Before(injectUnauthKubeconfig)
	ctx.Before(injectHostClient)
	ctx.Before(createAndInjectTestNamespace)
	ctx.Before(injectKubespaceNamespace)
	ctx.Before(injectWorkspacesNamespace)
	ctx.Before(injectScenarioId)
}

func injectHookCleanup(ctx *godog.ScenarioContext) {
	ctx.After(dumpResources)
	ctx.After(deleteTestNamespace)
	ctx.After(deleteResources)
}
