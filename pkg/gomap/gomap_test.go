package gomap_test

import (
	"fmt"
	"github.com/ispapp/psshclient/pkg/gomap"
	"testing"
)

// run test by:
func TestMain(m *testing.M) {
	var (
		proto    = "tcp"
		fastscan = true
		stealth  = false
	)

	// results, err := gomap.ScanIP("192.168.1.1", proto, fastscan, stealth)
	results, err := gomap.ScanRange(proto, fastscan, stealth, 22)
	if err != nil {
		panic(err)
	} else {
		fmt.Println(results.String())
	}

	j, err := results.Json()
	if err != nil {
		panic(err)
	} else {
		fmt.Println(j)
	}
}
