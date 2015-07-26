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
	StringPath(string, ...string) string
	String() string
	StringIndex(int32) string
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
	if p.basePinata.err != nil {
		return ""
	}
	if v, ok := p.contents.(string); ok {
		return v
	}
	p.basePinata.err = fmt.Errorf("String(): not a string")
	return ""
}

func (p *otherPinata) StringPath(pathStart string, path ...string) string {
	if p.basePinata.err != nil {
		return ""
	}
	p.basePinata.err = fmt.Errorf("StringPath(%s, %s): not a map so can't access by path", pathStart, strings.Join(path, ", "))
	return ""
}

func (p *otherPinata) StringIndex(index int32) string {
	if p.basePinata.err != nil {
		return ""
	}
	p.basePinata.err = fmt.Errorf("StringIndex(%d): not a slice so can't access by index", index)
	return ""
}

func (p *otherPinata) Contents() interface{} {
	return p.contents
}
