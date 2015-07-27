// Package pinata is a utility to beat data out of interface{}, []interface{}
// and map[string]interface{}.
package pinata

import (
	"fmt"
	"strings"
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
	p.err = fmt.Errorf("String(): not a string")
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

func (p *basePinata) indexErrorf(method string, index int, msg string) error {
	return fmt.Errorf("%s(%d): %s", method, index, msg)
}

func (p *basePinata) indexUnsupported(method string, index int) {
	if p.err != nil {
		return
	}
	p.err = p.indexErrorf(method, index, "not a slice so can't access by index")
}

func (p *basePinata) setIndexOutOfRange(method string, index int, contents []interface{}) bool {
	if index < 0 || index >= len(contents) {
		p.err = p.indexErrorf(method, index, fmt.Sprintf("index out of range: %d", index))
		return true
	}
	return false
}

func (p *basePinata) pathErrorf(method, pathStart string, path []string, msg string) error {
	return fmt.Errorf(`%s("%s"): %s`, method, strings.Join(toSlice(pathStart, path), `", "`), msg)
}

func (p *basePinata) pathUnsupported(method, pathStart string, path []string) {
	if p.err != nil {
		return
	}
	p.err = p.pathErrorf(method, pathStart, path, "not a map so can't access by path")
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
	p.err = fmt.Errorf("String(): not a string")
	return ""
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
		p.err = p.indexErrorf(method, index, "not a string")
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
		currentPinata := New(v)
		rest := path
		for len(rest) > 0 {
			tmp := currentPinata.PinataAtPath(rest[0])
			rest = rest[1:len(rest)]
			if currentPinata.Error() != nil {
				sofar := path[:len(path)-len(rest)]
				msg := fmt.Sprintf(`path ("%s", "%s") not found`, pathStart, strings.Join(sofar, `", "`))
				p.err = p.pathErrorf(method, pathStart, path, msg)
				return nil
			}
			currentPinata = tmp
		}
		return currentPinata
	}
	p.err = p.pathErrorf(method, pathStart, path, fmt.Sprintf(`path ("%s") not found`, pathStart))
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
		p.err = p.pathErrorf(method, pathStart, path, "not a string")
	}
	return s
}

func (p *mapPinata) Contents() interface{} {
	return p.contents
}

func toSlice(first string, rest []string) []string {
	slice := make([]string, len(rest)+1)
	i := 0
	slice[i] = first
	for _, v := range rest {
		i++
		slice[i] = v
	}
	return slice
}
