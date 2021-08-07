# promql

用于结构化地编写promql语句。

有的时候我们想按照某种规律生成一些promql语句，然而常规的字符串操作（比如替换、拼接等）无法很好地满足需求，那么这个小程序也许能帮上忙。


## Install

`go get -u github.com/chenlujjj/promql`

## Usage

例如想生成这条promql：

```sum by (job, mode) (rate(node_cpu_seconds_total[1m])) / on(job) group_left sum by (job) (rate(node_cpu_seconds_total[1m]))```

```go
package main

import (
	"fmt"

	"github.com/chenlujjj/promql"
)

func main() {
	q := promql.NewBinaryOp("/").
		WithMatcher(promql.NewVectorMatcher("on", "job").WithGroupLeft()).
		WithOperands(
			promql.NewAggregationOp("sum").
				WithByClause("job", "mode").
				SetOperand(promql.NewFunc("rate").WithParameters(promql.TSSelector{Name: "node_cpu_seconds_total"}.WithDuration("1m"))),
			promql.NewAggregationOp("sum").
				WithByClause("job").
				SetOperand(promql.NewFunc("rate").WithParameters(promql.TSSelector{Name: "node_cpu_seconds_total"}.WithDuration("1m"))),
		)
	fmt.Println(q.String())
}
```

通过链式调用，整个生成过程还是比较清晰的。

## 参考

主要参考了prometheus的官方文档 [QUERYING PROMETHEUS](https://prometheus.io/docs/prometheus/latest/querying/basics)。

并受到了[PromLens](https://demo.promlens.com/?example)的启发。
