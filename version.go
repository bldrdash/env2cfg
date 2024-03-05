package main

import (
	"fmt"

	"github.com/alecthomas/kong"
)

var version = "v0.0.2"

func (v versionFlag) Decode(_ *kong.DecodeContext) error { return nil }
func (v versionFlag) IsBool() bool                       { return true }
func (v versionFlag) BeforeApply(app *kong.Kong) error {
	fmt.Fprintln(app.Stderr, "env2cfg")
	fmt.Fprintf(app.Stderr, "Version: %s\nBuild Time: %s\nGit Hash: %s\n", version, BuildTime, Githash)
	app.Exit(0)
	return nil
}
