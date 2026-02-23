package cmd

import (
	"flag"
	"hutchisont/go-deployer/models"
)

var (
	defineInt = flag.Int
)

// ParseCMD will parse CMD flags and map to a struct.
// Functions returns `cmd` arguments when complete.
func ParseCMD() models.CMD {
	// Define CMD flags
	maxDeploymentsInParallel := defineInt("maxDeploymentsInParallel", 5, "Maximum number of deployments to run in parallel")

	// Parse CMD flags
	flag.Parse()

	// Format CMD flags
	return models.CMD{
		// Props
		MaxDeploymentsInParallel: *maxDeploymentsInParallel,
	}
}
