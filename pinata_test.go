package pinata

import (
	"fmt"
	"testing"
)

const str = "test string"

func TestValidString(t *testing.T) {
	p := New(str)
	result := p.String()
	if p.Error() != nil {
		t.Error()
	}
	if result != str {
		t.Error()
	}
}

func TestInvalidString(t *testing.T) {
	p := New(1)
	_ = p.String()
	if p.Error() == nil {
		t.Error()
	}
}

func TestInvalidStringAtPath(t *testing.T) {
	p := New(str)
	_ = p.StringAtPath("a", "b", "c")
	if p.Error() == nil {
		t.Error()
	} else {
		fmt.Println(p.Error())
	}
}

func TestInvalidStringAtPath2(t *testing.T) {
	p := New(str)
	_ = p.StringAtPath("a")
	if p.Error() == nil {
		t.Error()
	} else {
		fmt.Println(p.Error())
	}
}
