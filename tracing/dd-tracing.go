package tracing

import (
	ddotel "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/opentelemetry"
	ddtracer "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

type DataDogTracerConfig struct {
	ServiceName string `json:"service_name" validate:"required"`
	Host        string `json:"host" validate:"required"`
	Port        int    `json:"port" validate:"required"`
	Env         string `json:"env"`
	Version     string `json:"version"`
}

type DataDogProvider struct {
	TracerConfig   *DataDogTracerConfig
	TracerProvider *ddotel.TracerProvider
}

// NewTracerProvider initializes the datadog Tracer with the provided start option
func (tp *DataDogProvider) NewTracerProvider() {
	tp.TracerProvider = ddotel.NewTracerProvider(
		ddtracer.WithEnv(tp.TracerConfig.Env),
		ddtracer.WithService(tp.TracerConfig.ServiceName),
		ddtracer.WithServiceVersion(tp.TracerConfig.Version),
		ddtracer.WithHostname(tp.TracerConfig.Host),
	)
}
