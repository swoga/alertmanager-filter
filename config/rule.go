package config

import "time"

type Rule struct {
	Match    []Match `yaml:"match"`
	NotMatch []Match `yaml:"not_match"`
}

func (r *Rule) checkTimeIntervals(times map[string]struct{}) error {
	var err error

	for _, m := range r.Match {
		err = m.checkTimeIntervals(times)
		if err != nil {
			break
		}
	}
	if err == nil {
		for _, m := range r.NotMatch {
			err = m.checkTimeIntervals(times)
			if err != nil {
				break
			}
		}
	}

	return err
}

func (r *Rule) IsMatch(timeIntervalsMap TimeIntervalsMap, labels map[string]string, time time.Time) bool {
	match := false
	if len(r.Match) == 0 {
		match = true
	} else {
		for _, m := range r.Match {
			if m.IsMatch(timeIntervalsMap, labels, time) {
				match = true
				break
			}
		}
	}

	for _, m := range r.NotMatch {
		if m.IsMatch(timeIntervalsMap, labels, time) {
			match = false
			break
		}
	}

	return match
}
