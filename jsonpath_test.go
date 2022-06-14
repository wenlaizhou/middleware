package middleware

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

func TestJsonPath(t *testing.T) {
	if c, err := JsonpathCompile("$.person.name"); err == nil {
		data := map[string]map[string]string{
			"person": {
				"name": "hello name",
			},
		}
		if res, err := c.Lookup(data); err == nil {
			result := fmt.Sprintf("%v", res)
			println(result)
		} else {
			t.Fatalf("%v", err.Error())
		}
	} else {
		t.Fatalf("%v", err.Error())
	}
}

func TestJsonPathString(t *testing.T) {
	if c, err := JsonpathCompile("$"); err == nil {
		var data interface{}
		json.Unmarshal([]byte(`
{
	"person" : {
		"name" : "james"
	},
	"papa" : [

		{}, {
			"name" : "james"
		}
	]
}

`), &data)
		switch reflect.TypeOf(data).Kind() {
		case reflect.Map:
			println("map")
			println(len(reflect.ValueOf(data).MapKeys()))
			break
		case reflect.Slice:
			println("list")
			println(reflect.ValueOf(data).Len())
			break
		}
		if res, err := c.Lookup(data); err == nil {
			result := fmt.Sprintf("%v", res)
			println(result)
		} else {
			t.Fatalf("%v", err.Error())
		}
	} else {
		t.Fatalf("%v", err.Error())
	}
}
