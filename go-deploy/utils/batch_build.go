package utils

import (
	"errors"
	"fmt"
	"hutchisont/go-deployer/constants"
	"hutchisont/go-deployer/models"
	"os"
	"slices"
	"strings"
	"sync"
	"time"
)

// HandleBuildBatches will handle the build batchs
func HandleBuildBatches(listOfDirs []os.DirEntry, listOfFoldersToDeploy []string, providerConfig models.Provider, cmd models.CMD, isSelfHealingCycle bool) error {
	errorChannel := make(chan models.DeploymentError, len(listOfDirs))

	loggerString := "batch"

	if isSelfHealingCycle {
		loggerString = "self healing batch"
	}

	Logger(fmt.Sprintf("TRACE: Starting %s builds of %d in parallel...\n", loggerString, cmd.MaxBuildsInParallel), true)

	batchSize := cmd.MaxBuildsInParallel
	var currentBatch []os.DirEntry
	batchCounter := 0
	buildCounter := 0

	for _, dir := range listOfDirs {
		currentBatch = append(currentBatch, dir)
		batchCounter++
		buildCounter++

		if batchCounter == batchSize {
			// Process the batch
			ProcessBuildBatch(currentBatch, listOfFoldersToDeploy, providerConfig, errorChannel, cmd.PollingDelay, cmd.DelayBetweenBuildsMs, cmd.Verbose)

			Logger(fmt.Sprintf("TRACE: Processed %d out of %d builds...\n", buildCounter, len(listOfDirs)), true)

			// Reset batch
			currentBatch = nil
			batchCounter = 0
		}
	}

	// Process the last batch
	if batchCounter > 0 {
		ProcessBuildBatch(currentBatch, listOfFoldersToDeploy, providerConfig, errorChannel, cmd.PollingDelay, cmd.DelayBetweenFunctionsMs, cmd.Verbose)
	}

	Logger("TRACE: Closing error channel...\n", cmd.Verbose)
	close(errorChannel)

	if len(errorChannel) == 0 {
		Logger("TRACE: Builds successfully completed.\n", true)

		// Return success
		return nil
	}

	if len(errorChannel) != 0 && !isSelfHealingCycle {
		Logger("ERR: Failed to build and push all folders, will initiate Self Healing Cycle next.\n", true)

		// Handle build errors from error channel
		selfHealingFoldersToBuild := HandleBuildErrorsFromChannel(errorChannel, cmd.Verbose, true)

		// Handle self healing deployer configs
		err := HandleBuildBatches(listOfDirs, selfHealingFoldersToBuild, providerConfig, cmd, true)
		if err != nil {
			return err
		}

		// Return success
		return nil
	}

	Logger("ERR: Failed to process Self Healing Cycle builds.\n", true)

	// Handle build errors from error channel
	_ = HandleBuildErrorsFromChannel(errorChannel, cmd.Verbose, false)

	return errors.New(constants.UnableToPackageAndPush)
}

// ProcessBuildBatch will process the deployment batch
func ProcessBuildBatch(foldersBatch []os.DirEntry, listOfFoldersToDeploy []string, providerConfig models.Provider, errorChannel chan models.DeploymentError, pollingDelay int, delayBetweenBuildsMs int, verbose bool) {
	var wg sync.WaitGroup
	wg.Add(len(foldersBatch))

	for _, dir := range foldersBatch {
		dirName := dir.Name()

		// Ignore hidden directories
		if dirName == "token" || strings.Contains(dirName, ".") || strings.Contains(dirName, "deploy") || strings.Contains(dirName, "Jenkinsfile") || strings.Contains(dirName, "deployer") {
			Logger(fmt.Sprintf("TRACE: Skipping directory - %s\n", dirName), true)
			wg.Done()

			continue
		}

		// Skip directories that are not in the list of folders to deploy
		if !slices.Contains(listOfFoldersToDeploy, dirName) {
			Logger(fmt.Sprintf("TRACE: Skipping directory as it is not in the list of folders to deploy - %s\n", dirName), true)
			wg.Done()

			continue
		}

		go func() {
			err := PackageAndPushFolder(dirName, providerConfig, verbose, pollingDelay)
			if err != nil {
				errMessage := fmt.Sprintf("ERR: Unable to package and push folder - %s\n", err.Error())
				PipeOutError(errorChannel, errMessage, "", dirName, "")

				wg.Done()
				return
			}

			wg.Done()
		}()

		// Sleep between builds
		time.Sleep(time.Duration(delayBetweenBuildsMs) * time.Millisecond)
	}

	wg.Wait()
}
