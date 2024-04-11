package transport

import (
	"dialogues/pkg/domain"
	"dialogues/pkg/service"
	"dialogues/pkg/service/msgbroker"
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"strconv"
)

type Server struct {
	router          *chi.Mux
	dialogueService service.DialogueService
	broker          *msgbroker.MsgBroker
	metrics         *Metrics
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
	dialogueService service.DialogueService,
	broker *msgbroker.MsgBroker,
	metrics *Metrics,
) Server {
	s := Server{}
	s.dialogueService = dialogueService
	s.broker = broker

	s.router = chi.NewRouter()
	s.dialogueService = dialogueService
	s.metrics = metrics

	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)
	s.router.Use(metrics.CommonMetricsMiddleware)

	s.router.Handle("/metrics", promhttp.Handler())

	s.router.Post("/dialog/{user_id}/send", s.DialogSend)
	s.router.Get("/dialog/list/{user_id}:{withUserId}", s.DialogList)

	s.router.Handle("/metrics", promhttp.Handler())

	return s
}

func (s Server) Start() error {
	fmt.Println("server started")
	return http.ListenAndServe(":8081", s.router)
}

func (s Server) DialogSend(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	reqId := middleware.GetReqID(ctx)
	dialogue := &domain.DialogueMessage{}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&dialogue); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err := s.dialogueService.CreateMessages(dialogue.UserID, dialogue.ToUserID, dialogue.Text)

	if err != nil {
		errorResponse := ErrorResponse{
			Message:   "Failed to create dialogue",
			RequestID: reqId,
			ErrorCode: http.StatusInternalServerError,
		}
		responseJson, _ := json.Marshal(errorResponse)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, errWrite := w.Write(responseJson)
		if errWrite != nil {
			return
		}
		return
	}

	go func() {
		messageSend := domain.MessageSendBroker{
			From: dialogue.UserID,
			To:   dialogue.ToUserID,
		}

		messageBody, err := json.Marshal(messageSend)
		if err != nil {
			return
		}

		err = s.broker.Publish("message_send", string(messageBody))
		if err != nil {
			return
		}
	}()

	w.WriteHeader(http.StatusOK)
}

func (s Server) DialogList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqId := middleware.GetReqID(ctx)
	userId := chi.URLParam(r, "user_id")
	withUserId := chi.URLParam(r, "withUserId")

	userIdSrt, _ := strconv.Atoi(userId)
	withUserIdInt, _ := strconv.Atoi(withUserId)

	dialogues, counter, err := s.dialogueService.GetDialogue(userIdSrt, withUserIdInt)
	if err != nil {
		errorResponse := ErrorResponse{
			Message:   "Failed to get dialogue",
			RequestID: reqId,
			ErrorCode: http.StatusInternalServerError,
		}
		responseJson, _ := json.Marshal(errorResponse)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, errWrite := w.Write(responseJson)
		if errWrite != nil {
			return
		}
		return
	}

	dialoguesJson, err := json.Marshal(dialogues)
	if err != nil {
		errorResponse := ErrorResponse{
			Message:   "Failed to marshal dialogues",
			RequestID: reqId,
			ErrorCode: http.StatusInternalServerError,
		}
		responseJson, _ := json.Marshal(errorResponse)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, errWrite := w.Write(responseJson)
		if errWrite != nil {
			return
		}
		return
	}

	go func() {
		messagesRead := domain.MessageReadBroker{
			From:        userIdSrt,
			To:          withUserIdInt,
			ReadCounter: counter,
		}
		messageBody, err := json.Marshal(messagesRead)
		if err != nil {
			return
		}

		err = s.broker.Publish("message_read", string(messageBody))
		if err != nil {
			return
		}
	}()

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	_, writeErr := w.Write(dialoguesJson)
	if writeErr != nil {
		return
	}
	return
}
