package main

var (
	UnableToReadRepoError                          string = "ERR: Unable to read repo"
	UnableToReadProviderConfigError                string = "ERR: Unable to read provider config file"
	UnableToUnmarshalProviderConfigError           string = "ERR: Unable to unmarshal provider config file"
	UnableToReadGitDiffError                       string = "ERR: Unable to read git diff"
	UnableToGetDeployerConfigsForTheRepoError      string = "ERR: Unable to get deployer configs for the repo"
	NoCredentialsPathProvidedInProviderConfigError string = "ERR: No credentials path provided in provider config"
	UnableToSetupGcloudError                       string = "ERR: Unable to setup gcloud"
	DeploymentFailedError                          string = "ERR: Deployment failed."
)
