package yaml_test

import (
	"context"
	"fmt"
	"log"

	"github.com/goccy/go-yaml"
	"github.com/zoncoen/query-go/v2"

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
	got, err := q.Extract(context.Background(), v)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(got)
	// Output:
	// bar
}
