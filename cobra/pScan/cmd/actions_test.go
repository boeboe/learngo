package cmd

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/boeboe/learngo/cobra/pScan/scan"
)

func setup(t *testing.T, hosts []string, initList bool) (string, func()) {
	// create temp file
	tf, err := ioutil.TempFile("", "pScan")
	if err != nil {
		t.Fatal(err)
	}
	tf.Close()

	// initialize list if needed
	if initList {
		hl := &scan.HostsList{}
		for _, h := range hosts {
			if err := hl.Add(h); err != nil {
				t.Fatal(err)
			}
		}
		if err := hl.Save(tf.Name()); err != nil {
			t.Fatal(err)
		}
	}

	return tf.Name(), func() {
		os.Remove(tf.Name())
	}
}

func TestHostActions(t *testing.T) {

	hosts := []string{"host1", "host2", "host3"}
	testCases := []struct {
		name       string
		args       []string
		expOut     string
		initList   bool
		actionFunc func(io.Writer, string, []string) error
	}{
		{
			name:       "AddAction",
			args:       hosts,
			expOut:     "Added host: host1\nAdded host: host2\nAdded host: host3\n",
			initList:   false,
			actionFunc: addAction,
		},
		{
			name:       "ListAction",
			expOut:     "host1\nhost2\nhost3\n",
			initList:   true,
			actionFunc: listAction,
		},
		{
			name:       "DeleteAction",
			args:       []string{"host1", "host2"},
			expOut:     "Deleted host: host1\nDeleted host: host2\n",
			initList:   true,
			actionFunc: deleteAction,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tf, cleanup := setup(t, hosts, tc.initList)
			defer cleanup()

			// define var to capture action output
			var out bytes.Buffer

			// execute action and capture output
			if err := tc.actionFunc(&out, tf, tc.args); err != nil {
				t.Fatalf("expected no error, got %q instead\n", err)
			}

			// test actions output
			if out.String() != tc.expOut {
				t.Errorf("expected output %q, but got %q instead\n", tc.expOut, out.String())
			}
		})
	}
}

func TestIntegration(t *testing.T) {
	hosts := []string{"host1", "host2", "host3"}

	// setup intrgration test
	tf, cleanup := setup(t, hosts, false)
	defer cleanup()

	delHost := "host2"
	hostsEnd := []string{"host1", "host3"}

	// define var to capture output
	var out bytes.Buffer

	expectedOut := ""
	for _, v := range hosts {
		expectedOut += fmt.Sprintf("Added host: %s\n", v)
	}
	expectedOut += strings.Join(hosts, "\n")
	expectedOut += fmt.Sprintln()
	expectedOut += fmt.Sprintf("Deleted host: %s\n", delHost)
	expectedOut += strings.Join(hostsEnd, "\n")
	expectedOut += fmt.Sprintln()

	// add hosts to the list
	if err := addAction(&out, tf, hosts); err != nil {
		t.Fatalf("expected no error, got %q instead\n", err)
	}
	// list hosts in the list
	if err := listAction(&out, tf, nil); err != nil {
		t.Fatalf("expected no error, got %q instead\n", err)
	}
	// delete hosts to the list
	if err := deleteAction(&out, tf, []string{delHost}); err != nil {
		t.Fatalf("expected no error, got %q instead\n", err)
	}
	// list hosts in the list
	if err := listAction(&out, tf, nil); err != nil {
		t.Fatalf("expected no error, got %q instead\n", err)
	}

	// test integration output
	if out.String() != expectedOut {
		t.Errorf("expected ouput %q, but got %q instead\n", expectedOut, out.String())
	}
}
