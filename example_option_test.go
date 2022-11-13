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

func ExampleCaseInsensitive() {
	person := Person{
		Name: "Alice",
	}
	q := query.New(query.CaseInsensitive()).Key("NAME")
	name, _ := q.Extract(person)
	fmt.Println(name)
	// Output:
	// Alice
}

func ExampleExtractByStructTag() {
	person := Person{
		Name: "Alice",
	}
	q := query.New(query.ExtractByStructTag("json")).Key("name")
	name, _ := q.Extract(person)
	fmt.Println(name)
	// Output:
	// Alice
}

func ExampleCustomExtractFunc() {
	person := Person{
		Name: "Alice",
	}
	q := query.New(
		query.CustomExtractFunc(func(f query.ExtractFunc) query.ExtractFunc {
			return func(v reflect.Value) (reflect.Value, bool) {
				return reflect.ValueOf("Bob"), true
			}
		}),
	).Key("name")
	name, _ := q.Extract(person)
	fmt.Println(name)
	// Output:
	// Bob
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
