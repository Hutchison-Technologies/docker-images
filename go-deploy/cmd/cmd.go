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
	maxDeploymentsInParallel := defineInt("maxDeploymentsInParallel", 5, "Maximum number of deployments to run in parallel")
	verbose := defineBool("verbose", false, "Verbose output")
	pollingDelay := defineInt("pollingDelay", 15, "Delay between manual polling")
	delayBetweenFunctionsMs := defineInt("delayBetweenFunctionsMs", 300, "Delay between functions in ms")

	// Parse CMD flags
	flag.Parse()

	// Format CMD flags
	return models.CMD{
		// Props
		MaxDeploymentsInParallel: *maxDeploymentsInParallel,
		Verbose:                  *verbose,
		PollingDelay:             *pollingDelay,
		DelayBetweenFunctionsMs:  *delayBetweenFunctionsMs,
	}
}
