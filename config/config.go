package config

import (
	"github.com/swoga/alertmanager-filter/timeinterval"
)

type Config struct {
	Listen        string              `yaml:"listen"`
	MetricsPath   string              `yaml:"metrics_path"`
	AlertsPath    string              `yaml:"alerts_path"`
	Receivers     map[string]Receiver `yaml:"receivers"`
	TimeIntervals TimeIntervalsMap    `yaml:"time_intervals"`
}

type TimeIntervalsMap map[string][]timeinterval.TimeInterval

func DefaultConfig() Config {
	return Config{
		Listen:      ":80",
		MetricsPath: "/metrics",
		AlertsPath:  "/alerts",
	}
}

// UnmarshalYAML implements the Unmarshaller interface for Config
func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = DefaultConfig()

	type plain Config
	if err := unmarshal((*plain)(c)); err != nil {
		return err
	}

	times := map[string]struct{}{}

	for name := range c.TimeIntervals {
		times[name] = struct{}{}
	}

	// check if all rules contain valid time intervals
	for _, receivers := range c.Receivers {
		for _, rule := range receivers.Rules {
			err := rule.checkTimeIntervals(times)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
