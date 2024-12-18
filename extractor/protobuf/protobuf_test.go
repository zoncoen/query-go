package protobuf

import (
	"strings"
	"testing"

	"github.com/zoncoen/query-go"
	testpb "github.com/zoncoen/query-go/extractor/protobuf/testdata/gen/testpb"
)

func TestExtractFunc(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tests := map[string]struct {
			query  *query.Query
			v      any
			expect any
		}{
			"by field name": {
				query: query.New(
					query.CustomExtractFunc(ExtractFunc()),
				).Key("Value").Key("B").Key("BarValue"),
				v: testpb.OneofMessage{
					Value: &testpb.OneofMessage_B_{
						B: &testpb.OneofMessage_B{
							BarValue: "yyy",
						},
					},
				},
				expect: "yyy",
			},
			"by struct tag": {
				query: query.New(
					query.CustomExtractFunc(ExtractFunc()),
				).Key("Value").Key("B").Key("bar_value"),
				v: testpb.OneofMessage{
					Value: &testpb.OneofMessage_B_{
						B: &testpb.OneofMessage_B{
							BarValue: "yyy",
						},
					},
				},
				expect: "yyy",
			},
			"by struct tag (case insensitive)": {
				query: query.New(
					query.CaseInsensitive(),
					query.CustomExtractFunc(ExtractFunc()),
				).Key("Value").Key("B").Key("BAR_VALUE"),
				v: testpb.OneofMessage{
					Value: &testpb.OneofMessage_B_{
						B: &testpb.OneofMessage_B{
							BarValue: "yyy",
						},
					},
				},
				expect: "yyy",
			},
			"by struct tag json": {
				query: query.New(
					query.CustomExtractFunc(ExtractFunc()),
				).Key("Value").Key("B").Key("barValue"),
				v: testpb.OneofMessage{
					Value: &testpb.OneofMessage_B_{
						B: &testpb.OneofMessage_B{
							BarValue: "yyy",
						},
					},
				},
				expect: "yyy",
			},
			"by struct tag json (case insensitive)": {
				query: query.New(
					query.CaseInsensitive(),
					query.CustomExtractFunc(ExtractFunc()),
				).Key("Value").Key("B").Key("BARVALUE"),
				v: testpb.OneofMessage{
					Value: &testpb.OneofMessage_B_{
						B: &testpb.OneofMessage_B{
							BarValue: "yyy",
						},
					},
				},
				expect: "yyy",
			},
		}
		for name, test := range tests {
			test := test
			t.Run(name, func(t *testing.T) {
				got, err := test.query.Extract(test.v)
				if err != nil {
					t.Fatalf("unexpected error: %s", err)
				}
				if got != test.expect {
					t.Errorf("expect %v but got %v", test.expect, got)
				}
			})
		}
	})
	t.Run("failure", func(t *testing.T) {
		tests := map[string]struct {
			query  *query.Query
			v      any
			expect string
		}{
			"not found": {
				query: query.New(
					query.CustomExtractFunc(ExtractFunc()),
				).Key("Value").Key("B").Key("BAR_VALUE"),
				v: testpb.OneofMessage{
					Value: &testpb.OneofMessage_B_{
						B: &testpb.OneofMessage_B{
							BarValue: "yyy",
						},
					},
				},
				expect: `".Value.B.BAR_VALUE" not found`,
			},
		}
		for name, test := range tests {
			test := test
			t.Run(name, func(t *testing.T) {
				_, err := test.query.Extract(test.v)
				if err == nil {
					t.Fatal("no error")
				}
				if got := err.Error(); !strings.Contains(got, test.expect) {
					t.Errorf("expect %v but got %v", test.expect, got)
				}
			})
		}
	})
}

func TestOneofIsInlineStructFieldFunc(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tests := map[string]struct {
			query  *query.Query
			v      any
			expect any
		}{
			"omit .Value": {
				query: query.New(
					query.CustomIsInlineStructFieldFunc(OneofIsInlineStructFieldFunc()),
				).Key("B").Key("BarValue"),
				v: testpb.OneofMessage{
					Value: &testpb.OneofMessage_B_{
						B: &testpb.OneofMessage_B{
							BarValue: "yyy",
						},
					},
				},
				expect: "yyy",
			},
		}
		for name, test := range tests {
			test := test
			t.Run(name, func(t *testing.T) {
				got, err := test.query.Extract(test.v)
				if err != nil {
					t.Fatalf("unexpected error: %s", err)
				}
				if got != test.expect {
					t.Errorf("expect %v but got %v", test.expect, got)
				}
			})
		}
	})
}
