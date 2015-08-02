package pinata_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/robbiev/pinata"
)

func start() (pinata.Stick, pinata.Pinata) {
	const message = `
	{
		"Name": "Kevin",
		"Phone": ["+44 20 7123 4567", "+44 20 4567 7123"],
		"Address": {
			"Street": "1 Gopher Road",
			"City": null
		}
	}`

	var m map[string]interface{}

	err := json.Unmarshal([]byte(message), &m)
	if err != nil {
		fmt.Println(err)
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

func TestPinata(t *testing.T) {
	stick, pinata := start()

	stick.PathString(pinata, "Phone")
	if err := stick.ClearError(); err != nil {
		fmt.Println(err)
	}
	stick.IndexString(stick.Path(pinata, "Phone"), 3)
	if err := stick.ClearError(); err != nil {
		fmt.Println(err)
	}
	stick.PathString(pinata, "Address", "City", "Town")
	if err := stick.ClearError(); err != nil {
		fmt.Println(err)
	}
}

func TestNullvsAbsent(t *testing.T) {
	stick, pinata := start()

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
