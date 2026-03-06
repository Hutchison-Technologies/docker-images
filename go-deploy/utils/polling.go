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

func HandlePollingForFolderBuild(buildID string, directoryName string, provider models.Provider, errorChannel chan models.DeploymentError, tempDir string, verbose bool, deploymentStartTime time.Time, pollingDelay int) {
	// Fomart cmd for polling
	pollingCmd := []string{
		"builds",
		"describe",
		buildID,
		"--format=value(status)",
		"--region", provider.Region,
		"--project", provider.Project,
	}

	pollingStartTime := time.Now().UTC()

	// Poll the build status
	for {
		// Return if timeout is more than 15 minutes
		if time.Since(pollingStartTime) > time.Duration(constants.POLLING_TIMEOUT)*time.Second {
			errMessage := "ERR: Polling timed out after 15 minutes\n"
			PipeOutError(errorChannel, errMessage, "", directoryName, "")

			return
		}

		// Execute the polling command
		pollingCmdStruct := *exec.Command("gcloud", pollingCmd...)

		pollingCmdStruct.Env = append(os.Environ(),
			"CLOUDSDK_CONFIG="+tempDir,
			"GOOGLE_APPLICATION_CREDENTIALS="+provider.Credentials,
		)

		statusBytes, err := pollingCmdStruct.CombinedOutput()
		if err != nil {
			errMessage := fmt.Sprintf("ERR: Unable to get build status - %s\n", err.Error())
			PipeOutError(errorChannel, errMessage, "", directoryName, "")

			return
		}

		Logger(fmt.Sprintf("TRACE: Polling Output: %s\n", string(statusBytes)), verbose)

		status := strings.TrimSpace(string(statusBytes))

		if strings.Contains(constants.GCLOUD_BUILD_STATUS_SUCCESS, status) {
			Logger(fmt.Sprintf("TRACE: Build succeeded - %s\n", buildID), verbose)
			break
		}

		if slices.Contains(constants.GCLOUD_BUILD_FAILED_STATUSES, status) {
			// Fetch failure info
			failureCmd := exec.Command(
				"gcloud", "builds", "describe", buildID,
				"--format=yaml(failureInfo,statusDetail)",
			)

			failureCmd.Env = pollingCmdStruct.Env

			failureOut, err := failureCmd.CombinedOutput()
			if err != nil {
				errMessage := fmt.Sprintf("ERR: Unable to get failure info - %s\n", err.Error())
				PipeOutError(errorChannel, errMessage, "", directoryName, "")

				return
			}

			errMessage := fmt.Sprintf("ERR: Build failed for %s: %s", directoryName, string(failureOut))
			PipeOutError(errorChannel, errMessage, "", directoryName, "")

			return
		}

		Logger(fmt.Sprintf("TRACE: Build status - %s\n", status), verbose)

		// Sleep between polling
		time.Sleep(time.Duration(pollingDelay) * time.Second)
	}
}

func HandlePollingForDeployment(deployerConfigForFunction models.DeployerConfig, errorChannel chan models.DeploymentError, tempDir string, verbose bool, deploymentStartTime time.Time) {
	// Fomart cmd for status polling
	pollingCmd := []string{
		"run", "services",
		"describe",
		deployerConfigForFunction.DeploymentName,
		"--format=value(lastTransitionTime)",
		"--region", deployerConfigForFunction.Provider.Region,
		"--project", deployerConfigForFunction.Provider.Project,
	}

	Logger(fmt.Sprintf("TRACE: Executing polling command - %s\n", strings.Join(pollingCmd, " ")), verbose)

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
		lastTransitionBytes, err := pollingCmdStruct.Output()
		if err != nil {
			errMessage := fmt.Sprintf("ERR: Unable to poll cloud run service (Function: %s) (isDelete: %t): %s - %s\n", deployerConfigForFunction.Handler, deployerConfigForFunction.IsDelete, string(lastTransitionBytes), err.Error())
			PipeOutError(errorChannel, errMessage, deployerConfigForFunction.DeploymentName, deployerConfigForFunction.DirectoryName, deployerConfigForFunction.Handler)

			return
		}

		lastTransitionString := strings.TrimSpace(string(lastTransitionBytes))

		lastTransitionTime, err := time.Parse(time.RFC3339, lastTransitionString)
		if err != nil {
			fmt.Printf("ERR: Unable to parse time - %s\n", err.Error())
			return
		}

		if lastTransitionTime.After(deploymentStartTime) {
			break
		}

		Logger(fmt.Sprintf("TRACE: (Function: %s) processing (isDelete: %t)\n", deployerConfigForFunction.Handler, deployerConfigForFunction.IsDelete), verbose)

		time.Sleep(5 * time.Second)
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
