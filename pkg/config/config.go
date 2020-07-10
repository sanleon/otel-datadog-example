package config

import (
	"fmt"
	"os"

	"github.com/DataDog/datadog-go/statsd"
	"go.opentelemetry.io/contrib/exporters/metric/datadog"
)

func InitDatadogExporter(env string) (*datadog.Exporter, error) {
	datadogOpts := datadog.Options{
		StatsDOptions: []statsd.Option{statsd.WithoutTelemetry()},
		Tags:          []string{"env:" + env},
	}

	statsdHostIP := os.Getenv("DOGSTATSD_HOST_IP")
	if statsdHostIP != "" {
		datadogOpts.StatsAddr = fmt.Sprintf("%s:8125", statsdHostIP)
	}

	exporter, err := datadog.NewExporter(datadogOpts)
	if err != nil {
		return nil, err
	}

	return exporter, nil
}
