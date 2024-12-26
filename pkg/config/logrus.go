package config

type Logrus struct {
	Level          string `yaml:"level"`
	Format         string `yaml:"format"`
	Prefix         string `yaml:"prefix"`
	Directory      string `yaml:"directory"`
	ShowLine       bool   `yaml:"showLine"`
	EncodeLevel    string `yaml:"encodeLevel"`
	ShowStacktrace bool   `yaml:"showStacktrace"`
	LogInConsole   bool   `yaml:"logInConsole"`
	RetentionDay   int    `yaml:"retentionDay"`
}
