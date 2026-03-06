package models

// CMD struct to handle command line arguments
type CMD struct {
	// Props
	MaxBuildsInParallel              int  `json:"maxBuildsInParallel"`
	MaxFunctionDeploymentsInParallel int  `json:"maxFunctionDeploymentsInParallel"`
	PollingDelay                     int  `json:"pollingDelay"`
	DelayBetweenBuildsMs             int  `json:"delayBetweenBuildsMs"`
	DelayBetweenFunctionsMs          int  `json:"delayBetweenFunctionsMs"`
	RunGoBuild                       bool `json:"runGoBuild"`
	Verbose                          bool `json:"verbose"`
}
