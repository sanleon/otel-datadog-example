package handler

import (
	"fmt"
	"net/http"

	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/api/kv"
	"go.opentelemetry.io/otel/api/metric"
	"go.opentelemetry.io/otel/api/standard"
)

// Server HTTP Custom metrics
const (
	RequestCount          = "http.server.request_count"
	RequestCountByMethod  = "http.server.request_count_by_method"
	ResponseCountByStatus = "http.server.response_count_by_status_code"
)

type metricsHandler struct {
	handler        http.Handler
	meter          metric.Meter
	counters       map[string]metric.Int64Counter
	valueRecorders map[string]metric.Int64ValueRecorder
}

func NewMetricsHandler(handler http.Handler, meter metric.Meter) http.Handler {
	m := &metricsHandler{
		handler: handler,
		meter:   meter,
	}
	m.createMeasures()

	return m
}

func (h *metricsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	labels := h.httpBasicAttributesFromHTTPRequest(r)

	// Add request count for metrics
	h.counters[RequestCount].Add(ctx, int64(1), labels...)

	labels  = append(labels, kv.String("http.method", r.Method))
	h.counters[RequestCountByMethod].Add(ctx, int64(1), labels...)

	rww := &respWriterWrapper{ResponseWriter: w, ctx: ctx}

	h.handler.ServeHTTP(rww, r.WithContext(ctx))

	// Add response count per status code for metrics
	labels = append(labels, kv.Int("http.status", rww.statusCode))

	// Add error content to labels if have error on response.
	if rww.err != nil {
		labels = append(labels, kv.String("http.error", rww.err.Error()))
	}
	h.counters[ResponseCountByStatus].Add(ctx, int64(1), labels...)

}

func (h *metricsHandler) createMeasures() {
	h.counters = make(map[string]metric.Int64Counter)

	requestCounter, err := h.meter.NewInt64Counter(RequestCount)
	if err != nil {
		global.Handle(err)
	}

	responseCounter, err := h.meter.NewInt64Counter(ResponseCountByStatus)
	if err != nil {
		global.Handle(err)
	}

	requestCounterByMethod, err := h.meter.NewInt64Counter(RequestCountByMethod)
	if err != nil {
		global.Handle(err)
	}

	h.counters[RequestCount] = requestCounter
	h.counters[RequestCountByMethod] = requestCounterByMethod
	h.counters[ResponseCountByStatus] = responseCounter
}

func (h *metricsHandler) httpBasicAttributesFromHTTPRequest(request *http.Request) []kv.KeyValue {
	var attrs []kv.KeyValue

	if request.TLS != nil {
		attrs = append(attrs, standard.HTTPSchemeHTTPS)
	} else {
		attrs = append(attrs, standard.HTTPSchemeHTTP)
	}

	if request.Host != "" {
		attrs = append(attrs, standard.HTTPHostKey.String(request.Host))
	}

	flavor := ""
	if request.ProtoMajor == 1 {
		flavor = fmt.Sprintf("1.%d", request.ProtoMinor)
	} else if request.ProtoMajor == 2 {
		flavor = "2"
	}
	if flavor != "" {
		attrs = append(attrs, standard.HTTPFlavorKey.String(flavor))
	}

	return attrs
}




