package scan_test

import (
	"net"
	"strconv"
	"testing"

	"github.com/boeboe/learngo/cobra/pScan/scan"
)

func TestStateString(t *testing.T) {
	ps := scan.PortState{}

	if ps.Open.String() != "closed" {
		t.Errorf("expected port state %q, but got %q instead", "closed", ps.Open.String())
	}

	ps.Open = true
	if ps.Open.String() != "open" {
		t.Errorf("expected port state %q, but got %q instead", "open", ps.Open.String())
	}
}

func TestRunHostFound(t *testing.T) {
	testCases := []struct {
		name     string
		expState string
	}{
		{name: "OpenPort", expState: "open"},
		{name: "ClosedPort", expState: "closed"},
	}

	host := "localhost"
	hl := &scan.HostsList{}
	hl.Add(host)

	ports := []int{}
	for _, tc := range testCases {
		ln, err := net.Listen("tcp", net.JoinHostPort(host, "0"))
		if err != nil {
			t.Fatal(err)
		}
		defer ln.Close()

		_, portStr, err := net.SplitHostPort(ln.Addr().String())
		if err != nil {
			t.Fatal(err)
		}
		port, err := strconv.Atoi(portStr)
		if err != nil {
			t.Fatal(err)
		}
		ports = append(ports, port)

		if tc.name == "ClosedPort" {
			ln.Close()
		}
	}

	res := scan.Run(hl, ports)

	// verify results for HostFound test
	if len(res) != 1 {
		t.Fatalf("expected 1 results, got %d instead\n", len(res))
	}
	if res[0].Host != host {
		t.Errorf("expected host %q, got %q instead\n", host, res[0].Host)
	}
	if res[0].NotFound {
		t.Errorf("expected host %q to be found\n", host)
	}
	if len(res[0].PortStates) != 2 {
		t.Fatalf("expected 2 port states, got %d instead\n", len(res[0].PortStates))
	}
	for i, tc := range testCases {
		if res[0].PortStates[i].Port != ports[i] {
			t.Errorf("unexpected port %d, got %d instead\n", ports[i], res[0].PortStates[i].Port)
		}
		if res[0].PortStates[i].Open.String() != tc.expState {
			t.Errorf("expected port %d to be %s, got %s instead\n", ports[i], tc.expState, res[0].PortStates[i].Open.String())
		}
	}
}

func TestRunHostNotFound(t *testing.T) {
	host := "389.389.389.389"
	hl := &scan.HostsList{}
	hl.Add(host)

	res := scan.Run(hl, []int{})

	// verify results for HostNotFound test
	if len(res) != 1 {
		t.Fatalf("expected 1 results, got %d instead\n", len(res))
	}
	if res[0].Host != host {
		t.Errorf("expected host %q, got %q instead\n", host, res[0].Host)
	}
	if !res[0].NotFound {
		t.Errorf("expected host %q NOT to be found\n", host)
	}
	if len(res[0].PortStates) != 0 {
		t.Fatalf("expected 0 port states, got %d instead\n", len(res[0].PortStates))
	}
}
