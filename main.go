package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"

	"github.com/bldrdash/sflags/gen/gpflag"
	"github.com/spf13/afero"
	"github.com/spf13/pflag"
	"github.com/valyala/fasttemplate"
)

var (
	BuildTime     string
	Githash       string
	EnvFileVars   map[string]string
	Warnings      []string
	FileMode      fs.FileMode = 0640
	aFS                       = afero.NewOsFs()
	StandardError io.Writer   = os.Stderr
	Options       *AppOptions
)

func init() {
	EnvFileVars = make(map[string]string)
	Warnings = make([]string, 0)
	Options = NewAppOptions("env2cfg")

}

// InitFromCLI initializes the application options from the command line arguments.
func InitFromCLI() (*pflag.FlagSet, error) {

	fs := pflag.NewFlagSet(Options.Command, pflag.ContinueOnError)
	err := gpflag.ParseTo(Options, fs)
	if err != nil {
		panic(err)
	}

	fs.Usage = Usage(Options.Command, false, fs)
	fs.SortFlags = false

	err = fs.Parse(os.Args[1:])
	if err != nil {
		return fs, err
	}

	if Options.DetailedHelp {
		Usage(Options.Command, true, fs)
		os.Exit(0)
	}

	if Options.ShowVersion {
		ShowVersion()
		os.Exit(0)
	}

	Options.TemplateFile = fs.Arg(0)
	Options.EnvFile = fs.Arg(1)
	Options.OutputFile = fs.Arg(2)
	return fs, nil
}

func main() {

	Options = NewAppOptions("env2cfg")

	if fs, err := InitFromCLI(); err != nil {
		if err != pflag.ErrHelp {
			fs.Usage()
			fmt.Fprintf(os.Stderr, "\n%s\n", err.Error())
		}
		os.Exit(1)
	}

	if err := Validate(); err != nil {
		Fatal(err.Error())
	}

	contents, err := os.ReadFile(Options.TemplateFile)
	if err != nil {
		Fatal("Error reading template file: %s\n", err.Error())
	}

	if Options.Generate {
		GenerateEnvFile(contents)
	} else {
		ProcessTemplate(contents)
	}

	DisplayWarnings()

	if len(Warnings) != 0 {
		os.Exit(1)
	}
}

// ProcessTemplate reads environment variables and writes to outFile
// using template tplFile
func ProcessTemplate(contents []byte) {

	output := make([]string, 0)

	// Include variables from file (dotenv)
	if Options.EnvFile != "" {
		EnvFileVars = LoadEnvFile(Options.EnvFile)
		if !Options.NoEnvPerms {
			if err := CheckFilePerms(Options.EnvFile, Options.OutputPerms); err != nil && !Options.Quiet {
				Fatal(err.Error())
			}
		}
	}

	scanner := bufio.NewScanner(bytes.NewReader(contents))
	for scanner.Scan() {
		line := scanner.Text()
		t := fasttemplate.New(line, Options.DelimStart, Options.DelimEnd)
		s, err := t.ExecuteFuncStringWithErr(func(w io.Writer, tag string) (int, error) {
			if val, found := FindVar(tag); found {
				return w.Write([]byte(val))
			}
			return 0, fmt.Errorf("no matching key for %s", tag)
		})
		if err != nil {
			AddWarning(err.Error())
			output = append(output, line)
		} else {
			output = append(output, s)
		}
	}
	if scanner.Err() != nil {
		fmt.Fprintf(StandardError, "error: %s\n", scanner.Err())
	}
	if Options.OutputFile != "" {
		rendreredContents := strings.Join(output, "\n") + "\n"
		os.WriteFile(Options.OutputFile, []byte(rendreredContents), FileMode)
	} else {
		fmt.Fprintf(os.Stdout, "%s\n", strings.Join(output, "\n"))
	}
}

// GenerateEnvFile writes a list of variables from <template> to <dotenv> or stdout
func GenerateEnvFile(contents []byte) {

	dupCheck := make(map[string]interface{}, 0)
	keys := make([]string, 0)

	scanner := bufio.NewScanner(bytes.NewReader(contents))
	for scanner.Scan() {
		line := scanner.Text()
		t := fasttemplate.New(line, Options.DelimStart, Options.DelimEnd)
		_, err := t.ExecuteFuncStringWithErr(func(w io.Writer, tag string) (int, error) {
			if _, found := dupCheck[tag]; !found {
				dupCheck[tag] = nil
				keys = append(keys, tag)
			}
			return 0, nil
		})
		if err != nil {
			AddWarning(err.Error())
		}
	}
	if scanner.Err() != nil {
		fmt.Fprintf(StandardError, "error: %s\n", scanner.Err())
	}

	if Options.EnvFile != "" {
		os.WriteFile(Options.EnvFile, []byte(strings.Join(keys, "=\n")), FileMode)
	} else {
		fmt.Fprintf(os.Stdout, "%s\n", strings.Join(keys, "=\n"))
	}
}

// FindVar returns a value if the key can be found
func FindVar(tag string) (string, bool) {

	// Command line takes precedence
	if val, found := Options.EnvVars[tag]; found {
		return val, true
	}

	// Environment comes next
	val, found := os.LookupEnv(tag)
	if !found || Options.EnvOverride {
		val, found = EnvFileVars[tag]
	}
	return val, found
}

// AddWarning queues up warnings to be printed at the end
func AddWarning(s string, w ...interface{}) {
	Warnings = append(Warnings, fmt.Sprintf(s, w...))
}

// DisplayWarnings prints the queued warnings
func DisplayWarnings() {
	if len(Warnings) != 0 && !Options.Quiet {
		fmt.Fprintln(StandardError, "----------------------------------------------------------------------------")
		fmt.Fprintf(StandardError, "The following variables were referenced in %s but not found in any environment input:\n",
			Options.TemplateFile)
		for _, msg := range Warnings {
			fmt.Fprintln(StandardError, msg)
		}
	}
}

// Fatal writes to stderr and exits with a value of 1
func Fatal(s string, w ...interface{}) {
	fmt.Fprintf(StandardError, s+"\n", w...)
	os.Exit(1)
}
