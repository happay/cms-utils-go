package tracing

import (
	"fmt"

	ddotel "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/opentelemetry"
	ddtracer "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

type DataDogTracerConfig struct {
	ServiceName string
	Host        string
	Port        string
	Env         string
	Version     string
}

type DataDogProvider struct {
	TracerConfig   *DataDogTracerConfig
	TracerProvider *ddotel.TracerProvider
}

// NewTracerProvider initializes the datadog Tracer with the provided start option
func (tp *DataDogProvider) NewTracerProvider() {
	agentAddr := fmt.Sprintf("%s:%s", tp.TracerConfig.Host, tp.TracerConfig.Port)
	startOption := []ddtracer.StartOption{ddtracer.WithEnv(tp.TracerConfig.Env),
		ddtracer.WithService(tp.TracerConfig.ServiceName),
	}
	if !(tp.TracerConfig.Host == "" && tp.TracerConfig.Port == "") {
		startOption = append(startOption, ddtracer.WithAgentAddr(agentAddr))
	}
	if tp.TracerConfig.Version != "" {
		startOption = append(startOption, ddtracer.WithServiceVersion(tp.TracerConfig.Version))
	}
	tp.TracerProvider = ddotel.NewTracerProvider(startOption...)
}
