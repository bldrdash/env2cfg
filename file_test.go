package main

import (
	"bufio"
	"bytes"
	"io/fs"
	"os"
	"testing"

	"github.com/spf13/afero"
)

func TestFilePermissions_Mode(t *testing.T) {
	tests := []struct {
		name     string
		fp       FilePermissions
		expected os.FileMode
		wantErr  bool
	}{
		{
			name:     "Valid octal string",
			fp:       FilePermissions("0644"),
			expected: os.FileMode(0644),
			wantErr:  false,
		},
		{
			name:     "Invalid octal string",
			fp:       FilePermissions("abc"),
			expected: os.FileMode(0640),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mode, err := tt.fp.Mode()
			if (err != nil) != tt.wantErr {
				t.Errorf("FilePermissions.Mode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if mode != tt.expected {
				t.Errorf("FilePermissions.Mode() = %v, want %v", mode, tt.expected)
			}
		})
	}
}

func TestFilePermissions_Octal(t *testing.T) {
	tests := []struct {
		name     string
		fp       FilePermissions
		expected int64
		wantErr  bool
	}{
		{
			name:     "Valid octal string",
			fp:       FilePermissions("0644"),
			expected: 420,
			wantErr:  false,
		},
		{
			name:     "Invalid octal string",
			fp:       FilePermissions("abc"),
			expected: 0,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			octal, err := tt.fp.Octal()
			if (err != nil) != tt.wantErr {
				t.Errorf("FilePermissions.Octal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if octal != tt.expected {
				t.Errorf("FilePermissions.Octal() = %v, want %v", octal, tt.expected)
			}
		})
	}
}

func TestCheckFilePerms(t *testing.T) {
	aFS = afero.NewMemMapFs()

	tests := []struct {
		name        string
		filename    string
		fp          FilePermissions
		expectedErr bool
		expectedMsg string
	}{
		{
			name:        "Valid permissions",
			filename:    "testfile.txt",
			fp:          FilePermissions("0644"),
			expectedErr: false,
			expectedMsg: "",
		},
		{
			name:        "Invalid permissions",
			filename:    "testfile.txt",
			fp:          FilePermissions("0755"),
			expectedErr: true,
			expectedMsg: "testfile.txt doesn't have 0755 permissions. Use -p or -P to change behavior",
		},
	}

	afero.WriteFile(aFS, "testfile.txt", []byte("data"), 0644)
	aFS.Chmod("testfile.txt", 0644)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckFilePerms(tt.filename, tt.fp)
			if err != nil && tt.expectedErr != true {
				t.Errorf("CheckFilePerms() error = %v, expectedErr %v", err, tt.expectedErr)
			}
			if err != nil && err.Error() != tt.expectedMsg {
				t.Errorf("CheckFilePerms() error message = %v, expectedMsg %v", err.Error(), tt.expectedMsg)
			}
		})
	}
}
func TestChmod(t *testing.T) {
	aFS = afero.NewMemMapFs()

	tests := []struct {
		name         string
		file         string
		mode         fs.FileMode
		expectedErr  bool
		expectedFile fs.FileMode
	}{
		{
			name:         "Valid file and mode",
			file:         "testfile.txt",
			mode:         0644,
			expectedErr:  false,
			expectedFile: 0644,
		},
		{
			name:         "Invalid file",
			file:         "nonexistent.txt",
			mode:         0644,
			expectedErr:  true,
			expectedFile: 0,
		},
	}

	afero.WriteFile(aFS, "testfile.txt", []byte("data"), 0644)
	var buf bytes.Buffer
	StandardError = bufio.NewWriter(&buf)
	defer func() {
		StandardError = os.Stderr
	}()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Chmod(tt.file, tt.mode)

			fileInfo, err := aFS.Stat(tt.file)
			if (err != nil) != tt.expectedErr {
				t.Errorf("Chmod() error = %v, expectedErr %v", err, tt.expectedErr)
				return
			}

			if fileInfo != nil && fileInfo.Mode() != tt.expectedFile {
				t.Errorf("Chmod() file mode = %v, expected %v", fileInfo.Mode(), tt.expectedFile)
			}
		})
	}

}
