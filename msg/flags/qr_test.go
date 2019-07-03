package DNSFlags

import "testing"

func TestQ(t *testing.T) {
	var flags uint16 = 0x8000
	if HasQ(flags) {
		t.Errorf("has Q but undetected")
		return
	}
}

func TestNotQ(t *testing.T) {
	var flags uint16 = 0x0000
	if !HasQ(flags) {
		t.Errorf("has no Q but detected")
		return
	}
}

func TestR(t *testing.T) {
	var flags uint16 = 0x0000
	if HasR(flags) {
		t.Errorf("has R but undetected")
		return
	}
}

func TestNotR(t *testing.T) {
	var flags uint16 = 0x8000
	if !HasR(flags) {
		t.Errorf("has no R but detected")
		return
	}
}
