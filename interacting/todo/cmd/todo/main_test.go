package main_test

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

var (
	binName  = "todo"
	fileName = ".todo.json"
)

func TestMain(m *testing.M) {
	fmt.Println("Building tool...")
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}
	build := exec.Command("go", "build", "-o", binName)
	if err := build.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Cannot build tool %s: %s", binName, err)
		os.Exit(1)
	}

	fmt.Println("Running tests...")
	result := m.Run()

	fmt.Println("CLeaning up...")
	os.Remove(binName)
	os.Remove(fileName)

	os.Exit(result)
}

func TestTodoCLI(t *testing.T) {
	task1 := "test task number 1"
	task2 := "test task number 2"

	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	cmdPath := filepath.Join(dir, binName)

	t.Run("AddNewTaskFromArgs", func(t *testing.T) {
		cmd := exec.Command(cmdPath, "-add", task1)

		if err := cmd.Run(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("AddNewTaskFromStdIn", func(t *testing.T) {
		cmd := exec.Command(cmdPath, "-add")
		cmdStdIn, err := cmd.StdinPipe()
		if err != nil {
			t.Fatal(err)
		}
		io.WriteString(cmdStdIn, task2)
		cmdStdIn.Close()

		if err := cmd.Run(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("ListTasks", func(t *testing.T) {
		cmd := exec.Command(cmdPath, "-list")

		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatal(err)
		}

		expected := fmt.Sprintf("%s(%d) %s\n%s(%d) %s\n", "   ", 1, task1, "   ", 2, task2)
		if expected != string(out) {
			t.Errorf("Expected %q, got %q instead\n", expected, string(out))
		}
	})

	t.Run("CompleteTask", func(t *testing.T) {
		cmdComplete := exec.Command(cmdPath, "-complete", "1")

		if err := cmdComplete.Run(); err != nil {
			t.Fatal(err)
		}

		cmdList := exec.Command(cmdPath, "-list")

		out, err := cmdList.CombinedOutput()
		if err != nil {
			t.Fatal(err)
		}

		expected := fmt.Sprintf("%s(%d) %s\n%s(%d) %s\n", " X ", 1, task1, "   ", 2, task2)
		if expected != string(out) {
			t.Errorf("Expected %q, got %q instead\n", expected, string(out))
		}
	})

	t.Run("DeleteTask", func(t *testing.T) {
		cmdDelete := exec.Command(cmdPath, "-delete", "2")

		if err := cmdDelete.Run(); err != nil {
			t.Fatal(err)
		}

		cmdList := exec.Command(cmdPath, "-list")

		out, err := cmdList.CombinedOutput()
		if err != nil {
			t.Fatal(err)
		}

		expected := fmt.Sprintf("%s(%d) %s\n", " X ", 1, task1)
		if expected != string(out) {
			t.Errorf("Expected %q, got %q instead\n", expected, string(out))
		}
	})
}
