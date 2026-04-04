package xobs

import (
	"context"
	"time"
)

type Span struct {
	name         string
	startTime    time.Time
	cost         time.Duration
	traceId      string
	spanId       string
	parentSpanId string
}

func (s *Span) init(ctx context.Context, config *SpanConfig) {
	s.name = config.Name
	s.startTime = time.Now()
	if config.TraceId != "" {
		s.traceId = config.TraceId
		s.spanId = config.SpanId
		s.parentSpanId = config.ParentSpanId
	} else if parentSpan := GetObsContext(ctx).GetSpan(); parentSpan != nil {
		s.traceId = parentSpan.traceId
		s.parentSpanId = parentSpan.spanId
		s.spanId = GenerateSpanIdFunc(false)
	} else {
		s.traceId = GenerateTraceIdFunc()
		s.spanId = GenerateSpanIdFunc(false)
		s.parentSpanId = GenerateSpanIdFunc(true)
	}
}

func (s *Span) GetTraceId() string {
	return s.traceId
}

func (s *Span) GetSpanId() string {
	return s.spanId
}

func (s *Span) GetParentSpanId() string {
	return s.parentSpanId
}

func (s *Span) SetAttr(key string, val any) {

}
