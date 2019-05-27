package query_test

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/zoncoen/query-go"
)

// Person represents a person.
type Person struct {
	Name string `json:"name,omitempty"`
}

// getFieldNameByJSONTag returns the JSON field tag as field name if exists.
func getFieldNameByJSONTag(field reflect.StructField) string {
	tag, ok := field.Tag.Lookup("json")
	if ok {
		strs := strings.Split(tag, ",")
		return strs[0]
	}
	return field.Name
}

func ExampleCustomStructFieldNameGetter() {
	person := Person{
		Name: "Alice",
	}

	q := query.New(
		query.CustomStructFieldNameGetter(getFieldNameByJSONTag),
	).Key("name")
	name, _ := q.Extract(person)
	fmt.Println(name)
	// Output:
	// Alice
}
