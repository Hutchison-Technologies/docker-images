package models

// CMD struct to handle command line arguments
type CMD struct {
	// Props
	MaxDeploymentsInParallel int  `json:"maxDeploymentsInParallel"`
	Verbose                  bool `json:"verbose"`
	PollingDelay             int  `json:"pollingDelay"`
	DelayBetweenFunctionsMs  int  `json:"delayBetweenFunctionsMs"`
}
