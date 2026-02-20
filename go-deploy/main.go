package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"sync"

	"github.com/hashicorp/go-set/v2"
	yaml "gopkg.in/yaml.v3"
)

// NOTE: The service account need to have `roles/serviceusage.serviceUsageConsumer` set

func main() {
	fmt.Printf("TRACE: Looping through the repo...\n")

	// Loop through all the folders in the repo and get the deployer config file.
	listOfDirs, err := os.ReadDir("./")
	if err != nil {
		fmt.Printf("%s - %s\n", UnableToReadRepoError, err.Error())
		panic(UnableToReadRepoError)
	}

	// Get provider config
	providerConfigBytes, err := os.ReadFile("provider_config.yml")
	if err != nil {
		fmt.Printf("%s - %s\n", UnableToReadProviderConfigError, err.Error())
		panic(UnableToReadProviderConfigError)
	}

	fmt.Println("TRACE: Parsing provider config...")

	// Unmarshal the provider config
	providerConfig := Provider{}
	err = yaml.Unmarshal(providerConfigBytes, &providerConfig)
	if err != nil {
		fmt.Printf("%s - %s\n", UnableToUnmarshalProviderConfigError, err.Error())
		panic(UnableToUnmarshalProviderConfigError)
	}

	fmt.Println("TRACE: Parsed provider config successfully...")

	// Open diff file with git changes
	fmt.Printf("TRACE: Reading git diff...\n")
	diffOut, err := os.ReadFile("changes.diff")
	if err != nil {
		fmt.Printf("%s - %s\n", UnableToReadGitDiffError, err.Error())
		panic(UnableToReadGitDiffError)
	}

	// Parse the git diff output and get a list of functions to deploy
	listOfFunctionsToDeploy, listOfFunctionsToDelete, listOfFoldersToDeploy := parseDiffFunctions(diffOut)

	// Get the deployer config for the repo
	deployerConfigsForTheRepo, err := getDeployerConfigsForTheRepo(listOfDirs, listOfFoldersToDeploy, listOfFunctionsToDeploy, listOfFunctionsToDelete, providerConfig)
	if err != nil {
		fmt.Printf("%s - %s\n", UnableToGetDeployerConfigsForTheRepoError, err.Error())
		panic(UnableToGetDeployerConfigsForTheRepoError)
	}

	fmt.Printf("TRACE: %d functions to process...\n", len(deployerConfigsForTheRepo))

	credentialsPath := providerConfig.Credentials
	if credentialsPath == "" {
		fmt.Println(NoCredentialsPathProvidedInProviderConfigError)
		panic(NoCredentialsPathProvidedInProviderConfigError)
	}

	// Setup gcloud
	fmt.Println("TRACE: Setting up gcloud...")
	err = setupGcloud(credentialsPath, providerConfig)
	if err != nil {
		fmt.Printf("%s - %s\n", UnableToSetupGcloudError, err.Error())
		panic(UnableToSetupGcloudError)
	}

	fmt.Println("TRACE: Formatting inputs for deployment...")

	errorChannel := make(chan DeploymentError, len(deployerConfigsForTheRepo))

	fmt.Printf("TRACE: Starting batch deployment with %d batches and %d in parallel...\n", len(deployerConfigsForTheRepo)/MAX_DEPLOYMENTS_IN_PARALLEL, MAX_DEPLOYMENTS_IN_PARALLEL)

	batchSize := MAX_DEPLOYMENTS_IN_PARALLEL
	var currentBatch []DeployerConfig
	batchCounter := 0

	for i, deployerConfigForFunction := range deployerConfigsForTheRepo {
		currentBatch = append(currentBatch, deployerConfigForFunction)
		batchCounter++

		if batchCounter == batchSize {
			// Process the batch
			processDeploymentBatch(currentBatch, errorChannel)

			fmt.Printf("TRACE: Processed %d out of %d functions...\n", i+1, len(deployerConfigsForTheRepo))

			// Reset batch
			currentBatch = nil
			batchCounter = 0
		}
	}

	// Process the last batch
	if batchCounter > 0 {
		processDeploymentBatch(currentBatch, errorChannel)
	}

	fmt.Printf("TRACE: Closing error channel...\n")
	close(errorChannel)

	if len(errorChannel) == 0 {
		fmt.Println("TRACE: Deployment successfully completed.")
		return
	}

	fmt.Println("ERR: Deployment failed with the following errors:")

	// Check for errors
	for err := range errorChannel {
		fmt.Println("---------------------------------------------------------")
		fmt.Printf("%+v\n", err)
		fmt.Println("---------------------------------------------------------")
	}

	panic(DeploymentFailedError)
}

func parseDiffFunctions(diff []byte) ([]string, []string, []string) {
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

	fmt.Printf("TRACE: Found %d function(s) to add: %+v\n", functionsToBeAdded.Size(), functionsToBeAdded)
	fmt.Printf("TRACE: Found %d function(s) to delete: %+v\n", functionsToBeDeleted.Size(), functionsToBeDeleted)
	fmt.Printf("TRACE: Folder(s) to deploy as the go.mod/go.sum files were updated: %+v\n", foldersToDeploy)

	return functionsToBeAdded.Slice(), functionsToBeDeleted.Slice(), foldersToDeploy.Slice()
}

func getDeployerConfigsForTheRepo(listOfDirs []os.DirEntry, listOfFoldersToDeploy []string, listOfFunctionsToDeploy []string, listOfFunctionsToDelete []string, providerConfig Provider) ([]DeployerConfig, error) {
	deployerConfigsForTheRepo := []DeployerConfig{}

	for _, dir := range listOfDirs {
		dirName := dir.Name()

		// Ignore hidden directories
		if dirName == "token" || strings.Contains(dirName, ".") || strings.Contains(dirName, "deploy") || strings.Contains(dirName, "Jenkinsfile") {
			fmt.Printf("TRACE: Skipping directory - %s\n", dirName)
			continue
		}

		fmt.Printf("TRACE: Found directory - %s\n", dirName)

		fmt.Printf("TRACE: Running go mod tidy...\n")
		// Run go mod tidy inside the dir
		cmdStruct := exec.Command("go", "mod", "tidy")
		cmdStruct.Dir = dirName
		out, err := cmdStruct.CombinedOutput()
		if err != nil {
			fmt.Printf("ERR: Unable to process %s - %s\n", dirName, string(out))
			return nil, err
		}

		fmt.Printf("TRACE: Running go mod vendor...\n")
		// Run go mod vendor inside the dir
		cmdStruct = exec.Command("go", "mod", "vendor")
		cmdStruct.Dir = dirName
		out, err = cmdStruct.CombinedOutput()
		if err != nil {
			fmt.Printf("ERR: Unable to process %s - %s\n", dirName, string(out))
			return nil, err
		}

		// For each dir, cd into it and get the deployer config file
		configFile, err := os.ReadFile(dirName + "/deployer_config.yml")
		if err != nil {
			fmt.Printf("ERR: Unable to read deployer config file - %s\n", err.Error())
			return nil, err
		}

		// Unmarshal the config file
		functionsConfig := map[string]Function{}
		err = yaml.Unmarshal(configFile, &functionsConfig)
		if err != nil {
			fmt.Printf("ERR: Unable to unmarshal functions config file - %s\n", err.Error())
			return nil, err
		}

		for functionDeploymentName, functionConfig := range functionsConfig {
			if slices.Contains(listOfFunctionsToDelete, functionConfig.Handler) {
				// Add the function to the deployer config to be deleted
				deployerConfigsForTheRepo = append(deployerConfigsForTheRepo, DeployerConfig{
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
				deployerConfigsForTheRepo = append(deployerConfigsForTheRepo, DeployerConfig{
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

func setupGcloud(credentialsPath string, providerConfig Provider) error {
	fmt.Printf("TRACE: Authenticating with gcloud...\n")

	// Set GOOGLE_APPLICATION_CREDENTIALS for gcloud and SDK tools
	err := os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", credentialsPath)
	if err != nil {
		fmt.Printf("ERR: Unable to set GOOGLE_APPLICATION_CREDENTIALS - %s\n", err.Error())
		return err
	}

	// Authenticate with gcloud.
	gcloudAuth := exec.Command("gcloud", "auth",
		"activate-service-account",
		"--key-file", credentialsPath,
	)

	out, err := gcloudAuth.CombinedOutput()
	if err != nil {
		fmt.Printf("ERR: Unable to authenticate with gcloud - %s\n", string(out))
		return err
	}

	// Set the project
	gcloudProject := exec.Command("gcloud", "config", "set",
		"project", providerConfig.Project,
	)

	out, err = gcloudProject.CombinedOutput()
	if err != nil {
		fmt.Printf("ERR: Unable to set project - %s\n", string(out))
		return err
	}

	// Impersonate the service account
	impersonateServiceAccount := exec.Command("gcloud", "config", "set",
		"auth/impersonate_service_account",
		providerConfig.ServiceAccountEmail,
	)

	out, err = impersonateServiceAccount.CombinedOutput()
	if err != nil {
		fmt.Printf("ERR: Unable to set impersonate service account - %s\n", string(out))
		return err
	}

	// Return success
	return nil
}

func processDeploymentBatch(deploymentBatch []DeployerConfig, errorChannel chan DeploymentError) {
	var wg sync.WaitGroup
	wg.Add(len(deploymentBatch))

	for _, deployConfig := range deploymentBatch {
		go deployFunction(deployConfig, &wg, errorChannel)
	}

	wg.Wait()
}

func deployFunction(deployerConfigForFunction DeployerConfig, wg *sync.WaitGroup, errorChannel chan DeploymentError) {
	cmdStruct := exec.Cmd{}

	if deployerConfigForFunction.IsDelete {
		fmt.Printf("TRACE: Deleting %s...\n", deployerConfigForFunction.Handler)

		// Format cmd args
		cmdArgs := []string{
			"run", "services",
			"delete",
			deployerConfigForFunction.DeploymentName,
			"--region", deployerConfigForFunction.Provider.Region,
			"--quiet",
		}

		// Execute the command
		cmdStruct = *exec.Command("gcloud", cmdArgs...)

	} else {
		fmt.Printf("TRACE: Deploying %s...\n", deployerConfigForFunction.Handler)

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

		// Execute the command
		cmdStruct = *exec.Command("gcloud", cmdArgs...)
	}

	out, err := cmdStruct.CombinedOutput()
	if err != nil {
		// Format errMessage
		errMessage := fmt.Sprintf("ERR: Unable to process (Function: %s) (isDelete: %t): %s - %s\n", deployerConfigForFunction.Handler, deployerConfigForFunction.IsDelete, string(out), err.Error())
		fmt.Println(errMessage)

		// Format error
		deploymentError := DeploymentError{
			ErrorMessage:   errMessage,
			DeploymentName: deployerConfigForFunction.DeploymentName,
			DirectoryName:  deployerConfigForFunction.DirectoryName,
			Handler:        deployerConfigForFunction.Handler,
		}

		// Pipe error to the error channel
		errorChannel <- deploymentError
		wg.Done()
		return
	}

	// Return success
	fmt.Printf("TRACE: (Function: %s) processed (isDelete: %t) - %s\n", deployerConfigForFunction.Handler, deployerConfigForFunction.IsDelete, string(out))
	wg.Done()
}
