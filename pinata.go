// Package pinata is a utility to beat data out of interface{}, []interface{}
// and map[string]interface{}.
package pinata

import (
	"bytes"
	"fmt"
	"strings"
)

// Pinata holds a value and offers methods for extracting data from it.
type Pinata interface {
	Contents() interface{}
	Error() error
	ClearError()
	StringAtPath(...string) string
	String() string
	StringAtIndex(int) string
	PinataAtPath(...string) Pinata
	PinataAtIndex(int) Pinata
}

// New creates a new Pinata. Instances returned are not thread safe.
func New(contents interface{}) Pinata {
	switch t := contents.(type) {
	default:
		return &otherPinata{contents: t}
	case map[string]interface{}:
		return &mapPinata{contents: t}
	case []interface{}:
		return &slicePinata{}
	}
}

var _ = error(PinataError{})

// ErrorReason describes the reason for returning a PinataError.
type ErrorReason string

const (
	// ErrorReasonIncompatibleType indicates the contents of the Pinata is not compatible with the invoked method.
	ErrorReasonIncompatibleType ErrorReason = "incompatible type"
	// ErrorReasonNotFound indicates the input has not been found in the Pinata.
	ErrorReasonNotFound = "not found"
	// ErrorReasonInvalidInput indicates the input is not in the expected range or format.
	ErrorReasonInvalidInput = "invalid input"
)

// PinataError is set on the Pinata if something goes wrong.
type PinataError struct {
	Reason ErrorReason
	Method string
	Input  []interface{}
	Advice string
}

func (p PinataError) Error() string {
	var input string
	if len(p.Input) > 0 {
		var buf bytes.Buffer
		for i := range p.Input {
			buf.WriteString("%#v")
			if i < len(p.Input)-1 {
				buf.WriteString(", ")
			}
		}
		input = fmt.Sprintf(buf.String(), p.Input...)
	}
	return fmt.Sprintf("pinata: %s(%s) - %s (%s)", p.Method, input, p.Reason, p.Advice)
}

type basePinata struct {
	err error
}

func (p *basePinata) Error() error {
	return p.err
}

func (p *basePinata) ClearError() {
	p.err = nil
}

func (p *basePinata) String() string {
	if p.err != nil {
		return ""
	}
	p.err = &PinataError{
		Method: "String",
		Reason: ErrorReasonIncompatibleType,
		Input:  nil,
		Advice: "call this method on a string pinata",
	}
	return ""
}

func (p *basePinata) PinataAtIndex(index int) Pinata {
	if p.err != nil {
		return nil
	}
	p.indexUnsupported("PinataAtIndex", index)
	return nil
}

func (p *basePinata) PinataAtPath(path ...string) Pinata {
	if p.err != nil {
		return nil
	}
	p.pathUnsupported("PinataAtPath", path)
	return nil
}

func (p *basePinata) StringAtPath(path ...string) string {
	if p.err != nil {
		return ""
	}
	p.pathUnsupported("StringAtPath", path)
	return ""
}

func (p *basePinata) StringAtIndex(index int) string {
	if p.err != nil {
		return ""
	}
	p.indexUnsupported("StringAtIndex", index)
	return ""
}

func (p *basePinata) Contents() interface{} {
	return nil // should always override this method
}

// this method assumes p.err != nil
func (p *basePinata) indexUnsupported(method string, index int) {
	p.err = &PinataError{
		Method: method,
		Reason: ErrorReasonIncompatibleType,
		Input:  []interface{}{index},
		Advice: "call this method on a slice pinata",
	}
}

// this method assumes p.err != nil
func (p *basePinata) setIndexOutOfRange(method string, index int, contents []interface{}) bool {
	if index < 0 || index >= len(contents) {
		p.err = &PinataError{
			Method: method,
			Reason: ErrorReasonInvalidInput,
			Input:  []interface{}{index},
			Advice: fmt.Sprintf("specify an index from 0 to %d", len(contents)-1),
		}
		return true
	}
	return false
}

// this method assumes p.err != nil
func (p *basePinata) pathUnsupported(method string, path []string) {
	p.err = &PinataError{
		Method: method,
		Reason: ErrorReasonIncompatibleType,
		Input:  toInterfaceSlice(path),
		Advice: "call this method on a map pinata",
	}
}

type otherPinata struct {
	basePinata
	contents interface{}
}

func (p *otherPinata) String() string {
	if p.err != nil {
		return ""
	}
	if v, ok := p.contents.(string); ok {
		return v
	}
	return p.basePinata.String()
}

func (p *otherPinata) Contents() interface{} {
	return p.contents
}

type slicePinata struct {
	basePinata
	contents []interface{}
}

// this method assumes p.err != nil
func (p *slicePinata) pinataAtIndex(method string, index int) Pinata {
	if p.setIndexOutOfRange(method, index, p.contents) {
		return nil
	}
	return New(p.contents[index])
}

func (p *slicePinata) PinataAtIndex(index int) Pinata {
	if p.err != nil {
		return nil
	}
	return p.pinataAtIndex("PinataAtIndex", index)
}

func (p *slicePinata) StringAtIndex(index int) string {
	if p.err != nil {
		return ""
	}
	const method = "StringAtIndex"
	pinata := p.pinataAtIndex(method, index)
	if p.err != nil {
		return ""
	}
	s := pinata.String()
	if pinata.Error() != nil {
		p.err = &PinataError{
			Method: method,
			Reason: ErrorReasonIncompatibleType,
			Input:  []interface{}{index},
			Advice: "not a string, try another type",
		}
	}
	return s
}

func (p *slicePinata) Contents() interface{} {
	return p.contents
}

type mapPinata struct {
	basePinata
	contents map[string]interface{}
}

// this method assumes p.err != nil
func (p *mapPinata) pinataAtPath(method string, path ...string) Pinata {
	if len(path) == 0 {
		p.err = &PinataError{
			Method: method,
			Reason: ErrorReasonInvalidInput,
			Input:  toInterfaceSlice(path),
			Advice: "specify a path",
		}
		return nil
	}

	contents := p.contents
	for i := 0; i < len(path)-1; i++ {
		current := path[i]
		if v, ok := contents[current]; ok {
			if v, ok := v.(map[string]interface{}); ok {
				contents = v
			} else {
				p.err = &PinataError{
					Method: method,
					Reason: ErrorReasonIncompatibleType,
					Input:  toInterfaceSlice(path),
					Advice: fmt.Sprintf(`"%s" does not hold a pinata`, strings.Join(path[:i+1], `", "`)),
				}
				return nil
			}
		} else {
			p.err = &PinataError{
				Method: method,
				Reason: ErrorReasonNotFound,
				Input:  toInterfaceSlice(path),
				Advice: fmt.Sprintf(`"%s" does not exist`, strings.Join(path[:i+1], `", "`)),
			}
			return nil
		}
	}

	if v, ok := contents[path[len(path)-1]]; ok {
		return New(v)
	}

	p.err = &PinataError{
		Method: method,
		Reason: ErrorReasonNotFound,
		Input:  toInterfaceSlice(path),
		Advice: fmt.Sprintf(`"%s" does not exist`, strings.Join(path, `", "`)),
	}
	return nil
}

func (p *mapPinata) PinataAtPath(path ...string) Pinata {
	if p.err != nil {
		return nil
	}
	return p.pinataAtPath("PinataAtPath", path...)
}

func (p *mapPinata) StringAtPath(path ...string) string {
	if p.err != nil {
		return ""
	}
	const method = "StringAtPath"
	pinata := p.pinataAtPath(method, path...)
	if p.err != nil {
		return ""
	}
	s := pinata.String()
	if pinata.Error() != nil {
		p.err = &PinataError{
			Method: method,
			Reason: ErrorReasonIncompatibleType,
			Input:  toInterfaceSlice(path),
			Advice: "not a string, try another type",
		}
	}
	return s
}

func (p *mapPinata) Contents() interface{} {
	return p.contents
}

func toInterfaceSlice(c []string) []interface{} {
	ifaces := make([]interface{}, len(c))
	for i := range c {
		ifaces[i] = c[i]
	}
	return ifaces
}
