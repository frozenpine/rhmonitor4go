package tests

import (
	"regexp"
	"testing"
)

func TestParse(t *testing.T) {
	riskSvr := "tcp://210.22.96.58:11102"
	riskSvrPattern := regexp.MustCompile("tcp://([0-9.]+):([0-9]+)")

	match := riskSvrPattern.FindStringSubmatch(riskSvr)

	if len(match) != 3 {
		t.Fatal("mismatch")
	}

	t.Log(match)

	redisConn1 := "localhost:6379@2"
	redisConn2 := "pass#localhost:6379@2"
	redisSvrPattern := regexp.MustCompile("(?:(.+)#)?([a-z0-9A-Z.:].+)@([0-9]+)")

	match = redisSvrPattern.FindStringSubmatch(redisConn1)
	t.Log(match)

	match = redisSvrPattern.FindStringSubmatch(redisConn2)
	t.Log(match)
}
