package utils

import (
	"fmt"
	"hutchisont/go-deployer/models"
)

// Logger will print a string to console when verbose flag is set.
// Verbose flag can be overwritten (true) to log to console.
func Logger(message string, verbose bool) {
	// Block console logging when not verbose mode
	if !verbose {
		return
	}

	// Write to console
	_, _ = fmt.Printf("%s", message)
}

// PipeOutError will pipe out an error to the error channel
func PipeOutError(errorChannel chan models.DeploymentError, errMessage string, deploymentName string, directoryName string, handler string) {
	// Log error
	Logger(errMessage, true)

	// Format error
	deploymentError := models.DeploymentError{
		ErrorMessage:   errMessage,
		DeploymentName: deploymentName,
		DirectoryName:  directoryName,
		Handler:        handler,
	}

	// Pipe error to the error channel
	errorChannel <- deploymentError
}

func HandleErrorsFromChannel(errorChannel chan models.DeploymentError, verbose bool, formatConfigForSelfHealingCycle bool, deployerConfigsForTheRepo map[string]models.DeployerConfig) map[string]models.DeployerConfig {
	selfHealingDeployerConfigs := map[string]models.DeployerConfig{}

	Logger("---------------------------------------------------------\n", verbose)
	Logger("Deployment failed with the following errors:", verbose)

	failedFunctions := map[string][]string{}

	// Check for errors
	for err := range errorChannel {

		if formatConfigForSelfHealingCycle {
			selfHealingDeployerConfigs[err.DeploymentName] = deployerConfigsForTheRepo[err.DeploymentName]
		}

		failedFunctions[err.DirectoryName] = append(failedFunctions[err.DirectoryName], err.DeploymentName)
		Logger("---------------------------------------------------------\n", verbose)
		Logger(fmt.Sprintf("%+v\n", err), verbose)
		Logger("---------------------------------------------------------\n", verbose)
	}

	// Log all the functions that failed
	Logger("---------------------------------------------------------\n", true)
	Logger("The following Functions failed to deploy:", true)
	for directory, deployments := range failedFunctions {
		Logger("---------------------------------------------------------\n", true)
		Logger(fmt.Sprintf("Directory: %s\n", directory), true)

		for _, deploymentName := range deployments {
			Logger(fmt.Sprintf("  - %s\n", deploymentName), true)
		}

		Logger("---------------------------------------------------------\n", true)
	}

	Logger("---------------------------------------------------------\n", true)

	return selfHealingDeployerConfigs
}

func HandleBuildErrorsFromChannel(errorChannel chan models.DeploymentError, verbose bool, curateListOfFoldersForSelfHealingCycle bool) []string {
	selfHealingFoldersToBuild := []string{}

	Logger(fmt.Sprintln("ERR: Package and push failed with the following errors:"), true)

	// Check for errors
	for err := range errorChannel {
		if curateListOfFoldersForSelfHealingCycle {
			selfHealingFoldersToBuild = append(selfHealingFoldersToBuild, err.DirectoryName)
		}

		Logger("---------------------------------------------------------\n", true)
		Logger(fmt.Sprintf("%+v\n", err), true)
		Logger("---------------------------------------------------------\n", true)
	}

	Logger("TRACE: Curated list of folders for Self Healing Cycle...\n", true)

	return selfHealingFoldersToBuild
}
