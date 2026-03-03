package main

import (
	"fmt"
	"hutchisont/go-deployer/cmd"
	"hutchisont/go-deployer/constants"
	"hutchisont/go-deployer/models"
	"hutchisont/go-deployer/utils"
	"os"
	"time"

	yaml "gopkg.in/yaml.v3"
)

// NOTE: The service account need to have `roles/serviceusage.serviceUsageConsumer` and `roles/cloudbuild.builds.viewer` set

func main() {
	// Format deployment start time
	deploymentStartTime := time.Now().UTC()

	// Parse CMD flags
	cmd := cmd.ParseCMD()
	utils.Logger(fmt.Sprintf("CMD: %+v\n", cmd), true)

	utils.Logger("TRACE: Looping through the repo...\n", cmd.Verbose)

	// Loop through all the folders in the repo and get the deployer config file.
	listOfDirs, err := os.ReadDir("./")
	if err != nil {
		utils.Logger(fmt.Sprintf("ERR: %s - %s\n", constants.UnableToReadRepoError, err.Error()), true)
		panic(constants.UnableToReadRepoError)
	}

	// Get provider config
	providerConfigBytes, err := os.ReadFile("provider_config.yml")
	if err != nil {
		utils.Logger(fmt.Sprintf("ERR: %s - %s\n", constants.UnableToReadProviderConfigError, err.Error()), true)
		panic(constants.UnableToReadProviderConfigError)
	}

	utils.Logger("TRACE: Parsing provider config...\n", cmd.Verbose)

	// Unmarshal the provider config
	providerConfig := models.Provider{}
	err = yaml.Unmarshal(providerConfigBytes, &providerConfig)
	if err != nil {
		utils.Logger(fmt.Sprintf("ERR: %s - %s\n", constants.UnableToUnmarshalProviderConfigError, err.Error()), true)
		panic(constants.UnableToUnmarshalProviderConfigError)
	}

	utils.Logger("TRACE: Parsed provider config successfully...\n", cmd.Verbose)

	// Open diff file with git changes
	utils.Logger("TRACE: Reading git diff...\n", cmd.Verbose)
	diffOut, err := os.ReadFile("changes.diff")
	if err != nil {
		utils.Logger(fmt.Sprintf("ERR: %s - %s\n", constants.UnableToReadGitDiffError, err.Error()), true)
		panic(constants.UnableToReadGitDiffError)
	}

	// Parse the git diff output and get a list of functions to deploy
	listOfFunctionsToDeploy, listOfFunctionsToDelete, listOfFoldersToDeploy := utils.ParseDiffFunctions(diffOut, cmd.Verbose)

	// Get the deployer config for the repo
	deployerConfigsForTheRepo, err := utils.GetDeployerConfigsForTheRepo(listOfDirs, listOfFoldersToDeploy, listOfFunctionsToDeploy, listOfFunctionsToDelete, providerConfig, cmd)
	if err != nil {
		utils.Logger(fmt.Sprintf("ERR: %s - %s\n", constants.UnableToGetDeployerConfigsForTheRepoError, err.Error()), true)
		panic(constants.UnableToGetDeployerConfigsForTheRepoError)
	}

	utils.Logger(fmt.Sprintf("TRACE: %d functions to process...\n", len(deployerConfigsForTheRepo)), true)

	credentialsPath := providerConfig.Credentials
	if credentialsPath == "" {
		utils.Logger(fmt.Sprintln(constants.NoCredentialsPathProvidedInProviderConfigError), true)
		panic(constants.NoCredentialsPathProvidedInProviderConfigError)
	}

	utils.Logger("TRACE: Initiating deployment...\n", cmd.Verbose)

	// Handle the deployment in batches
	utils.HandleDeploymentBatches(deployerConfigsForTheRepo, cmd, deploymentStartTime, false)
}
