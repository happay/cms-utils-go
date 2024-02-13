package tracing

import (
	"fmt"

	"github.com/happay/cms-utils-go/v2/logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

// InitTracing, will set the tracerProvider as global tracer
// and sets propagator as the global TextMapPropagator.
func (tp *DataDogProvider) InitTracing() {
	otel.SetTracerProvider(tp.TracerProvider)
	p := NewPropagator()
	otel.SetTextMapPropagator(p)
	defer func() {
		if r := recover(); r != nil {
			err := fmt.Errorf("error rendering template: %s", r)
			logger.GetLoggerV3().Error(err.Error())
			panic(r)
		}
	}()
}

// NewPropagator, return the TextMapPropogator which allows distributed tracing
func NewPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}
