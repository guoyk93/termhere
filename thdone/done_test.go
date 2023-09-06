package thdone

import "testing"

func TestDone_Close(t *testing.T) {
	d := New()
	d.Close()
	d.Close()
	d.Close()
	d.Close()
	select {
	case <-d.C:
	default:
		t.Fatal("expected closed channel")
	}
}
