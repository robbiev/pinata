package pinata_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/robbiev/pinata"
)

const str = "test string"

func ExamplePinata() {
	const message = `{
		"Name": "Kevin",
		"Phone": ["+44 20 7123 4567", "+44 20 4567 7123"],
		"Address": {
			"Street": "1 Gopher Road",
			"City": "Gophertown"
		}
	}`

	var m map[string]interface{}

	err := json.Unmarshal([]byte(message), &m)
	if err != nil {
		fmt.Println(err)
		return
	}

	jsonPinata := pinata.New(m)

	type gopher struct {
		Name  string
		Phone string
		City  string
	}

	phonePinata := jsonPinata.PinataAtPath("Phone")
	if err := jsonPinata.Error(); err != nil {
		fmt.Println(err)
		return
	}

	kevin := gopher{
		Name:  jsonPinata.StringAtPath("Name"),
		Phone: phonePinata.StringAtIndex(0),
		City:  jsonPinata.StringAtPath("Address", "City"),
	}

	if err := phonePinata.Error(); err != nil {
		fmt.Println(err)
		return
	}

	if err := jsonPinata.Error(); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Name:", kevin.Name)
	fmt.Println("Phone:", kevin.Phone)
	fmt.Println("City:", kevin.City)

	// Output:
	// Name: Kevin
	// Phone: +44 20 7123 4567
	// City: Gophertown
}

func TestOtherPinataValidString(t *testing.T) {
	p := pinata.New(str)
	result := p.String()
	if p.Error() != nil {
		t.Error()
	}
	if result != str {
		t.Error()
	}
}

func TestOtherPinataInvalidString(t *testing.T) {
	p := pinata.New(1)
	_ = p.String()
	if p.Error() == nil {
		t.Error()
	}
}

func TestOtherPinataInvalidStringAtPath(t *testing.T) {
	p := pinata.New(str)
	_ = p.StringAtPath("a", "b", "c")
	if p.Error() == nil {
		t.Error()
	} else {
		fmt.Println(p.Error())
	}
}

func TestOtherPinataInvalidStringAtPath2(t *testing.T) {
	p := pinata.New(str)
	_ = p.StringAtPath("a")
	if p.Error() == nil {
		t.Error()
	} else {
		fmt.Println(p.Error())
	}
}

func TestMapPinata(t *testing.T) {
	m := make(map[string]interface{})
	m2 := make(map[string]interface{})
	m["one"] = m2
	m2["two"] = "three"
	p := pinata.New(m)
	if "three" != p.StringAtPath("one", "two") {
		fmt.Println(p.Error())
		t.Error()
	}
}

func TestMapPinataFailure(t *testing.T) {
	m := make(map[string]interface{})
	m2 := make(map[string]interface{})
	m["one"] = m2
	m2["two"] = "three"
	p := pinata.New(m)
	if p.PinataAtPath("one", "two", "three", "four") != nil {
		t.Error()
	}
	if p.Error() == nil {
		t.Error()
	}
}

func TestMapPinataDontPropagateErrorToParent(t *testing.T) {
	m := make(map[string]interface{})
	m2 := make(map[string]interface{})
	m["one"] = m2
	m2["two"] = "three"
	p := pinata.New(m)
	child := p.PinataAtPath("one")
	_ = child.String()
	if p.Error() != nil {
		t.Error()
	}
	if child.Error() == nil {
		t.Error()
	}
}

func TestSomething(t *testing.T) {
	m := make(map[string]interface{})
	m2 := make(map[string]interface{})
	m["one"] = m2
	m2["two"] = "three"
	p := pinata.New(m)
	_ = p.PinataAtPath("one", "three")
	if p.Error() == nil {
		t.Error()
	} else {
		fmt.Println(p.Error())
	}
	p.ClearError()
	_ = p.StringAtPath("one", "three")
	if p.Error() == nil {
		t.Error()
	} else {
		fmt.Println(p.Error())
	}
	p.ClearError()
	_ = p.PinataAtPath("one", "two", "three")
	if p.Error() == nil {
		t.Error()
	} else {
		fmt.Println(p.Error())
	}
	p.ClearError()
	_ = p.PinataAtPath("foo", "bar", "hello")
	if p.Error() == nil {
		t.Error()
	} else {
		fmt.Println(p.Error())
	}
}
