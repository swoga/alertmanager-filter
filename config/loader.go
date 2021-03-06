package config

import (
	"fmt"
	"os"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/yaml.v3"
)

var (
	configReloadSuccess = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "alertmanager_filter",
		Name:      "config_last_reload_successful",
		Help:      "UFiber exporter config loaded successfully.",
	})

	configReloadSeconds = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "alertmanager_filter",
		Name:      "config_last_reload_success_timestamp_seconds",
		Help:      "Timestamp of the last successful configuration reload.",
	})
)

func init() {
	prometheus.MustRegister(configReloadSuccess)
	prometheus.MustRegister(configReloadSeconds)
}

//New creates a new SafeConfig instance
func New(configFile string) SafeConfig {
	return SafeConfig{
		c:          &Config{},
		configFile: configFile,
	}
}

//SafeConfig is a thread safe config handler
type SafeConfig struct {
	sync.RWMutex
	configFile string
	c          *Config
}

//Get returns the current config
func (sc *SafeConfig) Get() *Config {
	sc.Lock()
	defer sc.Unlock()
	return sc.c
}

//LoadConfig reads and parses the file from disk
func (sc *SafeConfig) LoadConfig() (err error) {
	c := &Config{}
	defer func() {
		if err != nil {
			configReloadSuccess.Set(0)
		} else {
			configReloadSuccess.Set(1)
			configReloadSeconds.SetToCurrentTime()
		}
	}()

	yamlReader, err := os.Open(sc.configFile)
	if err != nil {
		return fmt.Errorf("error reading config file: %s", err)
	}
	defer yamlReader.Close()
	decoder := yaml.NewDecoder(yamlReader)
	decoder.KnownFields(true)

	err = decoder.Decode(c)
	if err != nil {
		return fmt.Errorf("error parsing config file: %s", err)
	}

	sc.Lock()
	sc.c = c
	defer sc.Unlock()

	return nil
}
