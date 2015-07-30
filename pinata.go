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

type ErrorReason string

const (
	ErrorReasonIncompatibleType ErrorReason = "incompatible type"
	ErrorReasonNotFound                     = "not found"
	ErrorReasonInvalidInput                 = "invalid input"
)

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
	p.indexUnsupported("PinataAtIndex", index)
	return nil
}

func (p *basePinata) PinataAtPath(path ...string) Pinata {
	p.pathUnsupported("PinataAtPath", path)
	return nil
}

func (p *basePinata) StringAtPath(path ...string) string {
	p.pathUnsupported("StringAtPath", path)
	return ""
}

func (p *basePinata) StringAtIndex(index int) string {
	p.indexUnsupported("StringAtIndex", index)
	return ""
}

func (p *basePinata) Contents() interface{} {
	return nil
}

func (p *basePinata) indexUnsupported(method string, index int) {
	if p.err != nil {
		return
	}
	p.err = &PinataError{
		Method: method,
		Reason: ErrorReasonIncompatibleType,
		Input:  []interface{}{index},
		Advice: "call this method on a slice pinata",
	}
}

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

func (p *basePinata) pathUnsupported(method string, path []string) {
	if p.err != nil {
		return
	}
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

func (p *slicePinata) pinataAtIndex(method string, index int) Pinata {
	if p.err != nil {
		return nil
	}
	if p.setIndexOutOfRange(method, index, p.contents) {
		return nil
	}
	return New(p.contents[index])
}

func (p *slicePinata) PinataAtIndex(index int) Pinata {
	return p.pinataAtIndex("PinataAtIndex", index)
}

func (p *slicePinata) StringAtIndex(index int) string {
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

func (p *mapPinata) pinataAtPath(method string, path ...string) Pinata {
	if p.err != nil {
		return nil
	}
	if len(path) == 0 {
		p.err = &PinataError{
			Method: method,
			Reason: ErrorReasonInvalidInput,
			Input:  toInterfaceSlice(path),
			Advice: "specify a path",
		}
		return nil
	}

	value := p.contents
	for i := 0; i < len(path)-1; i++ {
		current := path[i]
		if v, ok := value[current]; ok {
			if v, ok := v.(map[string]interface{}); ok {
				value = v
			} else {
				p.err = &PinataError{
					Method: method,
					Reason: ErrorReasonIncompatibleType,
					Input:  toInterfaceSlice(path),
					Advice: fmt.Sprintf(`path "%s" does not hold a pinata`, strings.Join(path[:i+1], `", "`)),
				}
				return nil
			}
		} else {
			p.err = &PinataError{
				Method: method,
				Reason: ErrorReasonNotFound,
				Input:  toInterfaceSlice(path),
				Advice: fmt.Sprintf(`path "%s" does not exist`, strings.Join(path[:i+1], `", "`)),
			}
			return nil
		}
	}

	if v, ok := value[path[len(path)-1]]; ok {
		return New(v)
	}

	p.err = &PinataError{
		Method: method,
		Reason: ErrorReasonNotFound,
		Input:  toInterfaceSlice(path),
		Advice: fmt.Sprintf(`path "%s" does not exist`, strings.Join(path, `", "`)),
	}
	return nil
}

func (p *mapPinata) PinataAtPath(path ...string) Pinata {
	return p.pinataAtPath("PinataAtPath", path...)
}

func (p *mapPinata) StringAtPath(path ...string) string {
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

func toSlice(first string, rest []string) []interface{} {
	slice := make([]interface{}, len(rest)+1)
	i := 0
	slice[i] = first
	for _, v := range rest {
		i++
		slice[i] = v
	}
	return slice
}
