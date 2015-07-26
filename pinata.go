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
		return &basePinata{}
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
	p.indexFail("PinataAtIndex", index)
	return nil
}

func (p *basePinata) PinataAtPath(pathStart string, path ...string) Pinata {
	p.pathFail("PinataAtPath", pathStart, path)
	return nil
}

func (p *basePinata) StringAtPath(pathStart string, path ...string) string {
	p.pathFail("StringAtPath", pathStart, path)
	return ""
}

func (p *basePinata) StringAtIndex(index int) string {
	p.indexFail("StringAtIndex", index)
	return ""
}

func (p *basePinata) Contents() interface{} {
	return nil
}

func (p *basePinata) indexFail(method string, index int) {
	if p.err != nil {
		return
	}
	p.err = fmt.Errorf("%s(%d): not a slice so can't access by index", method, index)
}

func (p *basePinata) pathFail(method, pathStart string, path []string) {
	if p.err != nil {
		return
	}
	p.err = fmt.Errorf(`%s("%s"): not a map so can't access by path`, method, strings.Join(toSlice(pathStart, path), `", "`))
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

func (p *slicePinata) PinataAtIndex(index int) Pinata {
	if p.err != nil {
		return nil
	}
	if index >= 0 && index < len(p.contents) {
		return New(p.contents[index])
	}
	p.indexFail("PinataAtIndex", index)
	return nil
}

func (p *slicePinata) StringAtIndex(index int) string {
	if p.err != nil {
		return ""
	}
	pinata := p.PinataAtIndex(index)
	if p.err != nil {
		return ""
	}
	s := pinata.String()
	if pinata.Error() != nil {
		p.StringAtIndex(index)
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

func (p *mapPinata) PinataAtPath(pathStart string, path ...string) Pinata {
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
				goto Fail
			}
			currentPinata = tmp
		}
		return currentPinata
	}
Fail:
	p.pathFail("PinataAtPath", pathStart, path)
	return nil
}

func (p *mapPinata) StringAtPath(pathStart string, path ...string) string {
	pinata := p.PinataAtPath(pathStart, path...)
	if p.err != nil {
		return ""
	}
	s := pinata.String()
	if pinata.Error() != nil {
		p.StringAtPath(pathStart, path...)
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
