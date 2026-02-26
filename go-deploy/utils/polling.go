package utils

import (
	"fmt"
	"hutchisont/go-deployer/constants"
	"hutchisont/go-deployer/models"
	"os"
	"os/exec"
	"slices"
	"strings"
	"time"
)

func HandlePollingForDeployment(deployerConfigForFunction models.DeployerConfig, errorChannel chan models.DeploymentError, tempDir string, verbose bool, deploymentStartTime time.Time, pollingDelay int) {
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
	Logger(fmt.Sprintf("TRACE: Executing get builds list command - %s\n", strings.Join(getBuildArgs, " ")), verbose)

	buildID := ""

	cloudBuildPollingStartTime := time.Now().UTC()

	// Manually Poll for the build ID
	for {
		// Return if timeout is more than 15 minutes
		if time.Since(cloudBuildPollingStartTime) > time.Duration(constants.POLLING_TIMEOUT)*time.Second {
			// Format error
			errMessage := fmt.Sprintf("ERR: Timeout (Function: %s) (isDelete: %t)\n", deployerConfigForFunction.Handler, deployerConfigForFunction.IsDelete)
			PipeOutError(errorChannel, errMessage, deployerConfigForFunction.DeploymentName, deployerConfigForFunction.DirectoryName, deployerConfigForFunction.Handler)

			return
		}

		// Format the build polling command
		buildCmdStruct := exec.Command("gcloud", getBuildArgs...)

		buildCmdStruct.Env = append(os.Environ(),
			"CLOUDSDK_CONFIG="+tempDir,
			"GOOGLE_APPLICATION_CREDENTIALS="+deployerConfigForFunction.Provider.Credentials,
		)

		// Execute the build polling
		buildOut, err := buildCmdStruct.Output()
		if err != nil {
			errMessage := fmt.Sprintf("ERR: Failed to fetch cloud build ID: %s - %s\n", string(buildOut), err)
			PipeOutError(errorChannel, errMessage, deployerConfigForFunction.DeploymentName, deployerConfigForFunction.DirectoryName, deployerConfigForFunction.Handler)

			return
		}

		// Parse build ID
		buildID = strings.TrimSpace(string(buildOut))

		if buildID != "" {
			// Log build ID
			Logger(fmt.Sprintf("TRACE: Initiated (buildID: %s) (Function: %s) (isDelete: %t)\n", buildID, deployerConfigForFunction.Handler, deployerConfigForFunction.IsDelete), verbose)
			break
		}

		Logger(fmt.Sprintf("TRACE: Waiting to get buildID (Function: %s) (isDelete: %t)...\n", deployerConfigForFunction.Handler, deployerConfigForFunction.IsDelete), verbose)
		time.Sleep(time.Duration(pollingDelay) * time.Second)
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

	// Manually Poll for the build status
	for {
		// Return if timeout is more than 15 minutes
		if time.Since(pollingStartTime) > time.Duration(constants.POLLING_TIMEOUT)*time.Second {
			// Format error
			errMessage := fmt.Sprintf("ERR: Timeout (Function: %s) (isDelete: %t)\n", deployerConfigForFunction.Handler, deployerConfigForFunction.IsDelete)
			PipeOutError(errorChannel, errMessage, deployerConfigForFunction.DeploymentName, deployerConfigForFunction.DirectoryName, deployerConfigForFunction.Handler)

			return
		}

		// Format the polling command
		pollingCmdStruct := exec.Command("gcloud", pollingCmd...)

		pollingCmdStruct.Env = append(os.Environ(),
			"CLOUDSDK_CONFIG="+tempDir,
			"GOOGLE_APPLICATION_CREDENTIALS="+deployerConfigForFunction.Provider.Credentials,
		)

		// Execute the polling
		statusBytes, err := pollingCmdStruct.Output()
		if err != nil {
			errMessage := fmt.Sprintf("ERR: Unable to poll cloud build (Function: %s) (isDelete: %t): %s - %s\n", deployerConfigForFunction.Handler, deployerConfigForFunction.IsDelete, string(statusBytes), err.Error())
			PipeOutError(errorChannel, errMessage, deployerConfigForFunction.DeploymentName, deployerConfigForFunction.DirectoryName, deployerConfigForFunction.Handler)

			return
		}

		status := strings.TrimSpace(string(statusBytes))

		if status == constants.GCLOUD_BUILD_STATUS_SUCCESS {
			successMessage := fmt.Sprintf("TRACE: Status: %s (Function: %s) (isDelete: %t)\n", constants.GCLOUD_BUILD_STATUS_SUCCESS, deployerConfigForFunction.Handler, deployerConfigForFunction.IsDelete)
			Logger(successMessage, verbose)

			break
		}

		if slices.Contains(constants.GCLOUD_BUILD_FAILED_STATUSES, status) {
			errMessage := fmt.Sprintf("ERR: Build failed (Function: %s) (isDelete: %t) (buildID: %s): - %s\n", deployerConfigForFunction.Handler, deployerConfigForFunction.IsDelete, buildID, status)
			PipeOutError(errorChannel, errMessage, deployerConfigForFunction.DeploymentName, deployerConfigForFunction.DirectoryName, deployerConfigForFunction.Handler)

			return
		}

		Logger(fmt.Sprintf("TRACE: (Function: %s) processing (isDelete: %t) - %s\n", deployerConfigForFunction.Handler, deployerConfigForFunction.IsDelete, status), verbose)

		time.Sleep(time.Duration(pollingDelay) * time.Second)
	}

	// Return success
	Logger(fmt.Sprintf("TRACE: (Function: %s) processed (isDelete: %t)\n", deployerConfigForFunction.Handler, deployerConfigForFunction.IsDelete), true)
}

func HandlePollingForDeletion(deployerConfigForFunction models.DeployerConfig, errorChannel chan models.DeploymentError, tempDir string, verbose bool, pollingDelay int) {
	// Fomart cmd for polling
	pollingCmd := []string{
		"run", "services",
		"describe",
		deployerConfigForFunction.DeploymentName,
		"--region", deployerConfigForFunction.Provider.Region,
	}

	pollingStartTime := time.Now().UTC()

	// Manually Poll for the build
	for {
		// Return if timeout is more than 15 minutes
		if time.Since(pollingStartTime) > time.Duration(constants.POLLING_TIMEOUT)*time.Second {
			// Format error
			errMessage := fmt.Sprintf("ERR: Timeout (Function: %s) (isDelete: %t)\n", deployerConfigForFunction.Handler, deployerConfigForFunction.IsDelete)
			PipeOutError(errorChannel, errMessage, deployerConfigForFunction.DeploymentName, deployerConfigForFunction.DirectoryName, deployerConfigForFunction.Handler)

			return
		}

		// Format the polling command
		pollingCmdStruct := exec.Command("gcloud", pollingCmd...)

		pollingCmdStruct.Env = append(os.Environ(),
			"CLOUDSDK_CONFIG="+tempDir,
			"GOOGLE_APPLICATION_CREDENTIALS="+deployerConfigForFunction.Provider.Credentials,
		)

		// Execute the polling
		statusBytes, err := pollingCmdStruct.Output()
		if err != nil {
			status := strings.TrimSpace(string(statusBytes))

			if strings.Contains(status, constants.CANNOT_FIND_SERVICE) {
				successMessage := fmt.Sprintf("TRACE: Deleted (Function: %s) (isDelete: %t)\n", deployerConfigForFunction.Handler, deployerConfigForFunction.IsDelete)
				Logger(successMessage, verbose)
				break
			}

			errMessage := fmt.Sprintf("ERR: Unable to poll cloud build (Function: %s) (isDelete: %t): %s - %s\n", deployerConfigForFunction.Handler, deployerConfigForFunction.IsDelete, string(statusBytes), err.Error())
			PipeOutError(errorChannel, errMessage, deployerConfigForFunction.DeploymentName, deployerConfigForFunction.DirectoryName, deployerConfigForFunction.Handler)

			return
		}

		Logger(fmt.Sprintf("TRACE: (Function: %s) deleting (isDelete: %t)...\n", deployerConfigForFunction.Handler, deployerConfigForFunction.IsDelete), verbose)

		time.Sleep(time.Duration(pollingDelay) * time.Second)
	}

	// Return success
	Logger(fmt.Sprintf("TRACE: (Function: %s) deleted (isDelete: %t)\n", deployerConfigForFunction.Handler, deployerConfigForFunction.IsDelete), true)
}
