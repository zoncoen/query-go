package yaml_test

import (
	"fmt"
	"log"

	"github.com/goccy/go-yaml"
	"github.com/zoncoen/query-go"

	yamlextractor "github.com/zoncoen/query-go/extractor/yaml"
)

func ExampleMapSliceExtractFunc() {
	b := []byte(`- foo: bar`)
	var v interface{}
	if err := yaml.UnmarshalWithOptions(b, &v, yaml.UseOrderedMap()); err != nil {
		log.Fatal(err)
	}

	q := query.New(
		query.CaseInsensitive(),
		query.CustomExtractFunc(yamlextractor.MapSliceExtractFunc()),
	).Index(0).Key("FOO")
	got, err := q.Extract(v)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(got)
	// Output:
	// bar
}
