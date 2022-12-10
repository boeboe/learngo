package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestRun(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed, skipping test")
	}

	var testCases = []struct {
		name     string
		proj     string
		out      string
		expErr   error
		setupGit bool
	}{
		{name: "success",
			proj:     "./testdata/tool",
			out:      "go build: SUCCESS\ngo test: SUCCESS\ngofmt: SUCCESS\ngit push: SUCCESS\n",
			expErr:   nil,
			setupGit: true},
		{name: "failed",
			proj:     "./testdata/toolErr",
			out:      "",
			expErr:   &stepErr{step: "go build"},
			setupGit: false},
		{name: "failFormat",
			proj:     "./testdata/toolFmtErr",
			out:      "",
			expErr:   &stepErr{step: "go fmt"},
			setupGit: false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setupGit {
				cleanup := setupGit(t, tc.proj)
				defer cleanup()
			}

			var out bytes.Buffer

			err := run(tc.proj, &out)
			if tc.expErr != nil {
				if err == nil {
					t.Errorf("expected error %q, got 'nil' instead", tc.expErr)
					return
				}
				if !errors.Is(err, tc.expErr) {
					t.Errorf("expected error %q, got %q", tc.expErr, err)
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error %q", err)
			}
			if out.String() != tc.out {
				t.Errorf("expected output: %q, but got %q instead", tc.out, out.String())
			}
		})
	}
}

func setupGit(t *testing.T, proj string) func() {
	t.Helper()

	gitExec, err := exec.LookPath("git")
	if err != nil {
		t.Fatal(err)
	}

	tmpDir, err := ioutil.TempDir("", "gocitest")
	if err != nil {
		t.Fatal(err)
	}
	remoteURI := fmt.Sprintf("file://%s", tmpDir)

	projPath, err := filepath.Abs(proj)
	if err != nil {
		t.Fatal(err)
	}

	var gitCmdList = []struct {
		args []string
		dir  string
		env  []string
	}{
		{args: []string{"init", "--bare"}, dir: tmpDir, env: nil},
		{args: []string{"init"}, dir: projPath, env: nil},
		{args: []string{"remote", "add", "origin", remoteURI}, dir: projPath, env: nil},
		{args: []string{"add", "."}, dir: projPath, env: nil},
		{args: []string{"commit", "-m", "test"}, dir: projPath,
			env: []string{
				"GIT_COMMITTER_NAME=test",
				"GIT_COMMITTER_EMAIL=test@example.com",
				"GIT_AUTHOR_NAME=test",
				"GIT_AUTHOR_EMAIL=test@example.com",
			}},
	}

	for _, g := range gitCmdList {
		gitCmd := exec.Command(gitExec, g.args...)
		gitCmd.Dir = g.dir

		if g.args != nil {
			gitCmd.Env = append(os.Environ(), g.env...)
		}

		if err := gitCmd.Run(); err != nil {
			t.Fatal(err)
		}
	}

	return func() {
		os.RemoveAll(tmpDir)
		os.RemoveAll(filepath.Join(projPath, ".git"))
	}
}
