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
	s.traceId = config.TraceId
	s.spanId = config.SpanId
	s.parentSpanId = config.ParentSpanId
	s.startTime = time.Now()

	parentSpan := GetObsContext(ctx).GetSpan()
	if parentSpan != nil {
		s.traceId = parentSpan.traceId
		s.parentSpanId = parentSpan.spanId
		s.spanId = GenerateSpanIdFunc()
	} else {
		if s.traceId == "" {
			s.traceId = GenerateTraceIdFunc()
		}
		if s.spanId == "" {
			s.spanId = GenerateSpanIdFunc()
		}
		if s.parentSpanId == "" {
			s.parentSpanId = GenerateSpanIdFunc()
		}
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
