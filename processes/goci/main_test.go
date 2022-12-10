package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

func TestRun(t *testing.T) {

	var testCases = []struct {
		name     string
		proj     string
		out      string
		expErr   error
		setupGit bool
		mockCmd  func(ctx context.Context, exe string, args ...string) *exec.Cmd
	}{
		{name: "success",
			proj:     "./testdata/tool",
			out:      "go build: SUCCESS\ngo test: SUCCESS\ngofmt: SUCCESS\ngit push: SUCCESS\n",
			expErr:   nil,
			setupGit: true,
			mockCmd:  nil},
		{name: "successMock",
			proj:     "./testdata/tool",
			out:      "go build: SUCCESS\ngo test: SUCCESS\ngofmt: SUCCESS\ngit push: SUCCESS\n",
			expErr:   nil,
			setupGit: false,
			mockCmd:  mockCmdContext},
		{name: "failed",
			proj:     "./testdata/toolErr",
			out:      "",
			expErr:   &stepErr{step: "go build"},
			setupGit: false,
			mockCmd:  nil},
		{name: "failFormat",
			proj:     "./testdata/toolFmtErr",
			out:      "",
			expErr:   &stepErr{step: "go fmt"},
			setupGit: false,
			mockCmd:  nil},
		{name: "failTimeout",
			proj:     "./testdata/tool",
			out:      "",
			expErr:   context.DeadlineExceeded,
			setupGit: false,
			mockCmd:  mockCmdTimeout},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setupGit {
				if _, err := exec.LookPath("git"); err != nil {
					t.Skip("git not installed, skipping test")
				}
				cleanup := setupGit(t, tc.proj)
				defer cleanup()
			}

			if tc.mockCmd != nil {
				command = tc.mockCmd
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

func mockCmdContext(ctx context.Context, exe string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess"}
	cs = append(cs, exe)
	cs = append(cs, args...)

	cmd := exec.CommandContext(ctx, os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	return cmd
}

func mockCmdTimeout(ctx context.Context, exe string, args ...string) *exec.Cmd {
	cmd := mockCmdContext(ctx, exe, args...)
	cmd.Env = append(cmd.Env, "GO_HELPER_TIMEOUT=1")
	return cmd
}

func TestHelperProcess(t *testing.T) {
	t.Helper()
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	if os.Getenv("GO_HELPER_TIMEOUT") == "1" {
		time.Sleep(15 * time.Second)
	}

	if os.Args[2] == "git" {
		fmt.Fprintln(os.Stdout, "everything up-to-date")
		os.Exit(0)
	}
	os.Exit(1)
}

func TestRunKill(t *testing.T) {
	var testCases = []struct {
		name   string
		proj   string
		sig    syscall.Signal
		expErr error
	}{
		{name: "SigInt", proj: "./testdata/tool", sig: syscall.SIGINT, expErr: ErrSignal},
		{name: "SigTerm", proj: "./testdata/tool", sig: syscall.SIGTERM, expErr: ErrSignal},
		{name: "SigQuit", proj: "./testdata/tool", sig: syscall.SIGQUIT, expErr: nil},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			command = mockCmdTimeout

			errCh := make(chan error)
			ignSigCh := make(chan os.Signal, 1)
			expSigCh := make(chan os.Signal, 1)

			signal.Notify(ignSigCh, syscall.SIGQUIT)
			defer signal.Stop(ignSigCh)

			signal.Notify(expSigCh, tc.sig)
			defer signal.Stop(expSigCh)

			go func() {
				errCh <- run(tc.proj, io.Discard)
			}()

			go func() {
				time.Sleep(2 * time.Second)
				syscall.Kill(syscall.Getpid(), tc.sig)
			}()

			select {
			case err := <-errCh:
				if err == nil {
					t.Errorf("expected error, got 'nil' instead")
					return
				}
				if !errors.Is(err, tc.expErr) {
					t.Errorf("expected error: %q, got %q instead", tc.expErr, err)
				}
				select {
				case rec := <-expSigCh:
					if rec != tc.sig {
						t.Errorf("execpted signal %q, got %q instead", tc.sig, rec)
					}
				default:
					t.Errorf("signal not received")
				}
			case <-ignSigCh:
			}
		})
	}
}
