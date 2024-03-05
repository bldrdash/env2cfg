package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/spf13/afero"
)

func TestFindVar(t *testing.T) {
	// From command line
	Options.EnvVars = map[string]string{
		"TAG1": "value1",
		"TAG2": "value2",
	}

	// From dotenv file
	EnvFileVars = map[string]string{
		"TAG3": "value3",
	}

	// Test when the value is found in Options.EnvVars
	val, found := FindVar("TAG1")
	if val != "value1" || !found {
		t.Errorf("Expected value1, found %s", val)
	}

	// Test when the value is not found in Options.EnvVars but found in os.LookupEnv
	os.Setenv("TAG2", "value2")
	val, found = FindVar("TAG2")
	if val != "value2" || !found {
		t.Errorf("Expected value2, found %s", val)
	}

	// Test when the value is not found in Options.EnvVars and os.LookupEnv, but found in EnvFileVars
	val, found = FindVar("TAG3")
	if val != "value3" || !found {
		t.Errorf("Expected value3, found %s", val)
	}

	// Test when the value is not found in any source
	val, found = FindVar("TAG4")
	if val != "" || found {
		t.Errorf("Expected empty value, found %s", val)
	}
}

func TestWriteLine(t *testing.T) {
	aFS = afero.NewMemMapFs()

	// Create a temporary file for testing
	tempFile, err := afero.TempFile(aFS, "", "test_file")
	// tempFile, err := aFS.WriteFile(tempFilename, []byte{}, 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tempFile.Name())

	// Write a line to the temporary file
	expectedLine := "Hello, World!"
	fmt.Fprintf(tempFile, "%s", expectedLine)
	tempFile.Close()

	// Read the contents of the temporary file
	fileContents, err := afero.ReadFile(aFS, tempFile.Name())
	if err != nil {
		t.Fatal(err)
	}

	// Convert the file contents to a string
	actualLine := string(fileContents)

	// Check if the written line matches the expected line
	if actualLine != expectedLine {
		t.Errorf("Expected line: %s, got: %s", expectedLine, actualLine)
	}
}
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
	if buf.String() != "could not convert abc to octal: using 0640\n" {
		t.Error("Expected a message, got empty string")
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
