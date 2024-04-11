package transport

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog"
	"net/http"
	"regexp"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// Metrics - metrics type
type Metrics struct {
	reqDuration *prometheus.HistogramVec
	reqTotal    *prometheus.CounterVec
	log         *zerolog.Logger
}

// InitMetrics - metrics initialize
func InitMetrics(logger *zerolog.Logger) *Metrics {
	return &Metrics{
		reqDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Help:    "request execution time in milliseconds",
				Name:    "counter_http_req_duration_ms",
				Buckets: []float64{1, 3, 5, 7, 9, 12, 15},
			},
			[]string{"endpoint"},
		),
		log: logger,
		reqTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Help: "total requests to http api",
			Name: "counter_http_req_count",
		}, []string{"endpoint", "status"}),
	}
}

func (m *Metrics) CommonMetricsMiddleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		method := strings.ToLower(r.Method)

		path := r.URL.Path

		var re = regexp.MustCompile(`/(\d+)`)
		path = re.ReplaceAllString(path, `/id`)
		path = strings.ReplaceAll(path, "/", "_")

		label := fmt.Sprintf("%s%s", method, path)

		defer func() {
			if r := recover(); r != nil {
				m.log.Error().Str("panic", "true").Str("stacktrace", string(debug.Stack())).Str("endpoint", label).Msg(fmt.Sprintf("%s", r))
				w.WriteHeader(http.StatusInternalServerError)
			}
		}()

		start := time.Now()
		next.ServeHTTP(rw, r)

		m.reqDuration.WithLabelValues(label).Observe(float64(time.Since(start).Milliseconds()))
		statusCode := rw.statusCode
		m.reqTotal.WithLabelValues(label, strconv.FormatInt(int64(statusCode), 10)).Inc()
	}

	return http.HandlerFunc(fn)
}
