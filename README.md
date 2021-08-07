# promql

用于结构化地编写promql语句。

有的时候我们想按照某种规律生成一些promql语句，然而常规的字符串操作（比如替换、拼接等）无法很好地满足需求，那么这个小程序也许能帮上忙。

## Usage

例如想生成这条promql：

```sum by (job, mode) (rate(node_cpu_seconds_total[1m])) / on(job) group_left sum by (job) (rate(node_cpu_seconds_total[1m]))```

写法就是：

```go
Func{Fun: "histogram_quantile"}.
    WithParameters(
        Scalar(0.9),
        AggregationOp{Operator: "sum"}.
            WithByClause("le", "method", "path").
            SetOperand(
                Func{Fun: "rate"}.
                WithParameters(TSSelector{Name:"demo_api_request_duration_seconds_bucket"}.WithDuration("5m")),
            ),
        )
```

感觉还是比较清晰的。

## 参考

主要参考了prometheus的官方文档 [QUERYING PROMETHEUS](https://prometheus.io/docs/prometheus/latest/querying/basics)。

并受到了[PromLens](https://demo.promlens.com/?example)的启发。 