package scheduler

import "testing"

func TestGetPrimaryDomain(t *testing.T) {
	host := "127.0.0.1"
	pd, err := getPrimaryDomain(host)
	if err != nil {
		t.Fatalf("An error occurs when getting primary domain: %s (host: %s)", err, host)
	}
	if pd != host {
		t.Fatalf("Inconsistent primay domain, expected: %s, actual: %s", host, pd)
	}
	host = "cn.bing.com"
	pd, err = getPrimaryDomain(host)
	if err != nil {
		t.Fatalf("An error occurs when getting the primary domain: %s, (host: %s)", err, host)
	}
	expectedPd := "bing.com"
	if pd != expectedPd {
		t.Fatalf("Inconsistent primary domain, expected: %s, actual: %s", expectedPd, pd)
	}
	_, err = getPrimaryDomain("")
	if err == nil {
		t.Fatalf("It can still get primary domain for an empty host")
	}
	host = "123.abc"
	_, err = getPrimaryDomain(host)
	if err == nil {
		t.Fatalf("It can still get primary domain for a unrecognized host %q", host)
	}
}
