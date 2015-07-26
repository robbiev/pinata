package pinata

import (
	"fmt"
	"testing"
)

const str = "test string"

func TestOtherPinataValidString(t *testing.T) {
	p := New(str)
	result := p.String()
	if p.Error() != nil {
		t.Error()
	}
	if result != str {
		t.Error()
	}
}

func TestOtherPinataInvalidString(t *testing.T) {
	p := New(1)
	_ = p.String()
	if p.Error() == nil {
		t.Error()
	}
}

func TestOtherPinataInvalidStringAtPath(t *testing.T) {
	p := New(str)
	_ = p.StringAtPath("a", "b", "c")
	if p.Error() == nil {
		t.Error()
	} else {
		fmt.Println(p.Error())
	}
}

func TestOtherPinataInvalidStringAtPath2(t *testing.T) {
	p := New(str)
	_ = p.StringAtPath("a")
	if p.Error() == nil {
		t.Error()
	} else {
		fmt.Println(p.Error())
	}
}

func TestMapPinata(t *testing.T) {
	m := make(map[string]interface{})
	m2 := make(map[string]interface{})
	m["one"] = m2
	m2["two"] = "three"
	p := New(m)
	if "three" != p.StringAtPath("one", "two") {
		t.Error()
	}
}

func TestMapPinataFailure(t *testing.T) {
	m := make(map[string]interface{})
	m2 := make(map[string]interface{})
	m["one"] = m2
	m2["two"] = "three"
	p := New(m)
	if p.PinataAtPath("one", "two", "three", "four") != nil {
		t.Error()
	}
}

func TestMapPinataDontPropagateErrorToParent(t *testing.T) {
	m := make(map[string]interface{})
	m2 := make(map[string]interface{})
	m["one"] = m2
	m2["two"] = "three"
	p := New(m)
	child := p.PinataAtPath("one")
	_ = child.String()
	if p.Error() != nil {
		t.Error()
	}
	if child.Error() == nil {
		t.Error()
	}
}
