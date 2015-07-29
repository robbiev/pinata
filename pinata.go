// Package pinata is a utility to beat data out of interface{}, []interface{}
// and map[string]interface{}.
package pinata

import (
	"bytes"
	"fmt"
)

// Pinata holds a value and offers methods for extracting data from it.
type Pinata interface {
	Contents() interface{}
	Error() error
	ClearError()
	StringAtPath(string, ...string) string
	String() string
	StringAtIndex(int) string
	PinataAtPath(string, ...string) Pinata
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
	ErrorReasonUnknown         ErrorReason = "unknown"
	ErrorReasonIncompatbleType             = "incompatible type"
	ErrorReasonNotFound                    = "not found"
	ErrorReasonInvalidInput                = "invalid input"
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
	return fmt.Sprintf("pinata: %s(%s): %s - %s", p.Method, input, p.Reason, p.Advice)
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
		Reason: ErrorReasonIncompatbleType,
		Input:  nil,
		Advice: "call this method on a string pinata",
	}
	return ""
}

func (p *basePinata) PinataAtIndex(index int) Pinata {
	p.indexUnsupported("PinataAtIndex", index)
	return nil
}

func (p *basePinata) PinataAtPath(pathStart string, path ...string) Pinata {
	p.pathUnsupported("PinataAtPath", pathStart, path)
	return nil
}

func (p *basePinata) StringAtPath(pathStart string, path ...string) string {
	p.pathUnsupported("StringAtPath", pathStart, path)
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
		Reason: ErrorReasonIncompatbleType,
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

func (p *basePinata) pathUnsupported(method, pathStart string, path []string) {
	if p.err != nil {
		return
	}
	p.err = &PinataError{
		Method: method,
		Reason: ErrorReasonIncompatbleType,
		Input:  toSlice(pathStart, path),
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
			Reason: ErrorReasonIncompatbleType,
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

func (p *mapPinata) pinataAtPath(method, pathStart string, path ...string) Pinata {
	if p.err != nil {
		return nil
	}
	if v, ok := p.contents[pathStart]; ok {
		//if v, ok := v.(map[string]interface{})
		currentPinata := New(v)
		rest := path
		for len(rest) > 0 {
			tmp := currentPinata.PinataAtPath(rest[0])
			rest = rest[1:len(rest)]
			if currentPinata.Error() != nil {
				// TODO need to customise the message based on the returned error
				//sofar := path[:len(path)-len(rest)]
				p.err = &PinataError{
					Method: method,
					Reason: ErrorReasonNotFound,
					Input:  toSlice(pathStart, path),
					Advice: "can't find that, sorry",
				}
				return nil
			}
			currentPinata = tmp
		}
		return currentPinata
	}
	p.err = &PinataError{
		Method: method,
		Reason: ErrorReasonNotFound,
		Input:  toSlice(pathStart, path),
		Advice: fmt.Sprintf(`no "%s" in this pinata`, pathStart),
	}
	return nil
}

func (p *mapPinata) PinataAtPath(pathStart string, path ...string) Pinata {
	return p.pinataAtPath("PinataAtPath", pathStart, path...)
}

func (p *mapPinata) StringAtPath(pathStart string, path ...string) string {
	const method = "StringAtPath"
	pinata := p.pinataAtPath(method, pathStart, path...)
	if p.err != nil {
		return ""
	}
	s := pinata.String()
	if pinata.Error() != nil {
		p.err = &PinataError{
			Method: method,
			Reason: ErrorReasonIncompatbleType,
			Input:  toSlice(pathStart, path),
			Advice: "not a string, try another type",
		}
	}
	return s
}

func (p *mapPinata) Contents() interface{} {
	return p.contents
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
