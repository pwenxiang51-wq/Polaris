package main

import (
	"net"
	"testing"
)

func TestParseProbeTargetsDefaultCandidates(t *testing.T) {
	targets, err := parseProbeTargets("")
	if err != nil {
		t.Fatal(err)
	}
	if len(targets) != 3 {
		t.Fatalf("expected three default targets, got %d", len(targets))
	}
	if targets[0].Name != "CT" || targets[1].Name != "CU" || targets[2].Name != "CM" {
		t.Fatalf("unexpected default target labels: %#v", targets)
	}
}

func TestParseProbeTargetsSupportsCustomAndIPv6Targets(t *testing.T) {
	targets, err := parseProbeTargets("Los-Angeles=1.1.1.1:443;IPv6=[2606:4700:4700::1111]:443")
	if err != nil {
		t.Fatal(err)
	}
	if len(targets) != 2 || targets[1].Address != "[2606:4700:4700::1111]:443" {
		t.Fatalf("custom target configuration was not parsed: %#v", targets)
	}
}

func TestParseProbeTargetsRejectsInvalidInput(t *testing.T) {
	for _, input := range []string{
		"CT=127.0.0.1:53;CT=1.1.1.1:443",
		"CT=example.com",
		"CT=example.com:99999",
		"=example.com:443",
	} {
		if _, err := parseProbeTargets(input); err == nil {
			t.Errorf("expected invalid target string %q to fail", input)
		}
	}
}

func TestPublicRouteAddressFilter(t *testing.T) {
	cases := []struct {
		value string
		want  bool
	}{
		{"8.8.8.8", true},
		{"203.0.113.10", false},
		{"192.168.1.2", false},
		{"10.0.0.1", false},
		{"127.0.0.1", false},
		{"100.64.0.1", false},
		{"::1", false},
		{"2001:db8::1", false},
	}
	for _, tc := range cases {
		if got := isPublicRoutedIP(net.ParseIP(tc.value)); got != tc.want {
			t.Errorf("isPublicRoutedIP(%s) = %v, want %v", tc.value, got, tc.want)
		}
	}
}
