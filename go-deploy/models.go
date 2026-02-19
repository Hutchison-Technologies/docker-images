package main

type Provider struct {
	Runtime             string            `yaml:"runtime"`
	Project             string            `yaml:"project"`
	Region              string            `yaml:"region"`
	ServiceAccountEmail string            `yaml:"serviceAccountEmail"`
	Credentials         string            `yaml:"credentials"`
	Environment         map[string]string `yaml:"environment"`
}

type Function struct {
	Handler    string `yaml:"handler"`
	MemorySize string `yaml:"memorySize"`
	Timeout    string `yaml:"timeout"`
}

type DeployerConfig struct {
	IsDelete       bool
	DeploymentName string
	Handler        string
	MemorySize     string
	Timeout        string
	DirectoryName  string
	Provider       Provider
}

type DeploymentError struct {
	ErrorMessage   string
	DeploymentName string
	DirectoryName  string
	Handler        string
}
