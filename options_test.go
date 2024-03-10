package main

import (
	"bytes"
	"io"
	"os"
	"testing"
)

func TestValidate(t *testing.T) {

	// Test when Options.OutputFile is the same as Options.EnvFile
	Options.OutputFile = Options.EnvFile
	Options.TemplateFile = "template.txt"
	err := Validate()
	if err == nil {
		t.Error("Expected an error, got nil")
	}

	// Test when Options.OutputFile is the same as Options.TemplateFile
	Options.OutputFile = "output.txt"
	Options.TemplateFile = Options.OutputFile
	err = Validate()
	if err == nil {
		t.Error("Expected an error, got nil")
	}

	// Test when Options.OutputPerms.Mode() returns an error
	var buf bytes.Buffer
	StandardError = io.Writer(&buf)
	defer func() {
		StandardError = os.Stderr
	}()

	Options.OutputFile = "output.txt"
	Options.TemplateFile = "template.txt"
	Options.OutputPerms = FilePermissions("abc") // Invalid file mode
	err = Validate()
	if err != nil {
		t.Error("Expected a nil, got error")
	}
	if FileMode != 0640 {
		t.Errorf("Expected 0640, got %v", FileMode)
	}
	StandardError = os.Stderr

	// Test when Options.OutputPerms.Mode() succeeds
	Options.OutputPerms = FilePermissions("0644") // Valid file mode
	err = Validate()
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if FileMode != 0644 {
		t.Errorf("Expected 0644, got %v", FileMode)
	}
}
