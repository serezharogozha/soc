package transport

import (
	"counter/pkg/service"
	"counter/pkg/service/msgbroker"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"strconv"
)

type Server struct {
	router         *chi.Mux
	counterService service.CounterService
	broker         *msgbroker.MsgBroker
	metrics        *Metrics
}

type ErrorResponse struct {
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
	ErrorCode int    `json:"code"`
}

type UserClaims struct {
	UserID int `json:"user_id"`
	jwt.StandardClaims
}

func NewServer(
	counterService service.CounterService,
	broker *msgbroker.MsgBroker,
	metrics *Metrics,
) Server {
	s := Server{}

	s.router = chi.NewRouter()
	s.counterService = counterService
	s.broker = broker
	s.metrics = metrics

	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)
	s.router.Use(metrics.CommonMetricsMiddleware)

	s.router.Handle("/metrics", promhttp.Handler())

	s.router.Handle("/metrics", promhttp.Handler())

	s.router.Get("/counter/{userId}:{withUserId}", s.GetCounter)

	return s
}

func (s Server) GetCounter(w http.ResponseWriter, r *http.Request) {
	//ctx := r.Context()
	userId, err := strconv.Atoi(chi.URLParam(r, "userId"))
	if err != nil {
		http.Error(w, "Invalid user id", http.StatusBadRequest)
		return
	}

	withUserId, err := strconv.Atoi(chi.URLParam(r, "withUserId"))
	if err != nil {
		http.Error(w, "Invalid with user id", http.StatusBadRequest)
		return
	}

	counter, err := s.counterService.GetCounter(userId, withUserId)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Write([]byte(strconv.FormatUint(counter, 10)))
}

func (s Server) Start() error {
	fmt.Println("server started")
	return http.ListenAndServe(":8082", s.router)
}
