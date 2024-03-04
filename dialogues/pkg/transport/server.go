package transport

import (
	"dialogues/pkg/domain"
	"dialogues/pkg/service"
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/go-chi/chi"
	"net/http"
	"strconv"
	"time"
)

type Server struct {
	router          *chi.Mux
	dialogueService service.DialogueService
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

var SigningKey = []byte("my-secret-key")
var ExpirationTime = 15 * time.Minute

func NewServer(
	dialogueService service.DialogueService,
) Server {
	s := Server{}

	s.router = chi.NewRouter()
	s.dialogueService = dialogueService

	s.router.Post("/dialog/{user_id}/send", s.DialogSend)
	s.router.Get("/dialog/list/{user_id}:{withUserId}", s.DialogList)

	return s
}

func (s Server) Start() error {
	fmt.Println("server started")
	return http.ListenAndServe(":8081", s.router)
}

func (s Server) DialogSend(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
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
			RequestID: strconv.FormatUint(ctx.Value("request_id").(uint64), 10),
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

	w.WriteHeader(http.StatusOK)
}

func (s Server) DialogList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userId := chi.URLParam(r, "user_id")
	withUserId := chi.URLParam(r, "withUserId")

	userIdSrt, _ := strconv.Atoi(userId)
	withUserIdInt, _ := strconv.Atoi(withUserId)

	dialogues, err := s.dialogueService.GetDialogue(userIdSrt, withUserIdInt)
	if err != nil {
		errorResponse := ErrorResponse{
			Message:   "Failed to get dialogue",
			RequestID: strconv.FormatUint(ctx.Value("request_id").(uint64), 10),
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
			RequestID: strconv.FormatUint(ctx.Value("request_id").(uint64), 10),
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

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	_, writeErr := w.Write(dialoguesJson)
	if writeErr != nil {
		return
	}
	return
}
