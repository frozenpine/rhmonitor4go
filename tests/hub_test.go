package tests

import (
	"testing"

	"github.com/gofrs/uuid"
)

func TestID(t *testing.T) {
	id1 := uuid.NewV3(uuid.NamespaceDNS, "test1")
	id2 := uuid.NewV3(uuid.NamespaceDNS, "test1")

	t.Log(id1, id2)
}
