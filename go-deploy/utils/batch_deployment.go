package utils

import (
	"fmt"
	"hutchisont/go-deployer/constants"
	"hutchisont/go-deployer/models"
	"sync"
	"time"
)

// HandleDeploymentBatchs will handle the deployment batchs
func HandleDeploymentBatchs(deployerConfigsForTheRepo map[string]models.DeployerConfig, cmd models.CMD, deploymentStartTime time.Time, isSelfHealingCycle bool) {
	errorChannel := make(chan models.DeploymentError, len(deployerConfigsForTheRepo))

	loggerString := "batch"

	if isSelfHealingCycle {
		loggerString = "self healing batch"
	}

	Logger(fmt.Sprintf("TRACE: Starting %s deployment of %d in parallel...\n", loggerString, cmd.MaxDeploymentsInParallel), true)

	batchSize := cmd.MaxDeploymentsInParallel
	var currentBatch []models.DeployerConfig
	batchCounter := 0
	deploymentCounter := 0

	for _, deployerConfigForFunction := range deployerConfigsForTheRepo {
		currentBatch = append(currentBatch, deployerConfigForFunction)
		batchCounter++
		deploymentCounter++

		if batchCounter == batchSize {
			// Process the batch
			ProcessDeploymentBatch(currentBatch, errorChannel, cmd.PollingDelay, cmd.DelayBetweenFunctionsMs, cmd.Verbose, deploymentStartTime)

			Logger(fmt.Sprintf("TRACE: Processed %d out of %d functions...\n", deploymentCounter, len(deployerConfigsForTheRepo)), true)

			// Reset batch
			currentBatch = nil
			batchCounter = 0
		}
	}

	// Process the last batch
	if batchCounter > 0 {
		ProcessDeploymentBatch(currentBatch, errorChannel, cmd.PollingDelay, cmd.DelayBetweenFunctionsMs, cmd.Verbose, deploymentStartTime)
	}

	Logger("TRACE: Closing error channel...\n", cmd.Verbose)
	close(errorChannel)

	if len(errorChannel) == 0 {
		Logger("TRACE: Deployment successfully completed.\n", true)
		return
	}

	if len(errorChannel) != 0 && !isSelfHealingCycle {
		Logger("ERR: Failed to deploy all functions, will initiate Self Healing Cycle next.\n", true)

		// Handle errors from error channel
		selfHealingDeployerConfigs := HandleErrorsFromChannel(errorChannel, cmd.Verbose, true, deployerConfigsForTheRepo)

		Logger("TRACE: Formatted Deploy Config for Self Healing Cycle...\n", true)

		// Handle self healing deployer configs
		HandleDeploymentBatchs(selfHealingDeployerConfigs, cmd, time.Now().UTC(), true)

		return
	}

	Logger("ERR: Failed to process Self Healing Cycle.\n", true)

	// Handle errors from error channel
	_ = HandleErrorsFromChannel(errorChannel, cmd.Verbose, false, nil)

	panic(constants.DeploymentFailedError)
}

// ProcessDeploymentBatch will process the deployment batch
func ProcessDeploymentBatch(deploymentBatch []models.DeployerConfig, errorChannel chan models.DeploymentError, pollingDelay int, delayBetweenFunctionsMs int, verbose bool, deploymentStartTime time.Time) {
	var wg sync.WaitGroup
	wg.Add(len(deploymentBatch))

	for _, deployConfig := range deploymentBatch {
		go DeployFunction(deployConfig, &wg, errorChannel, verbose, deploymentStartTime, pollingDelay)

		// Sleep between functions
		time.Sleep(time.Duration(delayBetweenFunctionsMs) * time.Millisecond)
	}

	wg.Wait()
}
