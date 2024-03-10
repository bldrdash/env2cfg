package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/spf13/afero"
)

func TestFindVar(t *testing.T) {
	Options = NewAppOptions("env2cfg")

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

func TestInitFromCLI(t *testing.T) {
	// Set up test environment
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"env2cfg", "example/config.tmpl.yaml", "envfile.env", "output.txt"}

	// Call the function under test
	fs, _ := InitFromCLI()

	// Verify the options
	if Options.Command != "env2cfg" {
		t.Errorf("Expected Options.Command to be 'env2cfg', got '%s'", Options.Command)
	}
	if Options.TemplateFile != "example/config.tmpl.yaml" {
		t.Errorf("Expected Options.TemplateFile to be 'example/config.tmpl.yaml', got '%s'", Options.TemplateFile)
	}
	if Options.EnvFile != "envfile.env" {
		t.Errorf("Expected Options.EnvFile to be 'envfile.env', got '%s'", Options.EnvFile)
	}
	if Options.OutputFile != "output.txt" {
		t.Errorf("Expected Options.OutputFile to be 'output.txt', got '%s'", Options.OutputFile)
	}

	// Verify the flag set
	if fs.Usage == nil {
		t.Error("Expected flag set usage to be set")
	}
	if fs.SortFlags {
		t.Error("Expected flag set SortFlags to be false")
	}

	// Verify the error handling
	t.Run("Error parsing flags", func(t *testing.T) {

		// Call the function under test with invalid flags
		os.Args = []string{"env2cfg", "--invalid-flag"}
		_, err := InitFromCLI()

		expectedErrMsg := "unknown flag: --invalid-flag"
		if err != nil && err.Error() != expectedErrMsg {
			t.Errorf("Expected error message '%s', got '%s'", expectedErrMsg, err.Error())
		}
	})
}
