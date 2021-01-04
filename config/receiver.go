package config

type Receiver struct {
	Rules  []Rule `yaml:"rules"`
	Target Target `yaml:"target"`
}
