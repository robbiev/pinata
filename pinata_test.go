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

	stick, pinata := pinata.New(m)

	type gopher struct {
		Name  string
		Phone string
		City  string
	}

	// no error handling here
	kevin := gopher{
		Name:  stick.StringAtPath(pinata, "Name"),
		Phone: stick.StringAtIndex(stick.PinataAtPath(pinata, "Phone"), 0),
		City:  stick.StringAtPath(pinata, "Address", "City"),
	}

	if err := stick.Error(); err != nil {
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

	stick, pinata := pinata.New(m)

	stick.StringAtPath(pinata, "Phone")
	if err := stick.Error(); err != nil {
		fmt.Println(err)
		stick.ClearError()
	}
	stick.StringAtIndex(stick.PinataAtPath(pinata, "Phone"), 3)
	if err := stick.Error(); err != nil {
		fmt.Println(err)
		stick.ClearError()
	}
	stick.StringAtPath(pinata, "Address", "City", "Town")
	if err := stick.Error(); err != nil {
		fmt.Println(err)
		stick.ClearError()
	}
}
