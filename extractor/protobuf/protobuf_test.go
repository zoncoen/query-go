package protobuf

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	testpb "github.com/zoncoen/query-go/extractor/protobuf/testdata/gen/testpb"
	"github.com/zoncoen/query-go/v2"
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
				got, err := test.query.Extract(context.Background(), test.v)
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
				_, err := test.query.Extract(context.Background(), test.v)
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
				got, err := test.query.Extract(context.Background(), test.v)
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

func TestExtractFunc_PropagatesFailures(t *testing.T) {
	// A custom func downstream of ExtractFunc that fails hard whenever the
	// probe value is offered, simulating e.g. a canceled blocking extractor.
	failing := func(f query.ExtractFunc) query.ExtractFunc {
		return func(ctx context.Context, v reflect.Value) (reflect.Value, error) {
			if v.IsValid() && v.CanInterface() {
				if _, ok := v.Interface().(query.KeyExtractor); ok {
					return reflect.Value{}, context.Canceled
				}
			}
			return f(ctx, v)
		}
	}
	q := query.New(
		query.CustomExtractFunc(ExtractFunc()),
		query.CustomExtractFunc(failing),
	).Key("nosuch")
	_, err := q.Extract(context.Background(), &testpb.OneofMessage_A{})
	if err == nil {
		t.Fatal("expected an error")
	}
	if errors.Is(err, query.ErrNotFound) {
		t.Fatalf("a failure must not be classified as absence: %s", err)
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled to propagate but got: %s", err)
	}
}
