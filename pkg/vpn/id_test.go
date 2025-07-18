package vpn

import (
	"encoding/json"
	"testing"
)

func verifyId[T int | uint8](t *testing.T, got, want T) {
	t.Helper()
	if got != want {
		t.Error(got, "!=", want)
	}
}

func TestId(t *testing.T) {
	id0 := Id(0)
	id1 := Id(1)
	verifyId(t, id0.Index(), 0)
	verifyId(t, id1.Index(), 1)
	verifyId(t, id0.Version(), 0)
	verifyId(t, id1.Version(), 0)
	t.Log("id0:", id0)
	t.Log("id1:", id1)
	id0.Revise()
	id1.Revise()
	verifyId(t, id0.Index(), 0)
	verifyId(t, id1.Index(), 1)
	verifyId(t, id0.Version(), 1)
	verifyId(t, id1.Version(), 1)
	t.Log("id0:", id0)
	t.Log("id1:", id1)
	data0, err := json.Marshal(id0)
	if err != nil {
		t.Fatal(err)
	}
	data1, err := json.Marshal(id1)
	if err != nil {
		t.Fatal(err)
	}

	var new0, new1 Id
	if err = json.Unmarshal(data0, &new0); err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(data1, &new1); err != nil {
		t.Fatal(err)
	}
	verifyId(t, new0.Index(), 0)
	verifyId(t, new1.Index(), 1)
	verifyId(t, new0.Version(), 1)
	verifyId(t, new1.Version(), 1)
	t.Log("new0:", new0)
	t.Log("new1:", new1)
}
