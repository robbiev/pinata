// Package pinata is a utility to beat data out of interface{}, []interface{}
// and map[string]interface{}.
package pinata

import (
	"bytes"
	"fmt"
	"strings"
)

// Stick offers methods for extracting data from a Pinata.
type Stick interface {
	Error() error
	ClearError()
	StringAtPath(Pinata, ...string) string
	String(Pinata) string
	StringAtIndex(Pinata, int) string
	PinataAtPath(Pinata, ...string) Pinata
	PinataAtIndex(Pinata, int) Pinata
}

type ErrorContext struct {
	method      string
	methodInput []interface{}
}

// Method returns the name of the method that caused the error.
func (ec ErrorContext) Method() string {
	return ec.method
}

// MethodInput returns the input parameters of the method that caused the error.
func (ec ErrorContext) MethodInput() []interface{} {
	return ec.methodInput
}

type Pinata struct {
	context  *ErrorContext
	contents contents
}

func (p Pinata) Value() interface{} {
	return p.contents.Value()
}

func (p Pinata) Map() (map[string]interface{}, bool) {
	return p.contents.Map()
}

func (p Pinata) Slice() ([]interface{}, bool) {
	return p.contents.Slice()
}

type contents interface {
	Value() interface{}
	Map() (map[string]interface{}, bool)
	Slice() ([]interface{}, bool)
}

type otherPinata struct {
	value interface{}
}

func (p otherPinata) Value() interface{} {
	return p.value
}

func (p otherPinata) Map() (map[string]interface{}, bool) {
	return nil, false
}

func (p otherPinata) Slice() ([]interface{}, bool) {
	return nil, false
}

var _ = contents(mapPinata{})

type mapPinata struct {
	otherPinata
	value map[string]interface{}
}

func (p mapPinata) Map() (map[string]interface{}, bool) {
	return p.value, true
}

type slicePinata struct {
	otherPinata
	value []interface{}
}

func (p slicePinata) Slice() ([]interface{}, bool) {
	return p.value, true
}

// New is a starting point for a pinata celebration.
func New(contents interface{}) (Stick, Pinata) {
	return NewStick(), NewPinata(contents)
}

func NewStick() Stick {
	return &stick{}
}

// New creates a new Stick. Instances returned are not thread safe.
func NewPinata(contents interface{}) Pinata {
	switch t := contents.(type) {
	case map[string]interface{}:
		return Pinata{contents: &mapPinata{value: t}}
	case []interface{}:
		return Pinata{contents: &slicePinata{value: t}}
	default:
		return Pinata{contents: &otherPinata{value: t}}
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
	reason ErrorReason
	method string
	input  []interface{}
	advice string
}

// Reason indicates why the error occurred.
func (p PinataError) Reason() ErrorReason {
	return p.reason
}

// Method returns the name of the method that caused the error.
func (p PinataError) Method() string {
	return p.method
}

// MethodInput returns the input parameters of the method that caused the error.
func (p PinataError) MethodInput() []interface{} {
	return p.input
}

// Advice contains a human readable hint detailing how to remedy this error.
func (p PinataError) Advice() string {
	return p.advice
}

func (p PinataError) Error() string {
	var methodInput = p.MethodInput()
	var input string
	if len(methodInput) > 0 {
		var buf bytes.Buffer
		for i := range methodInput {
			_, _ = buf.WriteString("%#v")
			if i < len(methodInput)-1 {
				_, _ = buf.WriteString(", ")
			}
		}
		input = fmt.Sprintf(buf.String(), methodInput...)
	}
	return fmt.Sprintf("pinata: %s(%s) - %s (%s)", p.Method(), input, p.Reason(), p.Advice())
}

type stick struct {
	err error
}

func (s *stick) ClearError() {
	s.err = nil
}

func (s *stick) Error() error {
	return s.err
}

// this method assumes s.err != nil
func (s *stick) stringUnsupported() string {
	s.err = &PinataError{
		method: "String",
		reason: ErrorReasonIncompatibleType,
		input:  nil,
		advice: "call this method on a string pinata",
	}
	return ""
}

// this method assumes s.err != nil
func (s *stick) indexUnsupported(method string, index int) {
	s.err = &PinataError{
		method: method,
		reason: ErrorReasonIncompatibleType,
		input:  []interface{}{index},
		advice: "call this method on a slice pinata",
	}
}

// this method assumes s.err != nil
func (s *stick) setIndexOutOfRange(method string, index int, contents []interface{}) bool {
	if index < 0 || index >= len(contents) {
		s.err = &PinataError{
			method: method,
			reason: ErrorReasonInvalidInput,
			input:  []interface{}{index},
			advice: fmt.Sprintf("specify an index from 0 to %d", len(contents)-1),
		}
		return true
	}
	return false
}

// this method assumes s.err != nil
func (s *stick) pathUnsupported(method string, path []string) {
	s.err = &PinataError{
		method: method,
		reason: ErrorReasonIncompatibleType,
		input:  toInterfaceSlice(path),
		advice: "call this method on a map pinata",
	}
}

func (s *stick) String(p Pinata) string {
	if s.err != nil {
		return ""
	}
	if _, ok := p.Map(); ok {
		return s.stringUnsupported()
	}
	if _, ok := p.Slice(); ok {
		return s.stringUnsupported()
	}
	if v, ok := p.Value().(string); ok {
		return v
	}
	return s.stringUnsupported()
}

// this method assumes s.err != nil
func (s *stick) pinataAtIndex(p Pinata, method string, index int) Pinata {
	if slice, ok := p.Slice(); ok {
		if s.setIndexOutOfRange(method, index, slice) {
			return Pinata{}
		}
		return NewPinata(slice[index])
	}
	s.indexUnsupported("pinataAtIndex", index)
	return Pinata{}
}

func (s *stick) PinataAtIndex(p Pinata, index int) Pinata {
	if s.err != nil {
		return Pinata{}
	}
	return s.pinataAtIndex(p, "PinataAtIndex", index)
}

func (s *stick) StringAtIndex(p Pinata, index int) string {
	if s.err != nil {
		return ""
	}
	const method = "StringAtIndex"
	pinata := s.pinataAtIndex(p, method, index)
	if s.err != nil {
		return ""
	}
	str := s.String(pinata)
	if s.err != nil {
		s.err = &PinataError{
			method: method,
			reason: ErrorReasonIncompatibleType,
			input:  []interface{}{index},
			advice: "not a string, try another type",
		}
		return ""
	}
	return str
}

// this method assumes s.err != nil
func (s *stick) pinataAtPath(p Pinata, method string, path ...string) Pinata {
	contents, ok := p.Map()

	if !ok {
		s.pathUnsupported(method, path)
		return Pinata{}
	}

	if len(path) == 0 {
		s.err = &PinataError{
			method: method,
			reason: ErrorReasonInvalidInput,
			input:  toInterfaceSlice(path),
			advice: "specify a path",
		}
		return Pinata{}
	}

	for i := 0; i < len(path)-1; i++ {
		current := path[i]
		if v, ok := contents[current]; ok {
			if v, ok := v.(map[string]interface{}); ok {
				contents = v
			} else {
				s.err = &PinataError{
					method: method,
					reason: ErrorReasonIncompatibleType,
					input:  toInterfaceSlice(path),
					advice: fmt.Sprintf(`"%s" does not hold a pinata`, strings.Join(path[:i+1], `", "`)),
				}
				return Pinata{}
			}
		} else {
			s.err = &PinataError{
				method: method,
				reason: ErrorReasonNotFound,
				input:  toInterfaceSlice(path),
				advice: fmt.Sprintf(`"%s" does not exist`, strings.Join(path[:i+1], `", "`)),
			}
			return Pinata{}
		}
	}

	if v, ok := contents[path[len(path)-1]]; ok {
		return NewPinata(v)
	}

	s.err = &PinataError{
		method: method,
		reason: ErrorReasonNotFound,
		input:  toInterfaceSlice(path),
		advice: fmt.Sprintf(`"%s" does not exist`, strings.Join(path, `", "`)),
	}
	return Pinata{}
}

func (s *stick) PinataAtPath(p Pinata, path ...string) Pinata {
	if s.err != nil {
		return Pinata{}
	}
	return s.pinataAtPath(p, "PinataAtPath", path...)
}

func (s *stick) StringAtPath(p Pinata, path ...string) string {
	if s.err != nil {
		return ""
	}
	const method = "StringAtPath"
	pinata := s.pinataAtPath(p, method, path...)
	if s.err != nil {
		return ""
	}
	str := s.String(pinata)
	if s.err != nil {
		s.err = &PinataError{
			method: method,
			reason: ErrorReasonIncompatibleType,
			input:  toInterfaceSlice(path),
			advice: "not a string, try another type",
		}
	}
	return str
}

func toInterfaceSlice(c []string) []interface{} {
	ifaces := make([]interface{}, len(c))
	for i := range c {
		ifaces[i] = c[i]
	}
	return ifaces
}
