package promql

import (
	"testing"
)

func TestPromql(t *testing.T) {
	expect := "histogram_quantile(0.9000, sum by (le, method, path) (rate(demo_api_request_duration_seconds_bucket[5m])))"
	q1 := NewFunc("histogram_quantile").Parameters(
		Float(0.9),
		NewAggregationOp("sum").
			By("le", "method", "path").
			Operand(
				NewFunc("rate").Parameters(NewTSSelector("demo_api_request_duration_seconds_bucket").Duration("5m")),
			),
	)
	if actual := q1.String(); actual != expect {
		t.Fatalf("expect: %s, actual: %s", expect, actual)
	}

	expect = "sum by (job, mode) (rate(node_cpu_seconds_total[1m])) / on(job) group_left sum by (job) (rate(node_cpu_seconds_total[1m]))"
	q2 := NewBinaryOp("/").
		Matcher(NewOnVectorMatcher("job").GroupLeft()).
		Operands(
			NewAggregationOp("sum").
				By("job", "mode").
				Operand(NewFunc("rate").Parameters(NewTSSelector("node_cpu_seconds_total").Duration("1m"))),
			NewAggregationOp("sum").
				By("job").
				Operand(NewFunc("rate").Parameters(NewTSSelector("node_cpu_seconds_total").Duration("1m"))),
		)
	if actual := q2.String(); actual != expect {
		t.Fatalf("expect: %s, actual: %s", expect, actual)
	}

	expect = `1 - (sum by (instance) (increase(node_cpu_seconds_total{mode="idle", instance="master"}[1m])) / sum by (instance) (increase(node_cpu_seconds_total{instance="master"}[1m])))`
	q3 := NewBinaryOp("-").Operands(Int(1), Parenthesis{NewBinaryOp("/").Operands(
		NewAggregationOp("sum").By("instance").
			Operand(NewFunc("increase").Parameters(NewTSSelector("node_cpu_seconds_total").Labels(
				NewLabel("mode", "=", "idle"), NewLabel("instance", "=", "master")).Duration("1m"))),

		NewAggregationOp("sum").By("instance").
			Operand(NewFunc("increase").Parameters(NewTSSelector("node_cpu_seconds_total").Labels(
				NewLabel("instance", "=", "master")).Duration("1m"))),
	)})

	if actual := q3.String(); actual != expect {
		t.Fatalf("expect: %s, actual: %s", expect, actual)
	}
}
