package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/instrumentation/othttp"
	"go.opentelemetry.io/otel/sdk/metric/aggregator/ddsketch"
	"go.opentelemetry.io/otel/sdk/metric/controller/push"
	"go.opentelemetry.io/otel/sdk/metric/selector/simple"

	"github.com/sanleon/otel-datadog-example/pkg/config"
	"github.com/sanleon/otel-datadog-example/pkg/handler"
)

const MeterName ="example.io/test"

func main() {

	ctx := context.Background()

	// Init datadog
	exp, err := config.InitDatadogExporter("local")
	if err != nil {
		panic(err)
	}

	// Set pusher for datadog metrics
	selector := simple.NewWithSketchDistribution(ddsketch.NewDefaultConfig())
	// Push collected data to datadog with period 10 second
	pusher := push.New(selector, exp, push.WithPeriod(time.Second*10))
	global.SetMeterProvider(pusher.Provider())
	defer pusher.Stop()

	go func() {
		pusher.Start()
	}()

	fmt.Println("Start Server")

	mux := http.NewServeMux()
	mux.HandleFunc("/", DefaultHandler)

	customMetricsHandler := handler.NewMetricsHandler(mux, global.Meter(MeterName))
	hs := &http.Server{
		Addr: fmt.Sprintf(":%s", "8080"),
		Handler: othttp.NewHandler(customMetricsHandler, "server",
			othttp.WithMeter(global.Meter(MeterName)),
			othttp.WithMessageEvents(othttp.ReadEvents, othttp.WriteEvents),
		),
	}

	defer func() {
		tctx, cancel := context.WithTimeout(context.Background(), time.Second * 30)

		defer cancel()
		if err := hs.Shutdown(tctx); err != nil {
			fmt.Printf("error: %v", err)
		}
	}()

	errCh := make(chan error, 1)
	go func() {
		if err := hs.ListenAndServe(); err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	<-ctx.Done()
}

func DefaultHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "default")
}


