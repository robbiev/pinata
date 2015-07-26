package pinata

import "testing"

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
