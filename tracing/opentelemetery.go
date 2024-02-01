package tracing

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

// InitTracing, will set the tracerProvider as global tracer
// and sets propagator as the global TextMapPropagator.
func (tp *DataDogProvider) InitTracing() {
	otel.SetTracerProvider(tp.TracerProvider)
	p := NewPropagator()
	otel.SetTextMapPropagator(p)
}

// NewPropagator, return the TextMapPropogator which allows distributed tracing
func NewPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}
