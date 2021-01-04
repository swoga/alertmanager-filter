package config

import (
	"fmt"
	"time"

	"github.com/swoga/alertmanager-filter/utils"
)

type Match struct {
	Labels map[string]utils.StringArray `yaml:"labels"`
	Times  []string                     `yaml:"times"`
}

func (m *Match) checkTimeIntervals(times map[string]struct{}) error {
	for _, time := range m.Times {
		_, ok := times[time]
		if !ok {
			return fmt.Errorf("undefined time interval: %s", time)
		}
	}
	return nil
}

func (m *Match) IsMatch(timeIntervalsMap TimeIntervalsMap, labels map[string]string, time time.Time) bool {
	return m.IsLabelMatch(labels) && m.IsTimeMatch(timeIntervalsMap, time)
}

func (m *Match) IsLabelMatch(labels map[string]string) bool {
	for mKey, mValues := range m.Labels {
		aValue, ok := labels[mKey]
		if !ok {
			return false
		}

		if !mValues.Contains(aValue) {
			return false
		}
	}

	return true
}

func (m *Match) IsTimeMatch(timeIntervalsMap TimeIntervalsMap, time time.Time) bool {
	for _, name := range m.Times {
		timeIntervals, ok := timeIntervalsMap[name]
		if !ok {
			continue
		}

		for _, timeInterval := range timeIntervals {
			if timeInterval.ContainsTime(time) {
				return true
			}
		}

	}
	return false
}
