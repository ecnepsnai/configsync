package main

import "testing"

func TestPathWithoutFile(t *testing.T) {
	filePath := "/etc/passwd"
	expected := "/etc"
	result := pathWithoutFile(filePath)
	if result != expected {
		t.Errorf("Unexpected result. Expected '%s' got '%s'", expected, result)
	}
}
