package trace

import (
	"io"
	"fmt"
)

// Tracer is the interface that describes an object capable of
// tracing events throughout code.
type Tracer interface {
	Trace(...interface{})
}

// 下面是一种实现，如果只由上面定义接口，import后自实现也可以
type tracer struct {
	out io.Writer
}

func (t *tracer) Trace(a ...interface{}) {
	fmt.Fprint(t.out, a...)
	fmt.Fprintln(t.out)
}

// 注意，tracer struct 是不可导出的
//the user will only ever see an object that satisfies the Tracer interface and will never even know about our private tracer type.
func New(w io.Writer) Tracer {
	return &tracer{out: w}
}

// 实现一种调用Off后返回一个不记录任何东西的tracer，以免写入太多无效信息
type nilTracer struct{}

func (t *nilTracer) Trace(a ...interface{}) {}

// Off creates a Tracer that will ignore calls to Trace.
func Off() Tracer {
	return &nilTracer{}
}