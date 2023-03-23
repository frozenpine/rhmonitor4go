package tests

import (
	"encoding/json"
	"testing"

	rohon "github.com/frozenpine/rhmonitor4go"
	"github.com/frozenpine/rhmonitor4go/service"
	"github.com/gogo/protobuf/proto"
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

	deser := service.RiskUser{}

	if err = proto.Unmarshal(buffer, &deser); err != nil {
		t.Fatal(err)
	}

	jsonStr, err := json.Marshal(proto_risk_user)
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(jsonStr)
	}

	t.Log(buffer, len(buffer), len(risk_user.UserID)+len(risk_user.Password)+len(risk_user.MACAddress))
}
