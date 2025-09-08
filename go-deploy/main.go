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
	"time"

	"github.com/hashicorp/go-set/v2"
	yaml "gopkg.in/yaml.v3"
)

// NOTE: The service account need to have `roles/serviceusage.serviceUsageConsumer` set

func main() {
	fmt.Printf("TRACE: Looping through the repo...\n")

	// Loop through all the folders in the repo and get the deployer config file
	listOfDirs, err := os.ReadDir("./")
	if err != nil {
		fmt.Printf("ERR: Unable to read repo - %s\n", err.Error())
		return
	}

	// Get provider config
	providerConfigBytes, err := os.ReadFile("provider_config.yml")
	if err != nil {
		fmt.Printf("ERR: Unable to read provider config file - %s\n", err.Error())
		return
	}

	fmt.Println("TRACE: Parsing provider config...")

	// Unmarshal the provider config
	providerConfig := Provider{}
	err = yaml.Unmarshal(providerConfigBytes, &providerConfig)
	if err != nil {
		fmt.Printf("ERR: Unable to unmarshal provider config file - %s\n", err.Error())
		return
	}

	fmt.Printf("TRACE: Found provider config - %+v\n", providerConfig)

	// Open diff file with git changes
	fmt.Printf("TRACE: Reading git diff...\n")
	diffOut, err := os.ReadFile("changes.diff")
	if err != nil {
		fmt.Printf("ERR: Unable to read git diff - %s\n", err.Error())
		return
	}

	// Parse the git diff output and get a list of functions to deploy
	listOfFunctionsToDeploy, listOfFunctionsToDelete, listOfFoldersToDeploy := parseDiffFunctions(diffOut)

	// Get the deployer config for the repo
	deployerConfigsForTheRepo, err := getDeployerConfigsForTheRepo(listOfDirs, listOfFoldersToDeploy, listOfFunctionsToDeploy, listOfFunctionsToDelete, providerConfig)
	if err != nil {
		fmt.Printf("ERR: Unable to get deployer configs for the repo - %s\n", err.Error())
		return
	}

	fmt.Printf("TRACE: %d functions to process...\n", len(deployerConfigsForTheRepo))
	fmt.Printf("TRACE: Generated deployer config for the repo - %+v\n", deployerConfigsForTheRepo)

	credentialsPath := providerConfig.Credentials
	if credentialsPath == "" {
		fmt.Println("ERR: No credentials path provided in provider config")
		return
	}

	// Setup gcloud
	fmt.Println("TRACE: Setting up gcloud...")
	err = setupGcloud(credentialsPath, providerConfig)
	if err != nil {
		fmt.Printf("ERR: Unable to setup gcloud - %s\n", err.Error())
		return
	}

	fmt.Println("TRACE: Formatting inputs for deployment...")

	wg := sync.WaitGroup{}
	wg.Add(len(deployerConfigsForTheRepo))

	inputChannel := make(chan DeployerConfig, len(deployerConfigsForTheRepo))
	errorChannel := make(chan DeploymentError, len(deployerConfigsForTheRepo))

	// Populate the input channel
	for _, deployerConfigForFunction := range deployerConfigsForTheRepo {
		inputChannel <- deployerConfigForFunction
	}

	// Close the input channel
	close(inputChannel)

	fmt.Printf("TRACE: Setting up ticker...\n")

	// Create a ticker that ticks every (60s / max deployments in parallel) seconds
	// This is used to limit the number of deployments that can be done within a minute
	deploymentTicker := time.NewTicker(time.Minute / MAX_DEPLOYMENTS_IN_PARALLEL)

	for inputForDeployment := range inputChannel {
		fmt.Printf("TRACE: Waiting for ticker to tick...\n")

		// Wait for the ticker to tick
		<-deploymentTicker.C

		// Kick off a goroutine for each function
		go deployFunction(inputForDeployment, &wg, errorChannel)
	}

	wg.Wait()

	fmt.Println("TRACE: Stopping ticker and closing error channel...")
	deploymentTicker.Stop()
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
		// Ignore hidden directories
		if strings.Contains(dir.Name(), ".") || strings.Contains(dir.Name(), "deploy") || strings.Contains(dir.Name(), "Jenkinsfile") {
			continue
		}

		fmt.Printf("TRACE: Found directory - %s\n", dir.Name())

		fmt.Printf("TRACE: Running go mod tidy...\n")
		// Run go mod tidy inside the dir
		cmdStruct := exec.Command("go", "mod", "tidy")
		cmdStruct.Dir = dir.Name()
		out, err := cmdStruct.CombinedOutput()
		if err != nil {
			fmt.Printf("ERR: Unable to process %s - %s\n", dir.Name(), string(out))
			return nil, err
		}

		fmt.Printf("TRACE: Running go mod vendor...\n")
		// Run go mod vendor inside the dir
		cmdStruct = exec.Command("go", "mod", "vendor")
		cmdStruct.Dir = dir.Name()
		out, err = cmdStruct.CombinedOutput()
		if err != nil {
			fmt.Printf("ERR: Unable to process %s - %s\n", dir.Name(), string(out))
			return nil, err
		}

		// For each dir, cd into it and check if it has a deployer config file
		// If it does, then parse it and get the functions to deploy
		// If it doesn't, then skip	it
		configFile, err := os.ReadFile(dir.Name() + "/deployer_config.yml")
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
					IsDelete:       true,
					DeploymentName: functionDeploymentName,
					DirectoryName:  dir.Name(),
					Provider:       providerConfig,
					Handler:        functionConfig.Handler,
					MemorySize:     functionConfig.MemorySize,
				})
			}

			if slices.Contains(listOfFunctionsToDeploy, functionConfig.Handler) ||
				(listOfFoldersToDeploy != nil && slices.Contains(listOfFoldersToDeploy, dir.Name())) {
				// Skip function if it is to be deleted
				if slices.Contains(listOfFunctionsToDelete, functionConfig.Handler) {
					continue
				}

				// Add the function to the deployer config
				deployerConfigsForTheRepo = append(deployerConfigsForTheRepo, DeployerConfig{
					IsDelete:       false,
					DeploymentName: functionDeploymentName,
					DirectoryName:  dir.Name(),
					Provider:       providerConfig,
					Handler:        functionConfig.Handler,
					MemorySize:     functionConfig.MemorySize,
				})
			}
		}
	}

	// Return the deployer configs
	return deployerConfigsForTheRepo, nil
}

func setupGcloud(credentialsPath string, providerConfig Provider) error {
	fmt.Printf("TRACE: Authenticating with gcloud...\n")

	// Authenticate with gcloud
	gcloudAuth := exec.Command("gcloud", "auth",
		"activate-service-account",
		"--key-file", credentialsPath,
	)

	out, err := gcloudAuth.CombinedOutput()
	if err != nil {
		fmt.Printf("ERR: Unable to authenticate with gcloud - %s\n", string(out))
		return err
	}

	fmt.Printf("TRACE: Setting up project...\n")

	// Set the project
	gcloudProject := exec.Command("gcloud", "config", "set",
		"project", providerConfig.Project,
	)

	out, err = gcloudProject.CombinedOutput()
	if err != nil {
		fmt.Printf("ERR: Unable to set project - %s\n", string(out))
		return err
	}

	// Return success
	return nil
}

func deployFunction(deployerConfigForFunction DeployerConfig, wg *sync.WaitGroup, errorChannel chan DeploymentError) {
	cmdStruct := exec.Cmd{}

	if deployerConfigForFunction.IsDelete {
		fmt.Printf("TRACE: Deleting %s...\n", deployerConfigForFunction.Handler)

		// Format cmd
		cmdStruct = *exec.Command("gcloud", "run", "services",
			"delete",
			deployerConfigForFunction.DeploymentName,
			"--region", deployerConfigForFunction.Provider.Region,
			"--quiet",
		)
	} else {
		fmt.Printf("TRACE: Deploying %s...\n", deployerConfigForFunction.Handler)

		// Format cmd
		cmdStruct = *exec.Command("gcloud", "run", "deploy", deployerConfigForFunction.DeploymentName,
			"--source", deployerConfigForFunction.DirectoryName,
			"--function", deployerConfigForFunction.Handler,
			"--base-image", deployerConfigForFunction.Provider.Runtime,
			"--memory", deployerConfigForFunction.MemorySize+"Mi",
			"--region", deployerConfigForFunction.Provider.Region,
			"--allow-unauthenticated",
			"--ingress", "internal",
			"--set-env-vars", fmt.Sprintf("GOOGLE_CLOUD_PROJECT=%s", deployerConfigForFunction.Provider.Environment["GOOGLE_CLOUD_PROJECT_ID"]),
			"--set-env-vars", fmt.Sprintf("PASSWORD_PEPPER=%s", deployerConfigForFunction.Provider.Environment["PASSWORD_PEPPER"]),
			"--set-env-vars", fmt.Sprintf("PERSONAL_DETAILS_ENC_KEY=%s", deployerConfigForFunction.Provider.Environment["PERSONAL_DETAILS_ENC_KEY"]),
			"--set-env-vars", fmt.Sprintf("PASS_ENCRYPTION_SECRET=%s", deployerConfigForFunction.Provider.Environment["PASS_ENCRYPTION_SECRET"]),
			"--set-env-vars", fmt.Sprintf("PASS_ENCRYPTION_IV=%s", deployerConfigForFunction.Provider.Environment["PASS_ENCRYPTION_IV"]),
		)
	}

	out, err := cmdStruct.CombinedOutput()
	if err != nil {
		// Format errMessage
		errMessage := fmt.Sprintf("ERR: Unable to process (Function: %s) (isDelete: %t) - %s\n", deployerConfigForFunction.Handler, deployerConfigForFunction.IsDelete, string(out))
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
