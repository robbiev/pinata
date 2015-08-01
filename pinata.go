// Package pinata is a utility to beat data out of interface{}, []interface{}
// and map[string]interface{}. It was origally designed for use with the
// encoding/json package but can be generally useful.
//
// Unlike other packages most methods do not return an error type. They become
// a no-op when the first error is found so the error can be checked after a
// series of operations instead of at each operation seperately. Because of
// this "late" error handling design special care is taken to return good
// errors so you can still find out where things went wrong.
package pinata

import (
	"bytes"
	"fmt"
	"strings"
)

// Stick offers methods for extracting data from a Pinata.
type Stick interface {
	// Error returns the first error encountered or nil if all operations so far
	// were successful.
	Error() error

	// ClearError clears the error. If there is no error the method has no effect.
	ClearError()

	// PathString gets the string value at the given path within the Pinata. The
	// last element in the path must be a string, the rest must be a
	// map[string]interface{}. The input Pinata must hold a
	// map[string]interface{} as well.
	PathString(Pinata, ...string) string

	// String returns the Pinata as a string if it is one.
	String(Pinata) string

	// IndexString gets the string value at the given index within the Pinata.
	// The input Pinata must hold a []interface{}.
	IndexString(Pinata, int) string

	// PathPinata gets the Pinata value at the given path within the Pinata. All
	// elements in the path must be of type map[string]interface{}. The input
	// Pinata must hold a map[string]interface{} as well.
	PathPinata(Pinata, ...string) Pinata

	// IndexString gets the Pinata value at the given index within the Pinata.
	// The input Pinata must hold a []interface{}.
	IndexPinata(Pinata, int) Pinata
}

type ErrorContext struct {
	method      string
	methodInput []interface{}
	next        *ErrorContext
}

// Method returns the name of the method that caused the error.
func (ec ErrorContext) Method() string {
	return ec.method
}

// MethodInput returns the input parameters of the method that caused the error.
func (ec ErrorContext) MethodInput() []interface{} {
	return ec.methodInput
}

func (ec ErrorContext) Next() *ErrorContext {
	return ec.next
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
	return newPinataWithContext(contents, nil)
}

func newPinataWithContext(contents interface{}, context *ErrorContext) Pinata {
	switch t := contents.(type) {
	case map[string]interface{}:
		return Pinata{contents: &mapPinata{value: t}, context: context}
	case []interface{}:
		return Pinata{contents: &slicePinata{value: t}, context: context}
	default:
		return Pinata{contents: &otherPinata{value: t}, context: context}
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
	reason  ErrorReason
	context *ErrorContext
	advice  string
}

// Reason indicates why the error occurred.
func (p PinataError) Reason() ErrorReason {
	return p.reason
}

// Context returns more information about the circumstances of the error.
func (p PinataError) Context() *ErrorContext {
	return p.context
}

// Advice contains a human readable hint detailing how to remedy this error.
func (p PinataError) Advice() string {
	return p.advice
}

func (p PinataError) Error() string {
	var summaries []string
	current := p.context
	for current != nil {
		var methodInput = current.MethodInput()
		var summary string
		if len(methodInput) > 0 {
			var buf bytes.Buffer
			_, _ = buf.WriteString(current.Method())
			_ = buf.WriteByte('(')
			for i := range methodInput {
				_, _ = buf.WriteString("%#v")
				if i < len(methodInput)-1 {
					_, _ = buf.WriteString(", ")
				}
			}
			_ = buf.WriteByte(')')
			summary = fmt.Sprintf(buf.String(), methodInput...)
			summaries = append(summaries, summary)
		}
		current = current.next
	}
	return fmt.Sprintf("pinata: %s (%s): \n\t%v", p.Reason(), p.Advice(), strings.Join(summaries, " :: "))
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
func (s *stick) stringUnsupported(errCtx *ErrorContext, method string, input []interface{}, advice string) string {
	s.err = &PinataError{
		context: &ErrorContext{
			method:      method,
			methodInput: input,
			next:        errCtx,
		},
		reason: ErrorReasonIncompatibleType,
		advice: advice,
	}
	return ""
}

// this method assumes s.err != nil
func (s *stick) indexUnsupported(errCtx *ErrorContext, method string, index int) {
	s.err = &PinataError{
		context: &ErrorContext{
			method:      method,
			methodInput: []interface{}{index},
			next:        errCtx,
		},
		reason: ErrorReasonIncompatibleType,
		advice: "call this method on a slice pinata",
	}
}

// this method assumes s.err != nil
func (s *stick) pathUnsupported(errCtx *ErrorContext, method string, path []string) {
	s.err = &PinataError{
		context: &ErrorContext{
			method:      method,
			methodInput: toInterfaceSlice(path),
			next:        errCtx,
		},
		reason: ErrorReasonIncompatibleType,
		advice: "call this method on a map pinata",
	}
}

// this method assumes s.err != nil
func (s *stick) internalString(p Pinata, method string, input []interface{}) string {
	if _, ok := p.Map(); ok {
		return s.stringUnsupported(p.context, method, input, "this is a map")
	}
	if _, ok := p.Slice(); ok {
		return s.stringUnsupported(p.context, method, input, "this is a slice")
	}
	if v, ok := p.Value().(string); ok {
		return v
	}
	return s.stringUnsupported(p.context, method, input, "this is not a string")
}

func (s *stick) String(p Pinata) string {
	if s.err != nil {
		return ""
	}
	return s.internalString(p, "String", nil)
}

// this method assumes s.err != nil
func (s *stick) internalIndexPinata(p Pinata, method string, index int) Pinata {
	if slice, ok := p.Slice(); ok {
		if index < 0 || index >= len(slice) {
			s.err = &PinataError{
				context: &ErrorContext{
					method:      method,
					methodInput: []interface{}{index},
					next:        p.context,
				},
				reason: ErrorReasonInvalidInput,
				advice: fmt.Sprintf("specify an index from 0 to %d", len(slice)-1),
			}
			return Pinata{}
		}
		return newPinataWithContext(slice[index], &ErrorContext{
			method:      method,
			methodInput: []interface{}{index},
			next:        p.context,
		})
	}
	s.indexUnsupported(p.context, method, index)
	return Pinata{}
}

func (s *stick) IndexPinata(p Pinata, index int) Pinata {
	if s.err != nil {
		return Pinata{}
	}
	return s.internalIndexPinata(p, "IndexPinata", index)
}

func (s *stick) IndexString(p Pinata, index int) string {
	if s.err != nil {
		return ""
	}
	const method = "IndexString"
	pinata := s.internalIndexPinata(p, method, index)
	if s.err != nil {
		return ""
	}
	pinata.context = p.context
	return s.internalString(pinata, method, []interface{}{index})
}

// this method assumes s.err != nil
func (s *stick) internalPathPinata(p Pinata, method string, path ...string) Pinata {
	contents, ok := p.Map()

	if !ok {
		s.pathUnsupported(p.context, method, path)
		return Pinata{}
	}

	if len(path) == 0 {
		s.err = &PinataError{
			context: &ErrorContext{
				method:      method,
				methodInput: toInterfaceSlice(path),
				next:        p.context,
			},
			reason: ErrorReasonInvalidInput,
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
					context: &ErrorContext{
						method:      method,
						methodInput: toInterfaceSlice(path),
						next:        p.context,
					},
					reason: ErrorReasonIncompatibleType,
					advice: fmt.Sprintf(`"%s" does not hold a pinata`, strings.Join(path[:i+1], `", "`)),
				}
				return Pinata{}
			}
		} else {
			s.err = &PinataError{
				context: &ErrorContext{
					method:      method,
					methodInput: toInterfaceSlice(path),
					next:        p.context,
				},
				reason: ErrorReasonNotFound,
				advice: fmt.Sprintf(`"%s" does not exist`, strings.Join(path[:i+1], `", "`)),
			}
			return Pinata{}
		}
	}

	if v, ok := contents[path[len(path)-1]]; ok {
		return newPinataWithContext(v, &ErrorContext{
			method:      method,
			methodInput: toInterfaceSlice(path),
			next:        p.context,
		})
	}

	s.err = &PinataError{
		context: &ErrorContext{
			method:      method,
			methodInput: toInterfaceSlice(path),
			next:        p.context,
		},
		reason: ErrorReasonNotFound,
		advice: fmt.Sprintf(`"%s" does not exist`, strings.Join(path, `", "`)),
	}
	return Pinata{}
}

func (s *stick) PathPinata(p Pinata, path ...string) Pinata {
	if s.err != nil {
		return Pinata{}
	}
	return s.internalPathPinata(p, "PathPinata", path...)
}

func (s *stick) PathString(p Pinata, path ...string) string {
	if s.err != nil {
		return ""
	}
	const method = "PathString"
	pinata := s.internalPathPinata(p, method, path...)
	if s.err != nil {
		return ""
	}
	pinata.context = p.context
	return s.internalString(pinata, method, toInterfaceSlice(path))
}

func toInterfaceSlice(c []string) []interface{} {
	ifaces := make([]interface{}, len(c))
	for i := range c {
		ifaces[i] = c[i]
	}
	return ifaces
}
