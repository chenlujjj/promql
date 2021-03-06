package promql

import (
	"fmt"
	"strings"
)

type Stringer interface {
	String() string
}

type Node interface {
	Stringer
	Self() string
	Children() []Node
}

// --- constant node

type ConstantStringNode struct {
	constant string
}

func NewConstantStringNode(constantString string) ConstantStringNode {
	return ConstantStringNode{constant: constantString}
}

func (m ConstantStringNode) String() string {
	return m.Self()
}

func (m ConstantStringNode) Self() string {
	return m.constant
}

func (m ConstantStringNode) Children() []Node {
	return nil
}

var _ Node = (*ConstantStringNode)(nil)

// --- time series selector, see https://prometheus.io/docs/prometheus/latest/querying/basics/#time-series-selectors

type TSSelector struct {
	name     string
	labels   []Label
	duration string // 可选, 比如5m
	offset   string // 可选，比如offset 5m
}

var _ Node = (*TSSelector)(nil)

func (m TSSelector) String() string {
	return m.Self()
}

func (m TSSelector) Self() string {
	if m.name == "" && len(m.labels) == 0 {
		panic("metric name and labels cannot be both empty")
	}
	s := m.name
	if len(m.labels) != 0 {
		labelStrings := make([]string, 0, len(m.labels))
		for _, label := range m.labels {
			labelStrings = append(labelStrings, label.String())
		}
		s += fmt.Sprintf("{%s}", strings.Join(labelStrings, ", "))
	}
	if m.duration != "" {
		s += fmt.Sprintf("[%s]", m.duration)
	}
	if m.offset != "" {
		s += fmt.Sprintf(" offset %s", m.offset)
	}
	return s
}

func (m TSSelector) Children() []Node {
	return nil
}

func NewTSSelector(name string, labels ...Label) TSSelector {
	return TSSelector{
		name:   name,
		labels: labels,
	}
}

func (m TSSelector) Labels(labels ...Label) TSSelector {
	m.labels = append(m.labels, labels...)
	return m
}

func (m TSSelector) Duration(duration string) TSSelector {
	m.duration = duration
	return m
}

func (m TSSelector) Offset(offset string) TSSelector {
	m.offset = offset
	return m
}

type Label struct {
	Key     string
	Value   string
	Matcher string // = != =~ !~
}

func NewLabel(key, matcher, value string) Label {
	return Label{Key: key, Value: value, Matcher: matcher}
}

func (l Label) String() string {
	return fmt.Sprintf(`%s%s"%s"`, l.Key, l.Matcher, l.Value)
}

// --- query functions, see https://prometheus.io/docs/prometheus/latest/querying/functions

type Func struct {
	fun        string
	parameters []Node // 长度不定， 1， 2，等
}

func NewFunc(fun string, parameters ...Node) Func {
	return Func{fun: fun, parameters: parameters}
}

func (f Func) Parameters(params ...Node) Func {
	f.parameters = append(f.parameters, params...)
	return f
}

var _ Node = (*Func)(nil)

func (f Func) String() string {
	params := make([]string, 0, len(f.parameters))
	for _, p := range f.parameters {
		params = append(params, p.String())
	}
	return fmt.Sprintf("%s(%s)", f.Self(), strings.Join(params, ", "))
}

func (f Func) Self() string {
	return f.fun
}

func (f Func) Children() []Node {
	return f.parameters
}

// --- 二元操作符, binary operators, see https://prometheus.io/docs/prometheus/latest/querying/operators/#binary-operators

type BinaryOp struct {
	operator    string         // + - * / == != > < >= <= and or unless
	left, right Node           // operands
	matcher     *VectorMatcher // 可选
}

func NewBinaryOp(operator string) BinaryOp {
	return BinaryOp{operator: operator}
}

func (bo BinaryOp) Operands(left, right Node) BinaryOp {
	bo.left = left
	bo.right = right
	return bo
}

func (bo BinaryOp) Left(l Node) BinaryOp {
	bo.left = l
	return bo
}

func (bo BinaryOp) Right(r Node) BinaryOp {
	bo.right = r
	return bo
}

func (bo BinaryOp) Matcher(vm VectorMatcher) BinaryOp {
	bo.matcher = &vm
	return bo
}

var _ Node = (*BinaryOp)(nil)

func (bo BinaryOp) String() string {
	return fmt.Sprintf("%s %s %s", bo.left.String(), bo.Self(), bo.right.String())
}

func (bo BinaryOp) Self() string {
	s := bo.operator
	if bo.matcher != nil {
		s += " " + bo.matcher.String()
	}
	return s
}

func (bo BinaryOp) Children() []Node {
	return []Node{bo.left, bo.right}
}

// --- See https://prometheus.io/docs/prometheus/latest/querying/operators/#vector-matching

type VectorMatcher struct {
	keyword string // on/ignoring
	labels  []string
	group   *GroupModifier // 可选
}

func NewVectorMatcher(keyword string, labels ...string) VectorMatcher {
	return VectorMatcher{keyword: keyword, labels: labels}
}

func NewOnVectorMatcher(labels ...string) VectorMatcher {
	return VectorMatcher{keyword: "on", labels: labels}
}

func NewIgnoringVectorMatcher(labels ...string) VectorMatcher {
	return VectorMatcher{keyword: "ignoring", labels: labels}
}

func (vm VectorMatcher) String() string {
	s := fmt.Sprintf("%s(%s)", vm.keyword, strings.Join(vm.labels, ", "))
	if vm.group != nil {
		s += " " + vm.group.String()
	}
	return s
}

func (vm VectorMatcher) Labels(labels ...string) VectorMatcher {
	vm.labels = append(vm.labels, labels...)
	return vm
}

func (vm VectorMatcher) GroupLeft(labels ...string) VectorMatcher {
	vm.group = &GroupModifier{left: true, labels: labels}
	return vm
}

func (vm VectorMatcher) GroupRight(labels ...string) VectorMatcher {
	vm.group = &GroupModifier{left: false, labels: labels}
	return vm
}

type GroupModifier struct {
	left   bool
	labels []string
}

func (gm GroupModifier) String() string {
	var group string
	if gm.left {
		group = "group_left"
	} else {
		group = "group_right"
	}
	if len(gm.labels) == 0 {
		return group
	}
	return fmt.Sprintf("%s(%s)", group, strings.Join(gm.labels, ", "))
}

// --- 聚合操作符, see https://prometheus.io/docs/prometheus/latest/querying/operators/#aggregation-operators

type AggregationOp struct {
	operator  string             // sum, min, max, avg, topk, count, quantile ...
	operand   Node               // aggregate a single instant vector
	clause    *AggregationClause // 可选，比如 by(code)
	parameter *Scalar            // only required for count_values, quantile, topk and bottomk.
}

func NewAggregationOp(operator string) AggregationOp {
	return AggregationOp{operator: operator}
}

func (ao AggregationOp) Operand(operand Node) AggregationOp {
	ao.operand = operand
	return ao
}

func (ao AggregationOp) By(labels ...string) AggregationOp {
	ao.clause = &AggregationClause{keyword: "by", labels: labels}
	return ao
}

func (ao AggregationOp) Without(labels ...string) AggregationOp {
	ao.clause = &AggregationClause{keyword: "without", labels: labels}
	return ao
}

func (ao AggregationOp) WithParameter(param Scalar) AggregationOp {
	ao.parameter = &param
	return ao
}

var _ Node = (*AggregationOp)(nil)

func (ao AggregationOp) Self() string {
	s := ao.operator
	if ao.clause != nil {
		s += " " + ao.clause.String()
	}
	return s
}

func (ao AggregationOp) String() string {
	if ao.parameter != nil {
		return fmt.Sprintf("%s (%f, %s)", ao.Self(), *ao.parameter, ao.operand.String())
	}
	return fmt.Sprintf("%s (%s)", ao.Self(), ao.operand.String())
}

func (ao AggregationOp) Children() []Node {
	return []Node{ao.operand}
}

type AggregationClause struct {
	keyword string // by, without
	labels  []string
}

func (ac AggregationClause) String() string {
	return fmt.Sprintf("%s (%s)", ac.keyword, strings.Join(ac.labels, ", "))
}

// --- Wrap origin node with parenthesis

type Parenthesis struct {
	Node
}

func (p Parenthesis) String() string {
	return fmt.Sprintf("(%s)", p.Node.String())
}

// --- Scalar, see https://prometheus.io/docs/prometheus/latest/querying/basics/#float-literals

type Scalar Node

// --- 浮点数表示的标量

type Float float64

var _ Scalar = (*Float)(nil)

func (f Float) String() string {
	return f.Self()
}

// TODO: 处理显示的精度
func (f Float) Self() string {
	return fmt.Sprintf("%.4f", f)
}

func (f Float) Children() []Node {
	return nil
}

// --- 整型表示的标量

type Int int

var _ Scalar = (*Int)(nil)

func (i Int) String() string {
	return i.Self()
}

func (i Int) Self() string {
	return fmt.Sprintf("%d", i)
}

func (i Int) Children() []Node {
	return nil
}
