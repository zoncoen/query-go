package protobuf_test

import (
	"context"
	"fmt"
	"log"

	"github.com/zoncoen/query-go/v2"

	protobufextractor "github.com/zoncoen/query-go/extractor/protobuf"
	testpb "github.com/zoncoen/query-go/extractor/protobuf/testdata/gen/testpb"
)

func ExampleExtractFunc() {
	v := &testpb.OneofMessage{
		Value: &testpb.OneofMessage_B_{
			B: &testpb.OneofMessage_B{
				BarValue: "yyy",
			},
		},
	}
	q := query.New(
		query.CustomExtractFunc(protobufextractor.ExtractFunc()),
		query.CustomIsInlineStructFieldFunc(protobufextractor.OneofIsInlineStructFieldFunc()),
	).Key("b").Key("bar_value")
	got, err := q.Extract(context.Background(), v)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(got)
	// Output:
	// yyy
}
