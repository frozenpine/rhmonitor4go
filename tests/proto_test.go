package tests

import (
	"testing"

	rohon "github.com/frozenpine/rhmonitor4go"
	"github.com/frozenpine/rhmonitor4go/service"
)

func TestSerial(t *testing.T) {
	risk_user := rohon.RiskUser{
		UserID:     "abcdefghijklmn",
		Password:   "lkjaklsdfdsf",
		MACAddress: "AA-BB-CC-DD-EE-FF",
	}

	proto_risk_user := service.RiskUser{
		UserId:   risk_user.UserID,
		Password: risk_user.Password,
		MacAddr:  risk_user.MACAddress,
	}

	buffer, err := proto_risk_user.Marshal()
	if err != nil {
		t.Fatal(err)
	}

	t.Log(buffer, len(buffer), len(risk_user.UserID)+len(risk_user.Password)+len(risk_user.MACAddress))
}
