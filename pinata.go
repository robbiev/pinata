package pinata

import (
	"fmt"
	"strings"
)

// TODO allow sub-pinatas, propagate the errors
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

type basePinata struct {
	err error
}

func (bp *basePinata) Error() error {
	return bp.err
}

func (bp *basePinata) ClearError() {
	bp.err = nil
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

type slicePinata struct {
	basePinata
	contents []interface{}
}

type mapPinata struct {
	basePinata
	contents map[string]interface{}
}

type otherPinata struct {
	basePinata
	contents interface{}
}

// TODO create constructor that takes a func returning the interface{} value
// and error for use with the JSON libs
func New(contents interface{}) Pinata {
	return &otherPinata{contents: contents}
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

func (p *otherPinata) PinataAtIndex(index int32) Pinata {
	p.indexFail("PinataAtIndex", index)
	return nil
}

func (p *otherPinata) PinataAtPath(pathStart string, path ...string) Pinata {
	p.pathFail("PinataAtPath", pathStart, path)
	return nil
}

func (p *otherPinata) StringAtPath(pathStart string, path ...string) string {
	p.pathFail("StringAtPath", pathStart, path)
	return ""
}

func (p *otherPinata) StringAtIndex(index int32) string {
	p.indexFail("StringAtIndex", index)
	return ""
}

func (p *otherPinata) Contents() interface{} {
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
