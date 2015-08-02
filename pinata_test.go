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

	stick, pinata := pinata.New(m)

	type gopher struct {
		Name  string
		Phone string
		City  string
	}

	// no error handling here
	kevin := gopher{
		Name:  stick.PathString(pinata, "Name"),
		Phone: stick.IndexString(stick.Path(pinata, "Phone"), 0),
		City:  stick.PathString(pinata, "Address", "City"),
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
