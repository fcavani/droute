// Copyright 2017 Felipe A. Cavani. All rights reserved.

package iplists

import (
	"os"
	"testing"
	"time"
)

func TestHaveIP(t *testing.T) {
	fbc, err := NewFBCrawler(
		"whois.radb.net",
		"43",
		5*time.Second,
		"whois.json",
		"AS32934",
		"AS35995",
		"AS13414",
	)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove("whois.json")
	if !fbc.HaveIP("173.252.91.164") {
		t.Fatal("ip not found")
	}
	if !fbc.HaveIP("199.16.157.183") {
		t.Fatal("ip not found")
	}
	if fbc.HaveIP("192.168.1.1") {
		t.Fatal("ip found")
	}

	_, err = NewFBCrawler(
		"www.google.com",
		"43",
		5*time.Second,
		"whois.json",
		"AS32934",
	)
	if err != nil {
		t.Fatal(err)
	}
}
