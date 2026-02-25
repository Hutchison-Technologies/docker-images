package main

import (
	"bufio"
	"bytes"
	"fmt"
	"hutchisont/go-deployer/cmd"
	"hutchisont/go-deployer/constants"
	"hutchisont/go-deployer/models"
	"hutchisont/go-deployer/utils"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/go-set/v2"
	yaml "gopkg.in/yaml.v3"
)

// NOTE: The service account need to have `roles/serviceusage.serviceUsageConsumer` set

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
	listOfFunctionsToDeploy, listOfFunctionsToDelete, listOfFoldersToDeploy := parseDiffFunctions(diffOut, cmd.Verbose)

	// Get the deployer config for the repo
	deployerConfigsForTheRepo, err := getDeployerConfigsForTheRepo(listOfDirs, listOfFoldersToDeploy, listOfFunctionsToDeploy, listOfFunctionsToDelete, providerConfig, cmd.Verbose)
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

	utils.Logger("TRACE: Formatting inputs for deployment...\n", cmd.Verbose)

	errorChannel := make(chan models.DeploymentError, len(deployerConfigsForTheRepo))

	utils.Logger(fmt.Sprintf("TRACE: Starting batch deployment of %d in parallel...\n", cmd.MaxDeploymentsInParallel), true)

	batchSize := cmd.MaxDeploymentsInParallel
	var currentBatch []models.DeployerConfig
	batchCounter := 0

	for i, deployerConfigForFunction := range deployerConfigsForTheRepo {
		currentBatch = append(currentBatch, deployerConfigForFunction)
		batchCounter++

		if batchCounter == batchSize {
			// Process the batch
			processDeploymentBatch(currentBatch, errorChannel, cmd.DelayBetweenBatches, cmd.Verbose, deploymentStartTime)

			utils.Logger(fmt.Sprintf("TRACE: Processed %d out of %d functions...\n", i+1, len(deployerConfigsForTheRepo)), true)

			// Reset batch
			currentBatch = nil
			batchCounter = 0
		}
	}

	// Process the last batch
	if batchCounter > 0 {
		processDeploymentBatch(currentBatch, errorChannel, cmd.DelayBetweenBatches, cmd.Verbose, deploymentStartTime)
	}

	utils.Logger("TRACE: Closing error channel...\n", cmd.Verbose)
	close(errorChannel)

	if len(errorChannel) == 0 {
		utils.Logger("TRACE: Deployment successfully completed.", true)
		return
	}

	utils.Logger("---------------------------------------------------------\n", cmd.Verbose)
	utils.Logger("Deployment failed with the following errors:", cmd.Verbose)

	failedFunctions := map[string][]string{}

	// Check for errors
	for err := range errorChannel {
		failedFunctions[err.DirectoryName] = append(failedFunctions[err.DirectoryName], err.DeploymentName)
		utils.Logger("---------------------------------------------------------\n", cmd.Verbose)
		utils.Logger(fmt.Sprintf("%+v\n", err), cmd.Verbose)
		utils.Logger("---------------------------------------------------------\n", cmd.Verbose)
	}

	// Log all the functions that failed
	utils.Logger("---------------------------------------------------------\n", true)
	utils.Logger("The following Functions failed to deploy:", true)
	for directory, deployments := range failedFunctions {
		utils.Logger("---------------------------------------------------------\n", true)
		utils.Logger(fmt.Sprintf("Directory: %s\n", directory), true)

		for _, deploymentName := range deployments {
			utils.Logger(fmt.Sprintf("  - %s\n", deploymentName), true)
		}

		utils.Logger("---------------------------------------------------------\n", true)
	}

	utils.Logger("---------------------------------------------------------\n", true)

	panic(constants.DeploymentFailedError)
}

func parseDiffFunctions(diff []byte, verbose bool) ([]string, []string, []string) {
	functionsToBeAdded := set.From([]string{})
	functionsToBeDeleted := set.From([]string{})
	foldersToDeploy := set.From([]string{})

	scanner := bufio.NewScanner(bytes.NewReader(diff))

	funcRegex := regexp.MustCompile(`\bfunc\s*(?:\([^\)]*\)\s*)?(\w+)\s*\(`)
	currentFile := ""

	// Loop through the diff file lines
	for scanner.Scan() {
		line := scanner.Text()

		// Get the folder if go.mod or go.sum was modified
		if strings.HasPrefix(line, "diff --git") {
			parts := strings.Split(line, " ")
			if len(parts) >= 3 {
				currentFile = strings.TrimPrefix(parts[2], "a/")
			}

			if strings.HasSuffix(currentFile, "go.mod") || strings.HasSuffix(currentFile, "go.sum") {
				dir := filepath.Dir(currentFile)
				foldersToDeploy.Insert(dir)
			}

			continue
		}

		// Added functions
		if strings.HasPrefix(line, "+func ") || strings.HasPrefix(line, "+ func") {
			matches := funcRegex.FindStringSubmatch(line[1:])
			if len(matches) > 1 && matches[1] != "" {
				funcName := matches[1]
				functionsToBeAdded.Insert(funcName)
			}
		}

		// Function definition modified in-place
		if strings.HasPrefix(line, "@@") {
			// Extract the part after @@ @@
			parts := strings.SplitN(line, "@@", 3)

			if len(parts) == 3 {
				trailingCode := strings.TrimSpace(parts[2])
				matches := funcRegex.FindStringSubmatch(trailingCode)

				if len(matches) > 1 && matches[1] != "" {
					funcName := matches[1]
					// Add to added functions
					functionsToBeAdded.Insert(funcName)
				}
			}
		}

		// Deleted functions
		if strings.HasPrefix(line, "-func ") || strings.HasPrefix(line, "- func") {
			matches := funcRegex.FindStringSubmatch(line[1:])
			if len(matches) > 1 && matches[1] != "" {
				funcName := matches[1]
				functionsToBeDeleted.Insert(funcName)
			}
		}
	}

	// Remove modified funcs from the delete list (appearing in both add and delete)
	for _, addedFunc := range functionsToBeAdded.Slice() {
		if functionsToBeDeleted.Contains(addedFunc) {
			functionsToBeDeleted.Remove(addedFunc)
		}
	}

	utils.Logger(fmt.Sprintf("TRACE: Found %d function(s) updated: %+v\n", functionsToBeAdded.Size(), functionsToBeAdded), verbose)
	utils.Logger(fmt.Sprintf("TRACE: Found %d function(s) removed: %+v\n", functionsToBeDeleted.Size(), functionsToBeDeleted), verbose)
	utils.Logger(fmt.Sprintf("TRACE: Folder(s) to deploy as the go.mod/go.sum files were updated: %+v\n", foldersToDeploy), verbose)

	return functionsToBeAdded.Slice(), functionsToBeDeleted.Slice(), foldersToDeploy.Slice()
}

func getDeployerConfigsForTheRepo(listOfDirs []os.DirEntry, listOfFoldersToDeploy []string, listOfFunctionsToDeploy []string, listOfFunctionsToDelete []string, providerConfig models.Provider, verbose bool) ([]models.DeployerConfig, error) {
	deployerConfigsForTheRepo := []models.DeployerConfig{}

	for _, dir := range listOfDirs {
		dirName := dir.Name()

		// Ignore hidden directories
		if dirName == "token" || strings.Contains(dirName, ".") || strings.Contains(dirName, "deploy") || strings.Contains(dirName, "Jenkinsfile") {
			utils.Logger(fmt.Sprintf("TRACE: Skipping directory - %s\n", dirName), true)
			continue
		}

		utils.Logger(fmt.Sprintf("TRACE: Found directory - %s\n", dirName), verbose)

		utils.Logger("TRACE: Running go mod tidy...\n", verbose)
		// Run go mod tidy inside the dir
		cmdStruct := exec.Command("go", "mod", "tidy")
		cmdStruct.Dir = dirName
		out, err := cmdStruct.CombinedOutput()
		if err != nil {
			utils.Logger(fmt.Sprintf("ERR: Unable to process %s - %s\n", dirName, string(out)), true)
			return nil, err
		}

		utils.Logger("TRACE: Running go mod vendor...\n", verbose)
		// Run go mod vendor inside the dir
		cmdStruct = exec.Command("go", "mod", "vendor")
		cmdStruct.Dir = dirName
		out, err = cmdStruct.CombinedOutput()
		if err != nil {
			utils.Logger(fmt.Sprintf("ERR: Unable to process %s - %s\n", dirName, string(out)), true)
			return nil, err
		}

		// For each dir, cd into it and get the deployer config file
		configFile, err := os.ReadFile(dirName + "/deployer_config.yml")
		if err != nil {
			utils.Logger(fmt.Sprintf("ERR: Unable to read deployer config file - %s\n", err.Error()), true)
			return nil, err
		}

		// Unmarshal the config file
		functionsConfig := map[string]models.Function{}
		err = yaml.Unmarshal(configFile, &functionsConfig)
		if err != nil {
			utils.Logger(fmt.Sprintf("ERR: Unable to unmarshal functions config file - %s\n", err.Error()), true)
			return nil, err
		}

		for functionDeploymentName, functionConfig := range functionsConfig {
			if slices.Contains(listOfFunctionsToDelete, functionConfig.Handler) {
				// Add the function to the deployer config to be deleted
				deployerConfigsForTheRepo = append(deployerConfigsForTheRepo, models.DeployerConfig{
					IsDelete:               true,
					DeploymentName:         functionDeploymentName,
					DirectoryName:          dirName,
					Provider:               providerConfig,
					Handler:                functionConfig.Handler,
					MemorySize:             functionConfig.MemorySize,
					Timeout:                functionConfig.Timeout,
					EnvironmentForFunction: functionConfig.EnvironmentForFunction,
				})
			}

			if slices.Contains(listOfFunctionsToDeploy, functionConfig.Handler) ||
				(listOfFoldersToDeploy != nil && slices.Contains(listOfFoldersToDeploy, dirName)) {
				// Skip function if it is to be deleted
				if slices.Contains(listOfFunctionsToDelete, functionConfig.Handler) {
					continue
				}

				// Add the function to the deployer config
				deployerConfigsForTheRepo = append(deployerConfigsForTheRepo, models.DeployerConfig{
					IsDelete:               false,
					DeploymentName:         functionDeploymentName,
					DirectoryName:          dirName,
					Provider:               providerConfig,
					Handler:                functionConfig.Handler,
					MemorySize:             functionConfig.MemorySize,
					Timeout:                functionConfig.Timeout,
					EnvironmentForFunction: functionConfig.EnvironmentForFunction,
				})
			}
		}
	}

	// Return the deployer configs
	return deployerConfigsForTheRepo, nil
}

func processDeploymentBatch(deploymentBatch []models.DeployerConfig, errorChannel chan models.DeploymentError, delayBetweenBatches int, verbose bool, deploymentStartTime time.Time) {
	var wg sync.WaitGroup
	wg.Add(len(deploymentBatch))

	for _, deployConfig := range deploymentBatch {
		go deployFunction(deployConfig, &wg, errorChannel, verbose, deploymentStartTime)
	}

	wg.Wait()

	// Sleep between batches
	time.Sleep(time.Duration(delayBetweenBatches) * time.Second)
}

func deployFunction(deployerConfigForFunction models.DeployerConfig, wg *sync.WaitGroup, errorChannel chan models.DeploymentError, verbose bool, deploymentStartTime time.Time) {
	defer wg.Done()

	// Create isolated gcloud config directory
	tempDir, err := os.MkdirTemp("", "gcloud-*")
	if err != nil {
		errMessage := fmt.Sprintf("ERR: Unable to create temp gcloud dir: %s", err.Error())
		pipeOutError(errorChannel, errMessage, deployerConfigForFunction.DeploymentName, deployerConfigForFunction.DirectoryName, deployerConfigForFunction.Handler)

		return
	}

	defer func(errorChannel chan models.DeploymentError) {
		err := os.RemoveAll(tempDir)
		if err != nil {
			errMessage := fmt.Sprintf("ERR: Unable to remove temp gcloud dir: %s", err.Error())
			pipeOutError(errorChannel, errMessage, deployerConfigForFunction.DeploymentName, deployerConfigForFunction.DirectoryName, deployerConfigForFunction.Handler)

			return
		}
	}(errorChannel)

	cmdStruct := exec.Cmd{}

	if deployerConfigForFunction.IsDelete {
		utils.Logger(fmt.Sprintf("TRACE: Deleting %s...\n", deployerConfigForFunction.Handler), true)

		// Format cmd args
		cmdArgs := []string{
			"run", "services",
			"delete",
			deployerConfigForFunction.DeploymentName,
			"--region", deployerConfigForFunction.Provider.Region,
			"--project", deployerConfigForFunction.Provider.Project,
			"--quiet",
			"--service-account", deployerConfigForFunction.Provider.ServiceAccountEmail,
			"--impersonate-service-account", deployerConfigForFunction.Provider.ServiceAccountEmail,
		}

		// Log CMD args
		utils.Logger(fmt.Sprintf("TRACE: Executing command - %s\n", strings.Join(cmdArgs, " ")), verbose)

		// Format the delete command
		cmdStruct = *exec.Command("gcloud", cmdArgs...)

	} else {
		utils.Logger(fmt.Sprintf("TRACE: Deploying %s...\n", deployerConfigForFunction.Handler), true)

		// Merge global and local env vars
		mergedEnv := map[string]string{}

		// Add global env
		for key, value := range deployerConfigForFunction.Provider.Environment {
			mergedEnv[key] = value
		}

		// Override with function level env
		for key, value := range deployerConfigForFunction.EnvironmentForFunction {
			mergedEnv[key] = value
		}

		// Format final env vars
		var envVars []string
		for key, value := range mergedEnv {
			envVars = append(envVars, fmt.Sprintf("%s=%s", key, value))
		}

		envVarsArg := strings.Join(envVars, ",")

		// Format label
		label := fmt.Sprintf("service=%s", strings.ToLower(deployerConfigForFunction.DirectoryName))

		// Format cmd
		cmdArgs := []string{
			"run", "deploy", deployerConfigForFunction.DeploymentName,
			"--source", deployerConfigForFunction.DirectoryName,
			"--function", deployerConfigForFunction.Handler,
			"--update-labels", label,
			"--base-image", deployerConfigForFunction.Provider.Runtime,
			"--memory", deployerConfigForFunction.MemorySize + "Mi",
			"--region", deployerConfigForFunction.Provider.Region,
			"--project", deployerConfigForFunction.Provider.Project,
			"--quiet",
			"--service-account", deployerConfigForFunction.Provider.ServiceAccountEmail,
			"--impersonate-service-account", deployerConfigForFunction.Provider.ServiceAccountEmail,
		}

		// Add timeout if provided
		if deployerConfigForFunction.Timeout != "" {
			cmdArgs = append(cmdArgs, "--timeout", deployerConfigForFunction.Timeout)
		}

		// TODO add vpc connector, firewall rule, network and subnet once we have the shared vpc setup

		// Add env vars
		if len(envVars) > 0 {
			cmdArgs = append(cmdArgs, "--set-env-vars", envVarsArg)
		}

		// Log CMD args
		utils.Logger(fmt.Sprintf("TRACE: Executing command - %s\n", strings.Join(cmdArgs, " ")), verbose)

		// Format the deploy command
		cmdStruct = *exec.Command("gcloud", cmdArgs...)
	}

	cmdStruct.Env = append(os.Environ(),
		"CLOUDSDK_CONFIG="+tempDir,
		"GOOGLE_APPLICATION_CREDENTIALS="+deployerConfigForFunction.Provider.Credentials,
	)

	// Run the gcloud run deploy command
	err = cmdStruct.Start()
	if err != nil {
		// Format errMessage
		errMessage := fmt.Sprintf("ERR: Unable to run deploy command (Function: %s) (isDelete: %t) - %s\n", deployerConfigForFunction.Handler, deployerConfigForFunction.IsDelete, err.Error())
		pipeOutError(errorChannel, errMessage, deployerConfigForFunction.DeploymentName, deployerConfigForFunction.DirectoryName, deployerConfigForFunction.Handler)

		return
	}

	// Handle polling
	if deployerConfigForFunction.IsDelete {
		handlePollingForDeletion(deployerConfigForFunction, errorChannel, tempDir, verbose)
	} else {
		handlePollingForDeployment(deployerConfigForFunction, errorChannel, tempDir, verbose, deploymentStartTime)
	}
}

func pipeOutError(errorChannel chan models.DeploymentError, errMessage string, deploymentName string, directoryName string, handler string) {
	// Log error
	utils.Logger(errMessage, true)

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

func handlePollingForDeployment(deployerConfigForFunction models.DeployerConfig, errorChannel chan models.DeploymentError, tempDir string, verbose bool, deploymentStartTime time.Time) {
	// Format filter
	filter := fmt.Sprintf("createTime>=%s AND tags=service_%s", deploymentStartTime.Format(time.RFC3339), deployerConfigForFunction.DeploymentName)

	// Format get build args
	getBuildArgs := []string{
		"builds", "list",
		"--region", deployerConfigForFunction.Provider.Region,
		"--project", deployerConfigForFunction.Provider.Project,
		"--filter", filter,
		"--sort-by", "~createTime",
		"--limit", "1",
		"--format=value(ID)",
		"--verbosity", "error",
	}

	// Log CMD args
	utils.Logger(fmt.Sprintf("TRACE: Executing get builds list command - %s\n", strings.Join(getBuildArgs, " ")), verbose)

	buildID := ""

	cloudBuildPollingStartTime := time.Now().UTC()

	// Poll every 5 seconds for the build ID
	for {
		// Return if timeout is more than 15 minutes
		if time.Since(cloudBuildPollingStartTime) > time.Duration(constants.POLLING_TIMEOUT)*time.Second {
			// Format error
			errMessage := fmt.Sprintf("ERR: Timeout (Function: %s) (isDelete: %t)\n", deployerConfigForFunction.Handler, deployerConfigForFunction.IsDelete)
			pipeOutError(errorChannel, errMessage, deployerConfigForFunction.DeploymentName, deployerConfigForFunction.DirectoryName, deployerConfigForFunction.Handler)

			return
		}

		// Format the build polling command
		buildCmdStruct := exec.Command("gcloud", getBuildArgs...)

		buildCmdStruct.Env = append(os.Environ(),
			"CLOUDSDK_CONFIG="+tempDir,
			"GOOGLE_APPLICATION_CREDENTIALS="+deployerConfigForFunction.Provider.Credentials,
		)

		// Execute the build polling
		buildOut, err := buildCmdStruct.CombinedOutput()
		if err != nil {
			errMessage := fmt.Sprintf("ERR: Failed to fetch cloud build ID: %s - %s\n", string(buildOut), err)
			pipeOutError(errorChannel, errMessage, deployerConfigForFunction.DeploymentName, deployerConfigForFunction.DirectoryName, deployerConfigForFunction.Handler)

			return
		}

		// Parse build ID
		buildID = strings.TrimSpace(string(buildOut))

		if buildID != "" {
			// Log build ID
			utils.Logger(fmt.Sprintf("TRACE: Initiated (buildID: %s) (Function: %s) (isDelete: %t)\n", buildID, deployerConfigForFunction.Handler, deployerConfigForFunction.IsDelete), verbose)
			break
		}

		// Sleep for 5 seconds
		utils.Logger(fmt.Sprintf("TRACE: Waiting to get buildID (Function: %s) (isDelete: %t)...\n", deployerConfigForFunction.Handler, deployerConfigForFunction.IsDelete), verbose)
		time.Sleep(5 * time.Second)
	}

	// Fomart cmd for status polling
	pollingCmd := []string{
		"builds",
		"describe",
		buildID,
		"--format=value(status)",
		"--region", deployerConfigForFunction.Provider.Region,
		"--project", deployerConfigForFunction.Provider.Project,
	}

	pollingStartTime := time.Now().UTC()

	// Poll every 5 seconds for the build status
	for {
		// Return if timeout is more than 15 minutes
		if time.Since(pollingStartTime) > time.Duration(constants.POLLING_TIMEOUT)*time.Second {
			// Format error
			errMessage := fmt.Sprintf("ERR: Timeout (Function: %s) (isDelete: %t)\n", deployerConfigForFunction.Handler, deployerConfigForFunction.IsDelete)
			pipeOutError(errorChannel, errMessage, deployerConfigForFunction.DeploymentName, deployerConfigForFunction.DirectoryName, deployerConfigForFunction.Handler)

			return
		}

		// Format the polling command
		pollingCmdStruct := exec.Command("gcloud", pollingCmd...)

		pollingCmdStruct.Env = append(os.Environ(),
			"CLOUDSDK_CONFIG="+tempDir,
			"GOOGLE_APPLICATION_CREDENTIALS="+deployerConfigForFunction.Provider.Credentials,
		)

		// Execute the polling
		statusBytes, err := pollingCmdStruct.CombinedOutput()
		if err != nil {
			errMessage := fmt.Sprintf("ERR: Unable to poll cloud build (Function: %s) (isDelete: %t): %s - %s\n", deployerConfigForFunction.Handler, deployerConfigForFunction.IsDelete, string(statusBytes), err.Error())
			pipeOutError(errorChannel, errMessage, deployerConfigForFunction.DeploymentName, deployerConfigForFunction.DirectoryName, deployerConfigForFunction.Handler)

			return
		}

		status := strings.TrimSpace(string(statusBytes))

		if status == constants.GCLOUD_BUILD_STATUS_SUCCESS {
			successMessage := fmt.Sprintf("TRACE: Status: %s (Function: %s) (isDelete: %t)\n", constants.GCLOUD_BUILD_STATUS_SUCCESS, deployerConfigForFunction.Handler, deployerConfigForFunction.IsDelete)
			utils.Logger(successMessage, verbose)

			break
		}

		if slices.Contains(constants.GCLOUD_BUILD_FAILED_STATUSES, status) {
			errMessage := fmt.Sprintf("ERR: Build failed (Function: %s) (isDelete: %t) (buildID: %s): - %s\n", deployerConfigForFunction.Handler, deployerConfigForFunction.IsDelete, buildID, status)
			pipeOutError(errorChannel, errMessage, deployerConfigForFunction.DeploymentName, deployerConfigForFunction.DirectoryName, deployerConfigForFunction.Handler)

			return
		}

		utils.Logger(fmt.Sprintf("TRACE: (Function: %s) processing (isDelete: %t) - %s\n", deployerConfigForFunction.Handler, deployerConfigForFunction.IsDelete, status), verbose)

		time.Sleep(5 * time.Second)
	}

	// Return success
	utils.Logger(fmt.Sprintf("TRACE: (Function: %s) processed (isDelete: %t)\n", deployerConfigForFunction.Handler, deployerConfigForFunction.IsDelete), true)
}

func handlePollingForDeletion(deployerConfigForFunction models.DeployerConfig, errorChannel chan models.DeploymentError, tempDir string, verbose bool) {
	// Fomart cmd for polling
	pollingCmd := []string{
		"run", "services",
		"describe",
		deployerConfigForFunction.DeploymentName,
		"--region", deployerConfigForFunction.Provider.Region,
	}

	pollingStartTime := time.Now().UTC()

	// Poll every 5 seconds for the build
	for {
		// Return if timeout is more than 15 minutes
		if time.Since(pollingStartTime) > time.Duration(constants.POLLING_TIMEOUT)*time.Second {
			// Format error
			errMessage := fmt.Sprintf("ERR: Timeout (Function: %s) (isDelete: %t)\n", deployerConfigForFunction.Handler, deployerConfigForFunction.IsDelete)
			pipeOutError(errorChannel, errMessage, deployerConfigForFunction.DeploymentName, deployerConfigForFunction.DirectoryName, deployerConfigForFunction.Handler)

			return
		}

		// Format the polling command
		pollingCmdStruct := exec.Command("gcloud", pollingCmd...)

		pollingCmdStruct.Env = append(os.Environ(),
			"CLOUDSDK_CONFIG="+tempDir,
			"GOOGLE_APPLICATION_CREDENTIALS="+deployerConfigForFunction.Provider.Credentials,
		)

		// Execute the polling
		statusBytes, err := pollingCmdStruct.CombinedOutput()
		if err != nil {
			status := strings.TrimSpace(string(statusBytes))

			if strings.Contains(status, constants.CANNOT_FIND_SERVICE) {
				successMessage := fmt.Sprintf("TRACE: Deleted (Function: %s) (isDelete: %t)\n", deployerConfigForFunction.Handler, deployerConfigForFunction.IsDelete)
				utils.Logger(successMessage, verbose)
				break
			}

			errMessage := fmt.Sprintf("ERR: Unable to poll cloud build (Function: %s) (isDelete: %t): %s - %s\n", deployerConfigForFunction.Handler, deployerConfigForFunction.IsDelete, string(statusBytes), err.Error())
			pipeOutError(errorChannel, errMessage, deployerConfigForFunction.DeploymentName, deployerConfigForFunction.DirectoryName, deployerConfigForFunction.Handler)

			return
		}

		utils.Logger(fmt.Sprintf("TRACE: (Function: %s) deleting (isDelete: %t)...\n", deployerConfigForFunction.Handler, deployerConfigForFunction.IsDelete), verbose)

		time.Sleep(5 * time.Second)
	}

	// Return success
	utils.Logger(fmt.Sprintf("TRACE: (Function: %s) deleted (isDelete: %t)\n", deployerConfigForFunction.Handler, deployerConfigForFunction.IsDelete), true)
}
