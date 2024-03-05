package main

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"os"

	"github.com/alecthomas/kong"
	"github.com/spf13/afero"
	"github.com/valyala/fasttemplate"
)

type versionFlag string

var (
	BuildTime     string
	Githash       string
	EnvFileVars   map[string]string
	Warnings      []string
	FileMode      fs.FileMode
	aFS                     = afero.NewOsFs()
	StandardError io.Writer = os.Stderr
)

var Options struct {
	DryRun       bool              `short:"D" help:"Don't write to output-file."`
	EnvOverride  bool              `short:"E" help:"Favor envfile over environment."`
	EnvVars      map[string]string `short:"e" help:"Set environment variables from command line."`
	NoEnvPerms   bool              `short:"P" help:"Don't check env-file file permissions."`
	OutputPerms  FilePermissions   `short:"p" help:"Set output-file permissions." default:"0640"`
	Quiet        bool              `short:"q" help:"Don't display warnings."`
	ShowVersion  versionFlag       `name:"version" short:"v" help:"Show version"`
	DelimStart   string            `help:"Starting delimiter string." default:"${"`
	DelimEnd     string            `help:"Ending delimiter string." default:"}"`
	EnvFile      string            `arg:"" help:"File to read variables from or - to use only environment."`
	TemplateFile string            `arg:"" help:"Template file to use."`
	OutputFile   string            `arg:"" optional:"" help:"File to output completed template or stdout if omitted."`
}

func init() {
	EnvFileVars = make(map[string]string)
	Warnings = make([]string, 0)
}

func main() {

	kong.Parse(&Options,
		kong.Name("env2cfg [OPTIONS]"),
		kong.Description("Substitutes variables in template with values found in environment and/or env file."),
		kong.UsageOnError(),
	)

	if err := Validate(); err != nil {
		Fatal(err.Error())
	}

	Run()
	DisplayWarnings()

	if len(Warnings) != 0 {
		os.Exit(1)
	}
}

// Validate performs basic validation and an assignment
func Validate() error {
	var err error

	// Don't clobber our files
	if Options.OutputFile == Options.EnvFile {
		return fmt.Errorf("env-file and output-file must differ")
	}
	if Options.OutputFile == Options.TemplateFile {
		return fmt.Errorf("template-file and output-file must differ")
	}

	FileMode, err = Options.OutputPerms.Mode()
	if err != nil {
		fmt.Fprintf(StandardError, "could not convert %s to octal: using 0640\n", Options.OutputPerms)
		FileMode = 0640
	}
	return nil
}

func Run() {

	// Include variables from file (dotenv)
	if Options.EnvFile != "-" {
		EnvFileVars = LoadEnvFile(Options.EnvFile)
		if !Options.NoEnvPerms {
			if err := CheckFilePerms(Options.EnvFile, Options.OutputPerms); err != nil && !Options.Quiet {
				Fatal(err.Error())
			}
		}
	}

	// Open template
	file, err := aFS.Open(Options.TemplateFile)
	if err != nil {
		Fatal("Error reading template file: %s\n", err.Error())
	}
	defer file.Close()

	// Output file or stdout (or nil for dry-run)
	var output afero.File
	if len(Options.OutputFile) != 0 && !Options.DryRun {
		output, err = aFS.OpenFile(Options.OutputFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, FileMode)
		if err != nil {
			Fatal("Error opening output file: %s\n", err.Error())
		}
		defer func() {
			output.Close()
			Chmod(Options.OutputFile, FileMode)
		}()

	} else if !Options.DryRun {
		output = os.Stdout
	}

	ProcessTemplate(file, output)
}

// ProcessTemplate reads environment variables and writes to outFile
// using template tplFile
func ProcessTemplate(tplFile, outFile afero.File) {
	scanner := bufio.NewScanner(tplFile)
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
			WriteLine(outFile, line)
		} else {
			WriteLine(outFile, s)
		}
	}
	if scanner.Err() != nil {
		fmt.Fprintf(StandardError, "error: %s\n", scanner.Err())
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

// WriteLine writes a single line with \n to file descriptor
func WriteLine(file afero.File, line string) {
	if file != nil {
		fmt.Fprintln(file, line)
	}
}

// Fatal writes to stderr and exits with a value of 1
func Fatal(s string, w ...interface{}) {
	fmt.Fprintf(StandardError, s+"\n", w...)
	os.Exit(1)
}
