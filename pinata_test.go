package pinata_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/robbiev/pinata"
)

func start(t *testing.T) (pinata.Stick, pinata.Pinata) {
	const message = `
	{
		"Name": "Kevin",
		"Phone": ["+44 20 7123 4567", "+44 20 4567 7123"],
		"Address": {
			"Street": "1 Gopher Road",
			"City": null
		},
		"Hobbies": [
		  {
				"Indoors": ["napping", "watching TV", "jumping up and down"],
				"Outdoors": ["napping", "hiking", "petanque"]
			}
		]
	}`

	var m map[string]interface{}

	err := json.Unmarshal([]byte(message), &m)
	if err != nil {
		t.Log(err)
		return nil, pinata.Pinata{}
	}

	return pinata.New(m)
}

func ExampleStick() {
	const message = `
	{
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

	stick, p := pinata.New(m)

	type gopher struct {
		Name  string
		Phone string
		City  string
	}

	// no error handling here
	kevin := gopher{
		Name:  stick.PathString(p, "Name"),
		Phone: stick.IndexString(stick.Path(p, "Phone"), 0),
		City:  stick.PathString(p, "Address", "City"),
	}

	if err := stick.ClearError(); err != nil {
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

func TestFailureScenario(t *testing.T) {
	const message = `
	{
		"Name": "Kevin",
		"Phone": ["+44 20 7123 4567", "+44 20 4567 7123"],
		"Address": {
			"Street": "1 Gopher Road",
			"City": "Gophertown"
		}
	}`

	var m map[string]interface{}

	{
		err := json.Unmarshal([]byte(message), &m)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	stick, thePinata := pinata.New(m)

	stick.IndexFloat64(stick.Path(thePinata, "Phone"), 1)
	err := stick.ClearError()
	if err == nil {
		t.Error("phone must not be a float64")
		return
	}
	pinataErr := err.(*pinata.Error)
	if pinataErr.Reason() != pinata.ErrorReasonIncompatibleType {
		t.Error("error reason must be incompatible type")
	}

	ctx, ok := pinataErr.Context()
	if !ok {
		t.Error("error must have a context")
	}
	if ctx.MethodName() != "IndexFloat64" {
		t.Error("error method name must be IndexFloat64")
	}
	if len(ctx.MethodArgs()) != 1 {
		t.Error("error method input must have one element")
	}
	if ctx.MethodArgs()[0] != 1 {
		t.Error("error method input must be 1")
	}

	ctx2, ok2 := ctx.Next()
	if !ok2 {
		t.Error("error must have a linked context")
	}
	if ctx2.MethodName() != "Path" {
		t.Error("error method name must be Path")
	}
	if len(ctx2.MethodArgs()) != 1 {
		t.Error("error method input must have one element")
	}
	if ctx2.MethodArgs()[0] != "Phone" {
		t.Error("error method input must be 1")
	}

	_, ok3 := ctx2.Next()
	if ok3 {
		t.Error("linked context must have no further linked contexts")
	}
}

func TestIndex(t *testing.T) {
	stick, thePinata := start(t)
	{
		stick.Index(pinata.Pinata{}, 0)
		err := stick.ClearError()
		if err == nil {
			t.Error("empty pinata must result in an error")
		} else {
			t.Log(err)
		}
	}
	{
		stick.Index(stick.Path(thePinata, "Address"), 0)
		err := stick.ClearError()
		if err == nil {
			t.Error("accessing a non-slice pinata by index must result in an error")
		} else {
			t.Log(err)
		}
	}
	{
		stick.Index(stick.Path(thePinata, "Phone"), -1)
		err := stick.ClearError()
		if err == nil {
			t.Error("negative index must result in an error")
		} else {
			t.Log(err)
		}
	}
	{
		stick.Index(stick.Path(thePinata, "Phone"), 1)
		if err := stick.ClearError(); err != nil {
			t.Error("Phone entry with index 1 must exist", err)
		}
	}
}

func TestPath(t *testing.T) {
	stick, thePinata := start(t)
	{
		stick.Path(pinata.Pinata{}, "nope")
		err := stick.ClearError()
		if err == nil {
			t.Error("empty pinata must result in an error")
		} else {
			t.Log(err)
		}
	}
	{
		stick.Path(thePinata)
		err := stick.ClearError()
		if err == nil {
			t.Error("empty path must result in an error")
		} else {
			t.Log(err)
		}
	}
	{
		stick.Path(thePinata, "nope")
		err := stick.ClearError()
		if err == nil {
			t.Error("non-existent path must result in an error")
		} else {
			t.Log(err)
		}
	}
	{
		stick.Path(thePinata, "nope", "nope")
		err := stick.ClearError()
		if err == nil {
			t.Error("non-existent path must result in an error")
		} else {
			t.Log(err)
		}
	}
	{
		stick.Path(thePinata, "Address", "nope", "nope")
		err := stick.ClearError()
		if err == nil {
			t.Error("non-existent path must result in an error")
		} else {
			t.Log(err)
		}
	}
	{
		stick.Path(thePinata, "Hobbies", "wrongtype", "Indoors")
		err := stick.ClearError()
		if err == nil {
			t.Error("path that hits the wrong type must result in an error")
		} else {
			t.Log(err)
		}
	}
	{
		stick.Path(thePinata, "Address")
		if err := stick.ClearError(); err != nil {
			t.Error("Address must exist", err)
		}
	}
	{
		stick.Path(thePinata, "Address", "Street")
		if err := stick.ClearError(); err != nil {
			t.Error("Address/Street must exist", err)
		}
	}
	{
		stick.Path(thePinata, "Address", "City")
		if err := stick.ClearError(); err != nil {
			t.Error("Address/City must exist", err)
		}
	}
}

func TestString(t *testing.T) {
	stick := pinata.NewStick()
	{
		stick.String(pinata.Pinata{})
		err := stick.ClearError()
		if err == nil {
			t.Error("empty pinata must not be string")
		} else {
			t.Log(err)
		}
	}
	{
		stick.String(pinata.NewPinata(0))
		err := stick.ClearError()
		if err == nil {
			t.Error("non-string pinata must result in an error")
		} else {
			t.Log(err)
		}
	}
	{
		stick.String(pinata.NewPinata(make(map[string]interface{})))
		err := stick.ClearError()
		if err == nil {
			t.Error("non-string pinata must result in an error")
		} else {
			t.Log(err)
		}
	}
	{
		stick.String(pinata.NewPinata(make([]interface{}, 0)))
		err := stick.ClearError()
		if err == nil {
			t.Error("non-string pinata must result in an error")
		} else {
			t.Log(err)
		}
	}
	if stick.String(pinata.NewPinata("hello")); stick.ClearError() != nil {
		t.Error("string pinata must not result in an error")
	}
}

func TestFloat64(t *testing.T) {
	stick := pinata.NewStick()
	{
		stick.Float64(pinata.Pinata{})
		err := stick.ClearError()
		if err == nil {
			t.Error("empty pinata must not be a float64")
		} else {
			t.Log(err)
		}
	}
	{
		stick.Float64(pinata.NewPinata(""))
		err := stick.ClearError()
		if err == nil {
			t.Error("non-float64 pinata must result in an error")
		} else {
			t.Log(err)
		}
	}
	{
		stick.Float64(pinata.NewPinata(make(map[string]interface{})))
		err := stick.ClearError()
		if err == nil {
			t.Error("non-float64 pinata must result in an error")
		} else {
			t.Log(err)
		}
	}
	{
		stick.Float64(pinata.NewPinata(make([]interface{}, 0)))
		err := stick.ClearError()
		if err == nil {
			t.Error("non-float64 pinata must result in an error")
		} else {
			t.Log(err)
		}
	}
	if stick.Float64(pinata.NewPinata(float64(0))); stick.ClearError() != nil {
		t.Error("float64 pinata must not result in an error")
	}
}

func TestBool(t *testing.T) {
	stick := pinata.NewStick()
	{
		stick.Bool(pinata.Pinata{})
		err := stick.ClearError()
		if err == nil {
			t.Error("empty pinata must not be a bool")
		} else {
			t.Log(err)
		}
	}
	{
		stick.Bool(pinata.NewPinata(""))
		err := stick.ClearError()
		if err == nil {
			t.Error("non-bool pinata must result in an error")
		} else {
			t.Log(err)
		}
	}
	{
		stick.Bool(pinata.NewPinata(make(map[string]interface{})))
		err := stick.ClearError()
		if err == nil {
			t.Error("non-bool pinata must result in an error")
		} else {
			t.Log(err)
		}
	}
	{
		stick.Bool(pinata.NewPinata(make([]interface{}, 0)))
		err := stick.ClearError()
		if err == nil {
			t.Error("non-bool pinata must result in an error")
		} else {
			t.Log(err)
		}
	}
	if stick.Bool(pinata.NewPinata(true)); stick.ClearError() != nil {
		t.Error("bool pinata must not result in an error")
	}
}

func TestNullvsAbsent(t *testing.T) {
	stick, pinata := start(t)

	stick.PathNil(pinata, "Address", "City")
	if err := stick.ClearError(); err != nil {
		t.Errorf("error: %s", err)
	}

	if stick.PathNil(pinata, "Address"); stick.ClearError() == nil {
		t.Error("Address must not be nil")
	}

	if stick.PathNil(pinata, "Address", "DoesNotExist"); stick.ClearError() == nil {
		t.Error("non-existent path must not be nil")
	}

	_ = stick.Path(pinata, "does", "not", "exist")
	if err := stick.ClearError(); err == nil {
		t.Error("non-existent path must result in an error")
	}
}
