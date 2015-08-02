// Package pinata is a utility to beat data out of interface{}, []interface{}
// and map[string]interface{}. It was origally designed for use with the
// encoding/json package but can be generally useful.
//
// Unlike other packages most methods do not return an error type. They become
// a no-op when the first error is found so the error can be checked after a
// series of operations instead of at each operation seperately. Because of
// this "late" error handling design special care is taken to return good
// errors so you can still find out where things went wrong.
//
// Here's an example:
// https://godoc.org/github.com/robbiev/pinata#example-Stick
//
// This API is not thread safe.
package pinata

import (
	"bytes"
	"fmt"
	"strings"
)

// Stick offers methods of hitting the Pinata and extracting its goodness.
type Stick interface {
	// Error returns the first error encountered or nil if all operations so far
	// were successful.
	Error() error

	// ClearError clears the error and returns it. If there is no error the
	// method has no effect and returns nil, otherwise it returns the error that
	// was cleared.
	ClearError() error

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

	// PathFloat64 gets the float64 value at the given path within the Pinata.
	// The last element in the path must be a float64, the rest must be a
	// map[string]interface{}. The input Pinata must hold a
	// map[string]interface{} as well.
	PathFloat64(Pinata, ...string) float64

	// Float64 returns the Pinata as a float64 if it is one.
	Float64(Pinata) float64

	// IndexFloat64 gets the string float64 at the given index within the Pinata.
	// The input Pinata must hold a []interface{}.
	IndexFloat64(Pinata, int) float64

	// PathBool gets the bool value at the given path within the Pinata.
	// The last element in the path must be a bool, the rest must be a
	// map[string]interface{}. The input Pinata must hold a
	// map[string]interface{} as well.
	PathBool(Pinata, ...string) bool

	// Bool returns the Pinata as a bool if it is one.
	Bool(Pinata) bool

	// IndexBool gets the string bool at the given index within the Pinata.
	// The input Pinata must hold a []interface{}.
	IndexBool(Pinata, int) bool

	// PathNil asserts nil value at the given path within the Pinata. The last
	// element in the path must be a nil, the rest must be a
	// map[string]interface{}. The input Pinata must hold a
	// map[string]interface{} as well.
	PathNil(Pinata, ...string)

	// Nil asserts the Pinata holds a nil value.
	Nil(Pinata)

	// IndexNil asserts a nil value at the given index within the Pinata. The
	// input Pinata must hold a []interface{}.
	IndexNil(Pinata, int)

	// Path gets the Pinata value at the given path within the Pinata. All
	// elements in the path must be of type map[string]interface{}. The input
	// Pinata must hold a map[string]interface{} as well.
	Path(Pinata, ...string) Pinata

	// Index gets the Pinata value at the given index within the Pinata.
	// The input Pinata must hold a []interface{}.
	Index(Pinata, int) Pinata
}

type stick struct {
	err error
}

func (s *stick) ClearError() error {
	err := s.err
	s.err = nil
	return err
}

func (s *stick) Error() error {
	return s.err
}

// this method assumes s.err != nil
func (s *stick) unsupported(errCtx *ErrorContext, methodName string, input func() []interface{}, advice string) {
	s.err = &Error{
		context: &ErrorContext{
			methodName: methodName,
			methodArgs: input,
			next:       errCtx,
		},
		reason: ErrorReasonIncompatibleType,
		advice: advice,
	}
}

// this method assumes s.err != nil
func (s *stick) indexUnsupported(errCtx *ErrorContext, methodName string, index int) {
	s.err = &Error{
		context: &ErrorContext{
			methodName: methodName,
			methodArgs: func() []interface{} { return []interface{}{index} },
			next:       errCtx,
		},
		reason: ErrorReasonIncompatibleType,
		advice: "call this method on a slice pinata",
	}
}

// this method assumes s.err != nil
func (s *stick) pathUnsupported(errCtx *ErrorContext, methodName string, path []string) {
	s.err = &Error{
		context: &ErrorContext{
			methodName: methodName,
			methodArgs: func() []interface{} { return toInterfaceSlice(path) },
			next:       errCtx,
		},
		reason: ErrorReasonIncompatibleType,
		advice: "call this method on a map pinata",
	}
}

// this method assumes s.err != nil
func (s *stick) internalString(p Pinata, methodName string, input func() []interface{}) string {
	if _, ok := p.Map(); ok {
		s.unsupported(p.context, methodName, input, "this is a map")
		return ""
	}
	if _, ok := p.Slice(); ok {
		s.unsupported(p.context, methodName, input, "this is a slice")
		return ""
	}
	if v, ok := p.Value().(string); ok {
		return v
	}
	s.unsupported(p.context, methodName, input, "this is not a string")
	return ""
}

// this method assumes s.err != nil
func (s *stick) internalFloat64(p Pinata, methodName string, input func() []interface{}) float64 {
	if _, ok := p.Map(); ok {
		s.unsupported(p.context, methodName, input, "this is a map")
		return 0
	}
	if _, ok := p.Slice(); ok {
		s.unsupported(p.context, methodName, input, "this is a slice")
		return 0
	}
	if v, ok := p.Value().(float64); ok {
		return v
	}
	s.unsupported(p.context, methodName, input, "this is not a float64")
	return 0
}

// this method assumes s.err != nil
func (s *stick) internalBool(p Pinata, methodName string, input func() []interface{}) bool {
	if _, ok := p.Map(); ok {
		s.unsupported(p.context, methodName, input, "this is a map")
		return false
	}
	if _, ok := p.Slice(); ok {
		s.unsupported(p.context, methodName, input, "this is a slice")
		return false
	}
	if v, ok := p.Value().(bool); ok {
		return v
	}
	s.unsupported(p.context, methodName, input, "this is not a bool")
	return false
}

// this method assumes s.err != nil
func (s *stick) internalNil(p Pinata, methodName string, input func() []interface{}) {
	if p.Value() == nil {
		return
	}
	if _, ok := p.Map(); ok {
		s.unsupported(p.context, methodName, input, "this is a map")
	}
	if _, ok := p.Slice(); ok {
		s.unsupported(p.context, methodName, input, "this is a slice")
	}
	s.unsupported(p.context, methodName, input, "this is not nil")
}

func (s *stick) String(p Pinata) string {
	if s.err != nil {
		return ""
	}
	return s.internalString(p, "String", func() []interface{} { return nil })
}

func (s *stick) Bool(p Pinata) bool {
	if s.err != nil {
		return false
	}
	return s.internalBool(p, "Bool", func() []interface{} { return nil })
}

func (s *stick) Float64(p Pinata) float64 {
	if s.err != nil {
		return 0
	}
	return s.internalFloat64(p, "Float64", func() []interface{} { return nil })
}

func (s *stick) Nil(p Pinata) {
	if s.err != nil {
		return
	}
	s.internalNil(p, "Nil", func() []interface{} { return nil })
}

// this method assumes s.err != nil
func (s *stick) internalIndex(p Pinata, methodName string, index int) Pinata {
	if slice, ok := p.Slice(); ok {
		if index < 0 || index >= len(slice) {
			s.err = &Error{
				context: &ErrorContext{
					methodName: methodName,
					methodArgs: func() []interface{} { return []interface{}{index} },
					next:       p.context,
				},
				reason: ErrorReasonInvalidInput,
				advice: fmt.Sprintf("specify an index from 0 to %d", len(slice)-1),
			}
			return Pinata{}
		}
		return newPinataWithContext(slice[index], &ErrorContext{
			methodName: methodName,
			methodArgs: func() []interface{} { return []interface{}{index} },
			next:       p.context,
		})
	}
	s.indexUnsupported(p.context, methodName, index)
	return Pinata{}
}

func (s *stick) Index(p Pinata, index int) Pinata {
	if s.err != nil {
		return Pinata{}
	}
	return s.internalIndex(p, "Index", index)
}

func (s *stick) IndexString(p Pinata, index int) string {
	if s.err != nil {
		return ""
	}
	const methodName = "IndexString"
	pinata := s.internalIndex(p, methodName, index)
	if s.err != nil {
		return ""
	}
	pinata.context = p.context
	return s.internalString(pinata, methodName, func() []interface{} { return []interface{}{index} })
}

func (s *stick) IndexFloat64(p Pinata, index int) float64 {
	if s.err != nil {
		return 0
	}
	const methodName = "IndexFloat64"
	pinata := s.internalIndex(p, methodName, index)
	if s.err != nil {
		return 0
	}
	pinata.context = p.context
	return s.internalFloat64(pinata, methodName, func() []interface{} { return []interface{}{index} })
}

func (s *stick) IndexBool(p Pinata, index int) bool {
	if s.err != nil {
		return false
	}
	const methodName = "IndexBool"
	pinata := s.internalIndex(p, methodName, index)
	if s.err != nil {
		return false
	}
	pinata.context = p.context
	return s.internalBool(pinata, methodName, func() []interface{} { return []interface{}{index} })
}

func (s *stick) IndexNil(p Pinata, index int) {
	if s.err != nil {
		return
	}
	const methodName = "IndexNil"
	pinata := s.internalIndex(p, methodName, index)
	if s.err != nil {
		return
	}
	pinata.context = p.context
	s.internalNil(pinata, methodName, func() []interface{} { return []interface{}{index} })
}

// this method assumes s.err != nil
func (s *stick) internalPath(p Pinata, methodName string, path ...string) Pinata {
	contents, ok := p.Map()

	if !ok {
		s.pathUnsupported(p.context, methodName, path)
		return Pinata{}
	}

	if len(path) == 0 {
		s.err = &Error{
			context: &ErrorContext{
				methodName: methodName,
				methodArgs: func() []interface{} { return toInterfaceSlice(path) },
				next:       p.context,
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
				s.err = &Error{
					context: &ErrorContext{
						methodName: methodName,
						methodArgs: func() []interface{} { return toInterfaceSlice(path) },
						next:       p.context,
					},
					reason: ErrorReasonIncompatibleType,
					advice: fmt.Sprintf(`"%s" does not hold a pinata`, strings.Join(path[:i+1], `", "`)),
				}
				return Pinata{}
			}
		} else {
			s.err = &Error{
				context: &ErrorContext{
					methodName: methodName,
					methodArgs: func() []interface{} { return toInterfaceSlice(path) },
					next:       p.context,
				},
				reason: ErrorReasonNotFound,
				advice: fmt.Sprintf(`"%s" does not exist`, strings.Join(path[:i+1], `", "`)),
			}
			return Pinata{}
		}
	}

	if v, ok := contents[path[len(path)-1]]; ok {
		return newPinataWithContext(v, &ErrorContext{
			methodName: methodName,
			methodArgs: func() []interface{} { return toInterfaceSlice(path) },
			next:       p.context,
		})
	}

	s.err = &Error{
		context: &ErrorContext{
			methodName: methodName,
			methodArgs: func() []interface{} { return toInterfaceSlice(path) },
			next:       p.context,
		},
		reason: ErrorReasonNotFound,
		advice: fmt.Sprintf(`"%s" does not exist`, strings.Join(path, `", "`)),
	}
	return Pinata{}
}

func (s *stick) Path(p Pinata, path ...string) Pinata {
	if s.err != nil {
		return Pinata{}
	}
	return s.internalPath(p, "Path", path...)
}

func (s *stick) PathString(p Pinata, path ...string) string {
	if s.err != nil {
		return ""
	}
	const methodName = "PathString"
	pinata := s.internalPath(p, methodName, path...)
	if s.err != nil {
		return ""
	}
	pinata.context = p.context
	return s.internalString(pinata, methodName, func() []interface{} { return toInterfaceSlice(path) })
}

func (s *stick) PathFloat64(p Pinata, path ...string) float64 {
	if s.err != nil {
		return 0
	}
	const methodName = "PathFloat64"
	pinata := s.internalPath(p, methodName, path...)
	if s.err != nil {
		return 0
	}
	pinata.context = p.context
	return s.internalFloat64(pinata, methodName, func() []interface{} { return toInterfaceSlice(path) })
}

func (s *stick) PathBool(p Pinata, path ...string) bool {
	if s.err != nil {
		return false
	}
	const methodName = "PathBool"
	pinata := s.internalPath(p, methodName, path...)
	if s.err != nil {
		return false
	}
	pinata.context = p.context
	return s.internalBool(pinata, methodName, func() []interface{} { return toInterfaceSlice(path) })
}

func (s *stick) PathNil(p Pinata, path ...string) {
	if s.err != nil {
		return
	}
	const methodName = "PathNil"
	pinata := s.internalPath(p, methodName, path...)
	if s.err != nil {
		return
	}
	pinata.context = p.context
	s.internalNil(pinata, methodName, func() []interface{} { return toInterfaceSlice(path) })
}

// Pinata holds the data.
type Pinata struct {
	context   *ErrorContext
	value     interface{}
	mapFunc   func() (map[string]interface{}, bool)
	sliceFunc func() ([]interface{}, bool)
}

// Value returns the raw Pinata value.
func (p Pinata) Value() interface{} {
	return p.value
}

// Map returns the Pinata value as a map if it is one (the bool indicates
// success).
func (p Pinata) Map() (map[string]interface{}, bool) {
	if p.mapFunc != nil {
		return p.mapFunc()
	}
	return noMap()
}

// Slice returns the Pinata value as a slice if it is one (the bool indicates
// success).
func (p Pinata) Slice() ([]interface{}, bool) {
	if p.sliceFunc != nil {
		return p.sliceFunc()

	}
	return noSlice()
}

// New is a starting point for a pinata celebration.
func New(contents interface{}) (Stick, Pinata) {
	return NewStick(), NewPinata(contents)
}

// NewStick returns a new Stick to hit a Pinata with.
func NewStick() Stick {
	return &stick{}
}

// NewPinata creates a new Pinata holding the specified value.
func NewPinata(contents interface{}) Pinata {
	return newPinataWithContext(contents, nil)
}

func noMap() (map[string]interface{}, bool) { return nil, false }
func noSlice() ([]interface{}, bool)        { return nil, false }

func newPinataWithContext(contents interface{}, context *ErrorContext) Pinata {
	switch t := contents.(type) {
	case map[string]interface{}:
		return Pinata{
			value:     t,
			sliceFunc: noSlice,
			mapFunc: func() (map[string]interface{}, bool) {
				return t, true
			},
			context: context,
		}
	case []interface{}:
		return Pinata{
			value: t,
			sliceFunc: func() ([]interface{}, bool) {
				return t, true
			},
			mapFunc: noMap,
			context: context,
		}
	default:
		return Pinata{
			value:     t,
			sliceFunc: noSlice,
			mapFunc:   noMap,
			context:   context,
		}
	}
}

// ErrorReason describes the reason for returning an Error.
type ErrorReason string

const (
	// ErrorReasonIncompatibleType indicates the contents of the Pinata is not compatible with the invoked method.
	ErrorReasonIncompatibleType ErrorReason = "incompatible type"
	// ErrorReasonNotFound indicates the input has not been found in the Pinata.
	ErrorReasonNotFound = "not found"
	// ErrorReasonInvalidInput indicates the input is not in the expected range or format.
	ErrorReasonInvalidInput = "invalid input"
)

// ErrorContext contains information about the circumstances of an error.
type ErrorContext struct {
	methodName string
	methodArgs func() []interface{}
	next       *ErrorContext
}

// MethodName returns the name of the method that caused the error.
func (ec ErrorContext) MethodName() string {
	return ec.methodName
}

// MethodArgs returns the input parameters of the method that caused the error.
func (ec ErrorContext) MethodArgs() []interface{} {
	return ec.methodArgs()
}

// Next gets additional context linked to this one.
func (ec ErrorContext) Next() (ErrorContext, bool) {
	if ec.next != nil {
		return *ec.next, true
	}
	return ErrorContext{}, false
}

// Error is set on the Pinata when something goes wrong.
type Error struct {
	reason  ErrorReason
	context *ErrorContext
	advice  string
}

// Reason indicates why the error occurred.
func (p Error) Reason() ErrorReason {
	return p.reason
}

// Context returns more information about the circumstances of the error.
func (p Error) Context() (ErrorContext, bool) {
	if p.context != nil {
		return *p.context, true
	}
	return ErrorContext{}, false
}

// Advice contains a human readable hint detailing how to remedy this error.
func (p Error) Advice() string {
	return p.advice
}

// Error returns a summary of the problem.
func (p Error) Error() string {
	var summaries []string
	current := p.context
	for current != nil {
		var methodArgs = current.MethodArgs()
		var summary string
		if len(methodArgs) > 0 {
			var buf bytes.Buffer
			_, _ = buf.WriteString(current.MethodName())
			_ = buf.WriteByte('(')
			for i := range methodArgs {
				_, _ = buf.WriteString("%#v")
				if i < len(methodArgs)-1 {
					_, _ = buf.WriteString(", ")
				}
			}
			_ = buf.WriteByte(')')
			summary = fmt.Sprintf(buf.String(), methodArgs...)
			summaries = append(summaries, summary)
		} else {
			summaries = append(summaries, current.MethodName()+"()")
		}
		current = current.next
	}
	return fmt.Sprintf("pinata: %s (%s) at %v", p.Reason(), p.Advice(), strings.Join(summaries, " at "))
}

func toInterfaceSlice(c []string) []interface{} {
	ifaces := make([]interface{}, len(c))
	for i := range c {
		ifaces[i] = c[i]
	}
	return ifaces
}
