package main

import (
	"fmt"
	"io/fs"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type FilePermissions string

func (fp FilePermissions) Mode() (os.FileMode, error) {
	mode, err := strconv.ParseInt(string(fp), 8, 64)
	if err != nil {
		return os.FileMode(0640), err
	}
	return os.FileMode(mode), nil
}

func (fp FilePermissions) Octal() (int64, error) {
	oct, err := strconv.ParseInt(string(fp), 8, 64)
	if err != nil {
		return 0, err
	}
	return oct, err
}

// Chmod changes the file's permissions to mode
// This nessesary as OpenFile with os.O_WRONLY|os.O_CREATE|os.O_TRUNC
// doesn't change filter permissions if the file exists. Due to OpenFile
// calling syscall.open which is influenced by umask.
func Chmod(file string, mode fs.FileMode) {
	if err := aFS.Chmod(file, mode); err != nil {
		fmt.Fprintln(StandardError, err.Error())
	}
}

// CheckFilePerms checks a file's permissions match specified
func CheckFilePerms(filename string, fp FilePermissions) error {
	info, err := aFS.Stat(filename)
	if err != nil {
		return err
	}
	fm := info.Mode().Perm()
	wantPerms, _ := fp.Octal()
	if fm^fs.FileMode(wantPerms) != 0 {
		return fmt.Errorf("%s doesn't have %s permissions. Use -p or -P to change behavior", filename, fp)
	}
	return nil
}

// LoadEnvFile loads specified dotenv file as a map[string]string
func LoadEnvFile(envFile string) map[string]string {
	envVars, e := godotenv.Read(envFile)
	if e != nil {
		Fatal("Error reading env file: %s\n", e.Error())
	}
	return envVars
}
