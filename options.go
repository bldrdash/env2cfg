package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/pflag"
)

type ENV map[string]string

type AppOptions struct {
	DryRun       bool            `flag:"dry-run D" desc:"Don't write to output-file."`
	Generate     bool            `flag:"gen G" desc:"Generate <dotenv> based on <template>."`
	EnvOverride  bool            `flag:"override E" desc:"Favor envfile over environment."`
	EnvVars      ENV             `flag:"vars e" desc:"Add variables from command line."`
	SetPerms     FilePermissions `flag:"perms p" desc:"Set <output> permissions." default:"0640"`
	IgnorePerm   bool            `flag:"ignore-perm I" desc:"Don't check <envfile> file permissions."`
	Quiet        bool            `flag:"quiet q" desc:"Don't display warnings."`
	ShowVersion  bool            `flag:"version v" desc:"Show version."`
	DetailedHelp bool            `flag:"detailed H" desc:"Show detailed help and example."`
	DelimStart   string          `desc:"Starting delimiter string." default:"${"`
	DelimEnd     string          `desc:"Ending delimiter string." default:"}"`
	Command      string          `flag:"-"`
	EnvFile      string          `flag:"-"`
	TemplateFile string          `flag:"-"`
	OutputFile   string          `flag:"-"`
}

// Usage returns a function that can be used to print the usage of the program
func Usage(command string, detailed bool, flags *pflag.FlagSet) func() {
	return func() {
		fmt.Printf(`%[1]s reads environment variables and produces a config file based on a template`, command)

		fmt.Printf(`


Usage: 
  %[1]s [FLAGS] <template> [<dotenv>] [<output>]
	%[1]s -G <template> [<dotenv>]
	%[1]s -H
`[1:], command)

		if detailed {

			fmt.Printf(`

Details:
  %[1]s will read environment variables from the system and/or the <dotenv> file and output
	to <output>.  If <dotenv> is omitted, only the system environment will be used for variables.
	<output> is optional and will default to stdout if not provided.
	
  When envoked with the -G flag, %[1]s will generate the <dotenv> file based on variables found
	in <template> If <dotenv> is omitted, the output will be written to stdout.
	
	<template> can be in any format and will be parsed for variables using --delim-start and
	--delim-end.  The default delimiters are "${" and "}".
	
Example Template:
  mqtt:
	broker: tcp://${MQTT_BROKER}:${MQTT_PORT}
	username: ${MQTT_USER}
	password: ${MQTT_PASS}
	`[1:], command)
		}

		fmt.Printf(`
Flags:
%s`, flags.FlagUsages())
	}
}

// String satisfies the pflag.Value interface but I have no idea what it does
// Looks like you can use it to print the value of the flag or the default value
func (e *ENV) String() string {
	return ""
}

// Set satisfies the pflag.Value interface and is used to set the value of the flag
func (e *ENV) Set(value string) error {
	parts := strings.Split(value, "=")
	if len(parts) != 2 {
		return fmt.Errorf("must be in the form key=value")
	}
	(*e)[parts[0]] = parts[1]
	return nil
}

// Type satisfies the pflag.Value interface and is used to return the type of the flag
func (e *ENV) Type() string {
	return "key=value;..."
}

// NewAppOptions returns a new AppOptions struct with default values
func NewAppOptions(command string) *AppOptions {
	return &AppOptions{
		Command:    command,
		EnvVars:    make(ENV),
		DelimStart: "${",
		DelimEnd:   "}",
		SetPerms:   "0640",
	}
}

// Validate performs basic validation
func Validate() error {
	var err error

	if Options.TemplateFile == "" {
		return fmt.Errorf("<template> is required. Run %s --help for more information", Options.Command)
	}

	if Options.Generate && Options.EnvFile != "" {
		if _, err := os.Stat(Options.EnvFile); err == nil {
			return fmt.Errorf("%s already exists: refusing to overwrite", Options.EnvFile)
		}
	}

	// Don't clobber our files
	if Options.OutputFile != "" {
		if Options.OutputFile == Options.EnvFile {
			return fmt.Errorf("<dotenv> and <output> must differ")
		}
		if Options.OutputFile == Options.TemplateFile {
			return fmt.Errorf("<template> and <output> must differ")
		}
	}

	FileMode, err = Options.SetPerms.Mode()
	if err != nil {
		fmt.Fprintf(StandardError, "could not convert %s to octal: using %s\n", Options.SetPerms, FileMode.String())
	}

	return nil
}
