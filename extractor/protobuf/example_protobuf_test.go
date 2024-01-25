package protobuf_test

import (
	"fmt"
	"log"

	"github.com/zoncoen/query-go"

	protobufextractor "github.com/zoncoen/query-go/extractor/protobuf"
	testpb "github.com/zoncoen/query-go/extractor/protobuf/testdata/gen/testpb"
)

func ExampleExtractFunc() {
	v := testpb.OneofMessage{
		Value: &testpb.OneofMessage_B_{
			B: &testpb.OneofMessage_B{
				BarValue: "yyy",
			},
		},
	}
	q := query.New(
		query.CustomExtractFunc(protobufextractor.ExtractFunc()),
		query.CustomIsInlineStructFieldFunc(protobufextractor.OneofIsInlineStructFieldFunc()),
	).Key("B").Key("bar_value")
	got, err := q.Extract(v)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(got)
	// Output:
	// yyy
}
