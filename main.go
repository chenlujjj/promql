package main

import (
	"fmt"
	"strings"
)

type Stringer interface {
	String() string
}

type EmptyString struct{}

func (es EmptyString) Stringer() string { return "" }

type Node interface {
	Stringer
	Self() string
	Children() []Node
}

var _ Node = (*TSSelector)(nil)

// time series selector
type TSSelector struct {
	Name     string
	Labels   []Label
	Duration string // 可选, 比如5m
	Offset   string // 可选，比如offset 5m
}

func (m TSSelector) String() string {
	return m.Self()
}

func (m TSSelector) Self() string {
	if m.Name == "" && len(m.Labels) == 0 {
		panic("metric name and labels cannot be both empty")
	}
	s := m.Name
	if len(m.Labels) != 0 {
		labelStrings := make([]string, 0, len(m.Labels))
		for _, label := range m.Labels {
			labelStrings = append(labelStrings, label.Stringer())
		}
		s += fmt.Sprintf("{%s}", strings.Join(labelStrings, ", "))
	}
	if m.Duration != "" {
		s += fmt.Sprintf("[%s]", m.Duration)
	}
	if m.Offset != "" {
		s += fmt.Sprintf(" offset %s", m.Offset)
	}
	return s
}

func (m TSSelector) Children() []Node {
	return nil
}

func (m TSSelector) WithLabels(labels ...Label) TSSelector {
	m.Labels = append(m.Labels, labels...)
	return m
}

func (m TSSelector) WithDuration(duration string) TSSelector {
	m.Duration = duration
	return m
}

func (m TSSelector) WithOffset(offset string) TSSelector {
	m.Offset = offset
	return m
}

type Label struct {
	Key     string
	Value   string
	Matcher string // = != =~
}

func (l Label) Stringer() string {
	return fmt.Sprintf(`%s%s"%s"`, l.Key, l.Matcher, l.Value)
}

// 函数， 比如 rate
type Func struct {
	Fun        string
	Parameters []Node // 长度不定， 1， 2，等
}

func (f Func) WithParameters(params ...Node) Func {
	f.Parameters = append(f.Parameters, params...)
	return f
}

var _ Node = (*Func)(nil)

func (f Func) String() string {
	params := make([]string, 0, len(f.Parameters))
	for _, p := range f.Parameters {
		params = append(params, p.String())
	}
	return fmt.Sprintf("%s(%s)", f.Self(), strings.Join(params, ", "))
}

func (f Func) Self() string {
	return f.Fun
}

func (f Func) Children() []Node {
	return f.Parameters
}

// 二元操作符
type BinaryOp struct {
	Operator string         // + - * / == != > < >= <= and or unless
	Operands []Node         // 长度为2
	Matcher  *VectorMatcher // 可选
}

func (bo BinaryOp) WithOperands(left, right Node) BinaryOp {
	bo.Operands = []Node{left, right}
	return bo
}

func (bo BinaryOp) WithMatcher(vm VectorMatcher) BinaryOp {
	bo.Matcher = &vm
	return bo
}

var _ Node = (*BinaryOp)(nil)

func (bo BinaryOp) String() string {
	return fmt.Sprintf("%s %s %s", bo.Operands[0].String(), bo.Self(), bo.Operands[1].String())
}

func (bo BinaryOp) Self() string {
	s := bo.Operator
	if bo.Matcher != nil {
		s += " " + bo.Matcher.String()
	}
	return s
}

func (bo BinaryOp) Children() []Node {
	return bo.Operands
}

type VectorMatcher struct {
	Keyword string // on/ignoring
	Labels  []string
	Group   *GroupModifier // 可选
}

func (vm VectorMatcher) String() string {
	s := fmt.Sprintf("%s(%s)", vm.Keyword, strings.Join(vm.Labels, ", "))
	if vm.Group != nil {
		s += " " + vm.Group.String()
	}
	return s
}

func (vm VectorMatcher) WithGroupLeft(labels ...string) VectorMatcher {
	vm.Group = &GroupModifier{Left: true, Labels: labels}
	return vm
}

func (vm VectorMatcher) WithGroupRight(labels ...string) VectorMatcher {
	vm.Group = &GroupModifier{Left: false, Labels: labels}
	return vm
}

type GroupModifier struct {
	Left   bool
	Labels []string
}

func (gm GroupModifier) String() string {
	var group string
	if gm.Left {
		group = "group_left"
	} else {
		group = "group_right"
	}
	if len(gm.Labels) == 0 {
		return group
	}
	return fmt.Sprintf("%s(%s)", group, strings.Join(gm.Labels, ", "))
}

// 聚合操作符
type AggregationOp struct {
	Operator  string             // sum, min, max, avg, topk, count, quantile ...
	Operand   Node               // aggregate a single instant vector
	Clause    *AggregationClause // 可选，比如 by(code)
	Parameter *Scalar            // only required for count_values, quantile, topk and bottomk.
}

func (ao AggregationOp) SetOperand(operand Node) AggregationOp {
	ao.Operand = operand
	return ao
}

func (ao AggregationOp) WithByClause(labels ...string) AggregationOp {
	ao.Clause = &AggregationClause{Keyword: "by", Labels: labels}
	return ao
}

func (ao AggregationOp) WithWithoutClause(labels ...string) AggregationOp {
	ao.Clause = &AggregationClause{Keyword: "without", Labels: labels}
	return ao
}

func (ao AggregationOp) WithClause(keyword string, labels ...string) AggregationOp {
	ao.Clause = &AggregationClause{Keyword: keyword, Labels: labels}
	return ao
}

func (ao AggregationOp) WithParameter(param Scalar) AggregationOp {
	ao.Parameter = &param
	return ao
}

var _ Node = (*AggregationOp)(nil)

func (ao AggregationOp) Self() string {
	s := ao.Operator
	if ao.Clause != nil {
		s += " " + ao.Clause.String()
	}
	return s
}

func (ao AggregationOp) String() string {
	if ao.Parameter != nil {
		return fmt.Sprintf("%s (%f, %s)", ao.Self(), *ao.Parameter, ao.Operand.String())
	}
	return fmt.Sprintf("%s (%s)", ao.Self(), ao.Operand.String())
}

func (ao AggregationOp) Children() []Node {
	return []Node{ao.Operand}
}

type AggregationClause struct {
	Keyword string
	Labels  []string
}

func (ac AggregationClause) String() string {
	return fmt.Sprintf("%s (%s)", ac.Keyword, strings.Join(ac.Labels, ", "))
}

// 浮点数标量
type Scalar float64

var _ Node = (*Scalar)(nil)

func (s Scalar) String() string {
	return s.Self()
}

// TODO: 妥善处理显示的精度
func (s Scalar) Self() string {
	return fmt.Sprintf("%.4f", s)
}

func (s Scalar) Children() []Node {
	return nil
}
