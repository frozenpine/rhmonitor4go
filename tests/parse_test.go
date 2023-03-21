package tests

import (
	"regexp"
	"testing"
)

func TestParse(t *testing.T) {
	conn := "tcp://210.22.96.58:11102"
	riskSvrConn := regexp.MustCompile("tcp://([0-9.]+):([0-9]+)")

	match := riskSvrConn.FindStringSubmatch(conn)

	if len(match) != 3 {
		t.Fatal("mismatch")
	}

	t.Log(match)
}
