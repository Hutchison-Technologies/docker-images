package cmd

import (
	"flag"
	"hutchisont/go-deployer/models"
)

var (
	defineInt  = flag.Int
	defineBool = flag.Bool
)

// ParseCMD will parse CMD flags and map to a struct.
// Functions returns `cmd` arguments when complete.
func ParseCMD() models.CMD {
	// Define CMD flags
	maxBuildsInParallel := defineInt("maxBuildsInParallel", 5, "Maximum number of builds to run in parallel")
	maxFunctionDeploymentsInParallel := defineInt("maxFunctionDeploymentsInParallel", 5, "Maximum number of function deployments to run in parallel")
	pollingDelay := defineInt("pollingDelay", 15, "Delay between manual polling")
	delayBetweenBuildsMs := defineInt("delayBetweenBuildsMs", 300, "Delay between builds in ms")
	delayBetweenFunctionsMs := defineInt("delayBetweenFunctionsMs", 300, "Delay between functions in ms")
	runGoBuild := defineBool("runGoBuild", false, "Run go build before deployment")
	runPackageAndPushToRegistry := defineBool("runPackageAndPushToRegistry", false, "Only run the package and push to registry step of the deployment")
	runDeployment := defineBool("runDeployment", false, "Run the deployment step of the deployment")
	imageRegion := flag.String("imageRegion", "", "Region to push to registry in multi region deployments, if not provided will default to provider config region")
	verbose := defineBool("verbose", false, "Verbose mode for logging")

	// Parse CMD flags
	flag.Parse()

	// Format CMD flags
	return models.CMD{
		// Props
		MaxBuildsInParallel:              *maxBuildsInParallel,
		MaxFunctionDeploymentsInParallel: *maxFunctionDeploymentsInParallel,
		PollingDelay:                     *pollingDelay,
		DelayBetweenBuildsMs:             *delayBetweenBuildsMs,
		DelayBetweenFunctionsMs:          *delayBetweenFunctionsMs,
		RunGoBuild:                       *runGoBuild,
		RunPackageAndPushToRegistry:      *runPackageAndPushToRegistry,
		RunDeployment:                    *runDeployment,
		ImageRegion:                      *imageRegion,
		Verbose:                          *verbose,
	}
}
