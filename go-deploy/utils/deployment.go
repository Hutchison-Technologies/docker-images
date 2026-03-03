package utils

import (
	"errors"
	"fmt"
	"hutchisont/go-deployer/models"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

func PackageAndPushFolder(folder string, provider models.Provider, verbose bool, pollingDelay int) error {
	// Create isolated gcloud config directory
	tempDir, err := os.MkdirTemp("", "gcloud-*")
	if err != nil {
		errMessage := fmt.Sprintf("ERR: Unable to create temp gcloud dir: %s", err.Error())
		Logger(errMessage, true)
		return errors.New(errMessage)
	}

	defer func() {
		Logger("TRACE: Removing temp gcloud dir...\n", verbose)

		err := os.RemoveAll(tempDir)
		if err != nil {
			errMessage := fmt.Sprintf("ERR: Unable to remove temp gcloud dir: %s", err.Error())
			Logger(errMessage, true)
			return
		}
	}()

	Logger(fmt.Sprintf("TRACE: Packaging folder %s...\n", folder), verbose)
	// europe-west2-docker.pkg.dev/vpc-test-worker-1-107d0240/cloud-run-source-deploy/

	imageTag := fmt.Sprintf("%s-docker.pkg.dev/%s/%s/%s:latest", provider.Region, provider.Project, provider.ArtifactRegistryRepo, strings.ToLower(folder))

	cmdArgs := []string{
		"builds", "submit",
		"--pack", fmt.Sprintf("image=%s", imageTag),
		"--project", provider.Project,
		"--region", provider.Region,
		folder,
		"--async",
		"--format=value(ID)",
		"--verbosity", "error",
	}

	Logger(fmt.Sprintf("TRACE: Executing command for directory image - %+v\n", cmdArgs), verbose)

	cmd := exec.Command("gcloud", cmdArgs...)
	cmd.Env = append(os.Environ(),
		"CLOUDSDK_CONFIG="+tempDir,
		"GOOGLE_APPLICATION_CREDENTIALS="+provider.Credentials,
	)

	out, err := cmd.Output()
	if err != nil {
		errMessage := fmt.Sprintf("ERR: Unable to build image - %s - %s\n", string(out), err.Error())
		Logger(errMessage, true)
		return errors.New(errMessage)
	}

	Logger(fmt.Sprintf("TRACE: Output - %s\n", string(out)), verbose)

	buildID := strings.TrimSpace(string(out))

	// Log build ID
	Logger(fmt.Sprintf("TRACE: Build ID - %s\n", buildID), verbose)

	// Poll for the build ID
	HandlePollingForFolderBuild(buildID, folder, provider, nil, tempDir, verbose, time.Now().UTC(), pollingDelay)

	fmt.Printf("TRACE: Built image %s\n", imageTag)

	return nil
}

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

		// Add function target
		envVars = append(envVars, "FUNCTION_TARGET="+deployerConfigForFunction.Handler)

		envVarsArg := strings.Join(envVars, ",")

		// Format label
		label := fmt.Sprintf("service=%s", strings.ToLower(deployerConfigForFunction.DirectoryName))

		// Format image tag
		imageTag := fmt.Sprintf("%s-docker.pkg.dev/%s/%s/%s:latest",
			deployerConfigForFunction.Provider.Region, deployerConfigForFunction.Provider.Project,
			deployerConfigForFunction.Provider.ArtifactRegistryRepo, strings.ToLower(deployerConfigForFunction.DirectoryName))

		// Format cmd
		cmdArgs := []string{
			"run", "deploy", deployerConfigForFunction.DeploymentName,
			"--image", imageTag,
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

	// Run the gcloud run deploy command
	deployOutBytes, err := cmdStruct.Output()
	if err != nil {
		// Format errMessage
		errMessage := fmt.Sprintf("ERR: Unable to run deploy command (Function: %s) (isDelete: %t): %s - %s\n", deployerConfigForFunction.Handler, deployerConfigForFunction.IsDelete, string(deployOutBytes), err.Error())
		PipeOutError(errorChannel, errMessage, deployerConfigForFunction.DeploymentName, deployerConfigForFunction.DirectoryName, deployerConfigForFunction.Handler)

		return
	}

	// Log deploy output
	Logger(fmt.Sprintf("TRACE: Triggered Deployment (Function: %s) - %s\n", deployerConfigForFunction.Handler, string(deployOutBytes)), verbose)

	// Handle polling
	if deployerConfigForFunction.IsDelete {
		HandlePollingForDeletion(deployerConfigForFunction, errorChannel, tempDir, verbose, pollingDelay)
	} else {
		HandlePollingForDeployment(deployerConfigForFunction, errorChannel, tempDir, verbose, deploymentStartTime)
	}
}
