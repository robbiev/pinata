// Package pinata is a utility to beat data out of interface{}, []interface{}
// and map[string]interface{}.
package pinata

import (
	"fmt"
	"strings"
)

// Holder of a value with methods for extracting data from it.
type Pinata interface {
	Contents() interface{}
	Error() error
	ClearError()
	StringAtPath(string, ...string) string
	String() string
	StringAtIndex(int32) string
	PinataAtPath(string, ...string) Pinata
	PinataAtIndex(int32) Pinata
}

// Creates a new Pinata. Instances returned are not thread safe.
func New(contents interface{}) Pinata {
	// TODO create constructor that takes a func returning the interface{} value
	// and error for use with the JSON libs
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

func (bp *basePinata) Error() error {
	return bp.err
}

func (bp *basePinata) ClearError() {
	bp.err = nil
}

func (p *basePinata) String() string {
	if p.err != nil {
		return ""
	}
	p.err = fmt.Errorf("String(): not a string")
	return ""
}

func (p *basePinata) PinataAtIndex(index int32) Pinata {
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

func (p *basePinata) StringAtIndex(index int32) string {
	p.indexFail("StringAtIndex", index)
	return ""
}

func (fp *basePinata) Contents() interface{} {
	return nil
}

func (bp *basePinata) indexFail(method string, index int32) {
	if bp.err != nil {
		return
	}
	bp.err = fmt.Errorf("%s(%d): not a slice so can't access by index", method, index)
}

func (bp *basePinata) pathFail(method, pathStart string, path []string) {
	if bp.err != nil {
		return
	}
	bp.err = fmt.Errorf(`%s("%s"): not a map so can't access by path`, method, strings.Join(toSlice(pathStart, path), `", "`))
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
		if len(path) == 0 {
			return currentPinata
		} else {
			first, rest := path[len(path)-1], path[:len(path)-1]
			return currentPinata.PinataAtPath(first, rest...)
		}
	}
	p.pathFail("PinataAtPath", pathStart, path)
	return nil
}

func (p *mapPinata) StringAtPath(pathStart string, path ...string) string {
	pinata := p.PinataAtPath(pathStart, path...)
	if p.err != nil {
		return ""
	}
	return pinata.String()
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
