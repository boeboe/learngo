package scan_test

import (
	"errors"
	"io/ioutil"
	"os"
	"testing"

	"github.com/boeboe/learngo/cobra/pScan/scan"
)

func TestAdd(t *testing.T) {
	testCases := []struct {
		name   string
		host   string
		expLen int
		expErr error
	}{
		{name: "AddNew", host: "host2", expLen: 2, expErr: nil},
		{name: "AddExisting", host: "host1", expLen: 1, expErr: scan.ErrExists},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			hl := &scan.HostsList{}

			// Initialize hosts list
			if err := hl.Add("host1"); err != nil {
				t.Fatal(err)
			}

			err := hl.Add(tc.host)
			if tc.expErr != nil {
				if err == nil {
					t.Fatalf("expected error, got nil instead\n")
				}
				if !errors.Is(err, tc.expErr) {
					t.Errorf("expected error %q, got %q instead\n", tc.expErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("expected no error, got %q instead\n", err)
			}
			if len(hl.Hosts) != tc.expLen {
				t.Errorf("expected hosts list length %d, got % d instead\n", tc.expLen, len(hl.Hosts))
			}
			if hl.Hosts[1] != tc.host {
				t.Errorf("expected host name %q as index 1, got %q instead\n", tc.host, hl.Hosts[1])
			}
		})
	}
}

func TestRemove(t *testing.T) {
	testCases := []struct {
		name   string
		host   string
		expLen int
		expErr error
	}{
		{name: "RemoveExisting", host: "host1", expLen: 1, expErr: nil},
		{name: "RemoveNotFound", host: "host3", expLen: 2, expErr: scan.ErrNotExists},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			hl := &scan.HostsList{}

			// Initialize hosts list
			for _, h := range []string{"host1", "host2"} {
				if err := hl.Add(h); err != nil {
					t.Fatal(err)
				}
			}

			err := hl.Remove(tc.host)
			if tc.expErr != nil {
				if err == nil {
					t.Fatalf("expected error, got nil instead\n")
				}
				if !errors.Is(err, tc.expErr) {
					t.Errorf("expected error %q, got %q instead\n", tc.expErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("expected no error, got %q instead\n", err)
			}
			if len(hl.Hosts) != tc.expLen {
				t.Errorf("expected hosts list length %d, got % d instead\n", tc.expLen, len(hl.Hosts))
			}
			if hl.Hosts[0] == tc.host {
				t.Errorf("host name %q should not be in the list\n", tc.host)
			}
		})
	}
}

func TestSaveLoad(t *testing.T) {
	hl1 := scan.HostsList{}
	hl2 := scan.HostsList{}

	hostName := "host1"
	hl1.Add(hostName)

	tf, err := ioutil.TempFile("", "hostList")
	if err != nil {
		t.Fatalf("error creating temp file: %s\n", err)
	}
	defer os.Remove(tf.Name())

	if err := hl1.Save(tf.Name()); err != nil {
		t.Fatalf("error saving list to file: %s\n", err)
	}
	if err := hl2.Load(tf.Name()); err != nil {
		t.Fatalf("error loading list from file: %s\n", err)
	}

	if hl1.Hosts[0] != hl2.Hosts[0] {
		t.Fatalf("saved host %q should match loaded host %q\n", hl1.Hosts[0], hl2.Hosts[0])
	}
}

func TestLoadNoFile(t *testing.T) {
	tf, err := ioutil.TempFile("", "hostList")
	if err != nil {
		t.Fatalf("error creating temp file: %s\n", err)
	}
	if err := os.Remove(tf.Name()); err != nil {
		t.Fatalf("error deleting temp file: %s\n", err)
	}

	h1 := &scan.HostsList{}
	if err := h1.Load(tf.Name()); err != nil {
		t.Errorf("expected no error, got %q instead\n", err)
	}
}
