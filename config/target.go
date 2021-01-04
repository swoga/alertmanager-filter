package config

import (
	"net/http"

	am_config "github.com/prometheus/alertmanager/config"
	commoncfg "github.com/prometheus/common/config"
)

type Target struct {
	URL        am_config.URL              `yaml:"url"`
	HTTPConfig commoncfg.HTTPClientConfig `yaml:"http_config"`
	HTTPClient *http.Client
}

// UnmarshalYAML implements the Unmarshaller interface for Target
func (t *Target) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain Target
	if err := unmarshal((*plain)(t)); err != nil {
		return err
	}

	client, err := commoncfg.NewClientFromConfig(t.HTTPConfig, "target", false)
	if err != nil {
		return err
	}
	t.HTTPClient = client

	return nil
}
