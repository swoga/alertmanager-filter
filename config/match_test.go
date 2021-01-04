package config

import (
	"testing"

	"github.com/swoga/alertmanager-filter/utils"
)

func TestMatchIsMatch(t *testing.T) {
	m := Match{
		Labels: map[string]utils.StringArray{
			"label1": utils.StringArray{"value1"},
			"label2": utils.StringArray{"value2"},
		},
	}

	match := map[string]string{
		"label1": "value1",
		"label2": "value2",
	}
	largermatch := map[string]string{
		"label1": "value1",
		"label2": "value2",
		"label3": "value3",
	}
	smallermatch := map[string]string{
		"label2": "value2",
	}
	othervalue := map[string]string{
		"label1": "value1",
		"label2": "value",
	}
	otherkey := map[string]string{
		"label3": "value",
	}

	if !m.IsLabelMatch(match) {
		t.Errorf("Match.IsMatch %v not matched in %v", match, m)
	}
	if !m.IsLabelMatch(largermatch) {
		t.Errorf("Match.IsMatch %v not matched in %v", largermatch, m)
	}
	if m.IsLabelMatch(smallermatch) {
		t.Errorf("Match.IsMatch %v matched in %v", smallermatch, m)
	}
	if m.IsLabelMatch(othervalue) {
		t.Errorf("Match.IsMatch %v matched in %v", othervalue, m)
	}
	if m.IsLabelMatch(otherkey) {
		t.Errorf("Match.IsMatch %v matched in %v", otherkey, m)
	}
}
