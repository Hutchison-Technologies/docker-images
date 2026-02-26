package utils

import (
	"fmt"
	"hutchisont/go-deployer/models"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

func DeployFunction(deployerConfigForFunction models.DeployerConfig, wg *sync.WaitGroup, errorChannel chan models.DeploymentError, verbose bool, deploymentStartTime time.Time, pollingDelay int) {
	defer wg.Done()

	// Create isolated gcloud config directory
	tempDir, err := os.MkdirTemp("", "gcloud-*")
	if err != nil {
		errMessage := fmt.Sprintf("ERR: Unable to create temp gcloud dir: %s", err.Error())
		PipeOutError(errorChannel, errMessage, deployerConfigForFunction.DeploymentName, deployerConfigForFunction.DirectoryName, deployerConfigForFunction.Handler)

		return
	}

	defer func(errorChannel chan models.DeploymentError) {
		Logger("TRACE: Removing temp gcloud dir...\n", verbose)

		err := os.RemoveAll(tempDir)
		if err != nil {
			errMessage := fmt.Sprintf("ERR: Unable to remove temp gcloud dir: %s", err.Error())
			PipeOutError(errorChannel, errMessage, deployerConfigForFunction.DeploymentName, deployerConfigForFunction.DirectoryName, deployerConfigForFunction.Handler)

			return
		}
	}(errorChannel)

	cmdStruct := exec.Cmd{}

	if deployerConfigForFunction.IsDelete {
		Logger(fmt.Sprintf("TRACE: Deleting %s...\n", deployerConfigForFunction.Handler), true)

		// Format cmd args
		cmdArgs := []string{
			"run", "services",
			"delete",
			deployerConfigForFunction.DeploymentName,
			"--region", deployerConfigForFunction.Provider.Region,
			"--project", deployerConfigForFunction.Provider.Project,
			"--quiet",
			"--async",
			"--service-account", deployerConfigForFunction.Provider.ServiceAccountEmail,
			"--impersonate-service-account", deployerConfigForFunction.Provider.ServiceAccountEmail,
		}

		// Log CMD args
		Logger(fmt.Sprintf("TRACE: Executing command - %s\n", strings.Join(cmdArgs, " ")), verbose)

		// Format the delete command
		cmdStruct = *exec.Command("gcloud", cmdArgs...)

	} else {
		Logger(fmt.Sprintf("TRACE: Deploying %s...\n", deployerConfigForFunction.Handler), true)

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
			"--async",
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
		Logger(fmt.Sprintf("TRACE: Executing command - %s\n", strings.Join(cmdArgs, " ")), verbose)

		// Format the deploy command
		cmdStruct = *exec.Command("gcloud", cmdArgs...)
	}

	cmdStruct.Env = append(os.Environ(),
		"CLOUDSDK_CONFIG="+tempDir,
		"GOOGLE_APPLICATION_CREDENTIALS="+deployerConfigForFunction.Provider.Credentials,
	)

	// Trigger the gcloud run deploy command
	err = cmdStruct.Start()
	if err != nil {
		// Format errMessage
		errMessage := fmt.Sprintf("ERR: Unable to run deploy command (Function: %s) (isDelete: %t) - %s\n", deployerConfigForFunction.Handler, deployerConfigForFunction.IsDelete, err.Error())
		PipeOutError(errorChannel, errMessage, deployerConfigForFunction.DeploymentName, deployerConfigForFunction.DirectoryName, deployerConfigForFunction.Handler)

		return
	}

	defer func() {
		err = cmdStruct.Wait()
		if err != nil {
			// Format errMessage
			errMessage := fmt.Sprintf("ERR: Unable to release deployment (Function: %s) (isDelete: %t) - %s\n", deployerConfigForFunction.Handler, deployerConfigForFunction.IsDelete, err.Error())
			PipeOutError(errorChannel, errMessage, deployerConfigForFunction.DeploymentName, deployerConfigForFunction.DirectoryName, deployerConfigForFunction.Handler)
		}
	}()

	// Log deploy output
	Logger(fmt.Sprintf("TRACE: Triggered Deployment (Function: %s)\n", deployerConfigForFunction.Handler), verbose)

	// Handle polling
	if deployerConfigForFunction.IsDelete {
		HandlePollingForDeletion(deployerConfigForFunction, errorChannel, tempDir, verbose, pollingDelay)
	} else {
		HandlePollingForDeployment(deployerConfigForFunction, errorChannel, tempDir, verbose, deploymentStartTime, pollingDelay)
	}
}
