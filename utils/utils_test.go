package utils

import "testing"

func TestStringArray(t *testing.T) {
	array := StringArray{"a", "b", "c"}
	ok := "a"
	nok := "d"

	if !array.Contains(ok) {
		t.Errorf("StringArray %s not found in %v", ok, array)
	}
	if array.Contains(nok) {
		t.Errorf("StringArray %s found in %v", nok, array)
	}
}
