package motel

import (
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	trace "go.opentelemetry.io/otel/trace"
)

type Span struct {
	// Screw your private method
	sdktrace.ReadOnlySpan

	name       string
	spanKind   trace.SpanKind
	resource   *resource.Resource
	attributes []attribute.KeyValue

	traceId  [16]byte
	spanId   [8]byte
	parentId [8]byte
	flag     byte
	remote   bool

	status    sdktrace.Status
	startTime time.Time
	endTime   time.Time
}

// Creates a Open Telemetry span from parameterized inputs
func CreateSpan(
	name string,
	spanKind trace.SpanKind,
	resource *resource.Resource,
	tid [16]byte,
	pid [8]byte,
	rid [8]byte,
	flag byte,
	success bool,
	startTime time.Time,
	endTime time.Time,
) Span {
	code := codes.Ok
	if !success {
		code = codes.Error
	}

	return Span{
		name:     name,
		spanKind: spanKind,
		resource: resource,
		traceId:  tid,
		spanId:   rid,
		parentId: pid,
		flag:     flag,
		remote:   true,
		status: sdktrace.Status{
			Code:        code,
			Description: "",
		},
		startTime: startTime,
		endTime:   endTime,
	}
}

// Adds an attribute to the span
func (s *Span) WithAttribute(
	key attribute.Key,
	value attribute.Value,
) {
	s.attributes = append(
		s.attributes,
		attribute.KeyValue{
			Key:   key,
			Value: value,
		},
	)
}

// ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ //
// ~ ~ ~ ~ ~ ~ ~ ~ Open Telemetry  Interface Functions ~ ~ ~ ~ ~ ~ ~ ~ //
// ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ ~ //

func (s *Span) Name() string {
	return s.name
}

func (s *Span) SpanContext() trace.SpanContext {
	return trace.NewSpanContext(
		trace.SpanContextConfig{
			TraceID:    s.traceId,
			SpanID:     s.spanId,
			TraceFlags: trace.TraceFlags(s.flag),
			TraceState: trace.TraceState{},
			Remote:     s.remote,
		},
	)
}

func (s *Span) Parent() trace.SpanContext {
	return trace.NewSpanContext(
		trace.SpanContextConfig{
			TraceID:    s.traceId,
			SpanID:     s.parentId,
			TraceFlags: trace.TraceFlags(s.flag),
			TraceState: trace.TraceState{},
			Remote:     s.remote,
		},
	)
}

func (s *Span) SpanKind() trace.SpanKind {
	return s.spanKind
}

func (s *Span) StartTime() time.Time {
	return s.startTime
}

func (s *Span) EndTime() time.Time {
	return s.endTime
}

func (s *Span) Attributes() []attribute.KeyValue {
	return s.attributes
}

// TODO: Actually implement
func (s *Span) Links() []sdktrace.Link {
	return []sdktrace.Link{}
}

// TODO: Actually implement
func (s *Span) Events() []sdktrace.Event {
	return []sdktrace.Event{}
}

func (s *Span) Status() sdktrace.Status {
	return s.status
}

// TODO: Actually implement
func (s *Span) InstrumentationScope() instrumentation.Scope {
	return instrumentation.Scope{
		Name:      "",
		Version:   "",
		SchemaURL: "",
	}
}

// TODO: Actually implement
func (s *Span) InstrumentationLibrary() instrumentation.Library {
	return instrumentation.Library{
		Name:      "",
		Version:   "",
		SchemaURL: "",
	}
}

func (s *Span) Resource() *resource.Resource {
	return s.resource
}

// TODO: What even is this
func (s *Span) DroppedAttributes() int {
	return 0
}

// TODO: What even is this
func (s *Span) DroppedLinks() int {
	return 0
}

// TODO: What even is this
func (s *Span) DroppedEvents() int {
	return 0
}

// TODO: Is this even useful
func (s *Span) ChildSpanCount() int {
	return 0
}

