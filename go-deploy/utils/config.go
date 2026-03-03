package utils

import (
	"bufio"
	"bytes"
	"fmt"
	"hutchisont/go-deployer/models"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/hashicorp/go-set/v2"
	"gopkg.in/yaml.v3"
)

func ParseDiffFunctions(diff []byte, verbose bool) ([]string, []string) {
	functionsToBeAdded := set.From([]string{})
	functionsToBeDeleted := set.From([]string{})
	foldersToDeploy := set.From([]string{})

	scanner := bufio.NewScanner(bytes.NewReader(diff))

	funcRegex := regexp.MustCompile(`\bfunc\s*(?:\([^\)]*\)\s*)?(\w+)\s*\(`)
	currentFile := ""
	currentDir := ""

	// Loop through the diff file lines
	for scanner.Scan() {
		line := scanner.Text()

		// Get the folder if go.mod or go.sum was modified
		if strings.HasPrefix(line, "diff --git") {
			parts := strings.Split(line, " ")
			if len(parts) >= 3 {
				currentFile = strings.TrimPrefix(parts[2], "a/")

				// Get the current folder
				currentDir = filepath.Dir(currentFile)
			}

			if strings.HasSuffix(currentFile, "go.mod") || strings.HasSuffix(currentFile, "go.sum") {
				dir := filepath.Dir(currentFile)

				// Deploy the current folder
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

				// Deploy the current folder
				foldersToDeploy.Insert(currentDir)
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
					functionsToBeAdded.Insert(funcName)

					// Deploy the current folder
					foldersToDeploy.Insert(currentDir)
				}
			}
		}

		// Deleted functions
		if strings.HasPrefix(line, "-func ") || strings.HasPrefix(line, "- func") {
			matches := funcRegex.FindStringSubmatch(line[1:])
			if len(matches) > 1 && matches[1] != "" {
				funcName := matches[1]
				functionsToBeDeleted.Insert(funcName)

				// Deploy the current folder as the file was updated
				foldersToDeploy.Insert(currentDir)
			}
		}
	}

	// Remove modified funcs from the delete list (appearing in both add and delete)
	for _, addedFunc := range functionsToBeAdded.Slice() {
		if functionsToBeDeleted.Contains(addedFunc) {
			functionsToBeDeleted.Remove(addedFunc)
		}
	}

	fmt.Printf("TRACE: Found %d function(s) to delete: %+v\n", functionsToBeDeleted.Size(), functionsToBeDeleted)
	fmt.Printf("TRACE: Folder(s) to deploy: %+v\n", foldersToDeploy)

	return functionsToBeDeleted.Slice(), foldersToDeploy.Slice()
}

func GetDeployerConfigsForTheRepo(listOfDirs []os.DirEntry, listOfFoldersToDeploy []string, listOfFunctionsToDelete []string, providerConfig models.Provider, cmd models.CMD) (map[string]models.DeployerConfig, error) {
	deployerConfigsForTheRepo := map[string]models.DeployerConfig{}

	for _, dir := range listOfDirs {
		dirName := dir.Name()

		// Ignore hidden directories
		if dirName == "token" || strings.Contains(dirName, ".") || strings.Contains(dirName, "deploy") || strings.Contains(dirName, "Jenkinsfile") {
			Logger(fmt.Sprintf("TRACE: Skipping directory - %s\n", dirName), true)
			continue
		}

		Logger(fmt.Sprintf("TRACE: Found directory - %s\n", dirName), cmd.Verbose)

		Logger("TRACE: Running go mod tidy...\n", cmd.Verbose)
		// Run go mod tidy inside the dir
		cmdStruct := exec.Command("go", "mod", "tidy")
		cmdStruct.Dir = dirName
		out, err := cmdStruct.Output()
		if err != nil {
			Logger(fmt.Sprintf("ERR: Unable to process %s - %s\n", dirName, string(out)), true)
			return nil, err
		}

		Logger("TRACE: Running go mod vendor...\n", cmd.Verbose)
		// Run go mod vendor inside the dir
		cmdStruct = exec.Command("go", "mod", "vendor")
		cmdStruct.Dir = dirName
		out, err = cmdStruct.Output()
		if err != nil {
			Logger(fmt.Sprintf("ERR: Unable to process %s - %s\n", dirName, string(out)), true)
			return nil, err
		}

		// For each dir, cd into it and get the deployer config file
		configFile, err := os.ReadFile(dirName + "/deployer_config.yml")
		if err != nil {
			Logger(fmt.Sprintf("ERR: Unable to read deployer config file - %s\n", err.Error()), true)
			return nil, err
		}

		// Unmarshal the config file
		functionsConfig := map[string]models.Function{}
		err = yaml.Unmarshal(configFile, &functionsConfig)
		if err != nil {
			Logger(fmt.Sprintf("ERR: Unable to unmarshal functions config file - %s\n", err.Error()), true)
			return nil, err
		}

		for functionDeploymentName, functionConfig := range functionsConfig {
			// Format deployer config
			deployerConfigForFunction := models.DeployerConfig{
				IsDelete:               false,
				DeploymentName:         functionDeploymentName,
				DirectoryName:          dirName,
				Provider:               providerConfig,
				Handler:                functionConfig.Handler,
				MemorySize:             functionConfig.MemorySize,
				Timeout:                functionConfig.Timeout,
				EnvironmentForFunction: functionConfig.EnvironmentForFunction,
			}

			if slices.Contains(listOfFunctionsToDelete, functionConfig.Handler) {
				// Mark the function as to be deleted
				deployerConfigForFunction.IsDelete = true

				// Add the deployer config to the map
				deployerConfigsForTheRepo[deployerConfigForFunction.DeploymentName] = deployerConfigForFunction

				// Continue to the next function
				continue
			}

			if listOfFoldersToDeploy != nil && slices.Contains(listOfFoldersToDeploy, dirName) {
				// Add the deployer config to the map
				deployerConfigsForTheRepo[deployerConfigForFunction.DeploymentName] = deployerConfigForFunction
			}
		}
	}

	err := HandleBuildBatches(listOfDirs, listOfFoldersToDeploy, providerConfig, cmd, false)
	if err != nil {
		errMessage := fmt.Sprintf("ERR: Unable to build and push folders - %s\n", err.Error())
		Logger(errMessage, true)
		return nil, err
	}

	// Return the deployer configs map
	Logger("TRACE: Package and push complete.\n", true)
	return deployerConfigsForTheRepo, nil
}
