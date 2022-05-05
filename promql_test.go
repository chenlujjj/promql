package promql

import (
	"testing"
)

func TestPromql(t *testing.T) {
	expect := "histogram_quantile(0.9000, sum by (le, method, path) (rate(demo_api_request_duration_seconds_bucket[5m])))"
	q1 := NewFunc("histogram_quantile").WithParameters(
		Scalar(0.9),
		NewAggregationOp("sum").
			WithByClause("le", "method", "path").
			SetOperand(
				NewFunc("rate").WithParameters(TSSelector{Name: "demo_api_request_duration_seconds_bucket"}.WithDuration("5m")),
			),
	)
	if actual := q1.String(); actual != expect {
		t.Fatalf("expect: %s, actual: %s", expect, actual)
	}

	expect = "sum by (job, mode) (rate(node_cpu_seconds_total[1m])) / on(job) group_left sum by (job) (rate(node_cpu_seconds_total[1m]))"
	q2 := NewBinaryOp("/").
		WithMatcher(NewOnVectorMatcher("job").WithGroupLeft()).
		WithOperands(
			NewAggregationOp("sum").
				WithByClause("job", "mode").
				SetOperand(NewFunc("rate").WithParameters(TSSelector{Name: "node_cpu_seconds_total"}.WithDuration("1m"))),
			NewAggregationOp("sum").
				WithByClause("job").
				SetOperand(NewFunc("rate").WithParameters(TSSelector{Name: "node_cpu_seconds_total"}.WithDuration("1m"))),
		)
	if actual := q2.String(); actual != expect {
		t.Fatalf("expect: %s, actual: %s", expect, actual)
	}

	expect = `1 - (sum by (instance) (increase(node_cpu_seconds_total{mode="idle", instance="master"}[1m])) / sum by (instance) (increase(node_cpu_seconds_total{instance="master"}[1m])))`
	q3 := NewBinaryOp("-").WithOperands(Int(1), NewBinaryOp("/").WithOperands(
		NewAggregationOp("sum").WithByClause("instance").
			SetOperand(NewFunc("increase").WithParameters(TSSelector{Name: "node_cpu_seconds_total"}.WithLabels(Label{
				Key:     "mode",
				Matcher: "=",
				Value:   "idle",
			}).WithLabels(Label{
				Key:     "instance",
				Matcher: "=",
				Value:   "master",
			}).WithDuration("1m"))),
		NewAggregationOp("sum").WithByClause("instance").
			SetOperand(NewFunc("increase").WithParameters(TSSelector{Name: "node_cpu_seconds_total"}.WithLabels(Label{
				Key:     "instance",
				Matcher: "=",
				Value:   "master",
			}).WithDuration("1m"))),
	))

	if actual := q3.String(); actual != expect {
		t.Fatalf("expect: %s, actual: %s", expect, actual)
	}
}
