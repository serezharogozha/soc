package transport

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/valyala/fasthttp"
	"io/ioutil"
	"net/http"
	"soc/pkg/domain"
	"soc/pkg/service"
	"soc/pkg/service/msgbroker"
	"strconv"
	"time"
)

type Server struct {
	router        *chi.Mux
	userService   service.UserService
	friendService service.FriendService
	postService   service.PostService
	db            *pgxpool.Pool
	broker        *msgbroker.MsgBroker
	wsHandler     *service.WsService
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
	db *pgxpool.Pool,
	service service.UserService,
	friendService service.FriendService,
	postService service.PostService,
	msgBroker *msgbroker.MsgBroker,
	wsHandler *service.WsService,
) Server {
	s := Server{}

	s.router = chi.NewRouter()
	s.db = db
	s.broker = msgBroker
	s.userService = service
	s.friendService = friendService
	s.postService = postService
	s.wsHandler = wsHandler

	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)

	s.router.Get("/user/get/{id}", s.GetUser)

	s.router.Post("/login", s.Login)
	s.router.Post("/user/register", s.CreateUser)
	s.router.Post("/user/search", s.UserSearch)

	s.router.With(AuthMiddleware).Post("/dialog/{user_id}/send", s.DialogSend)
	s.router.With(AuthMiddleware).Get("/dialog/{user_id}/list", s.DialogList)

	s.router.With(AuthMiddleware).Put("/friend/set/{id}", s.FriendSet)
	s.router.With(AuthMiddleware).Put("/friend/delete/{id}", s.FriendDelete)
	s.router.With(AuthMiddleware).Put("/post/get_feed", s.GetFeed)

	s.router.Post("/post/create", s.PostCreate)
	s.router.Put("/post/update", s.PostUpdate)

	s.router.With(AuthMiddleware).Get("/post/feed/posted", s.HandleWS)

	return s
}

func (s Server) HandleWS(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims, ok := ctx.Value("claims").(*UserClaims)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	userId := strconv.Itoa(claims.UserID)

	conn, err := service.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("failed upgrading connection"))
		return
	}
	defer conn.Close()
	clientID := uuid.New().String()

	s.wsHandler.ProcessMessage(conn, clientID, userId)
}

func (s Server) PostCreate(w http.ResponseWriter, r *http.Request) {
	post := &domain.Post{}
	ctx := r.Context()
	reqId := middleware.GetReqID(r.Context())

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&post); err != nil {
		errorResponse := ErrorResponse{
			Message:   "Failed to unmarshal post",
			RequestID: reqId,
			ErrorCode: http.StatusInternalServerError,
		}
		responseJson, _ := json.Marshal(errorResponse)

		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		_, writeErr := w.Write(responseJson)

		if writeErr != nil {
			return
		}
		return
	}

	postId, err := s.postService.CreatePost(ctx, *post)
	post.Id = postId
	if err != nil {
		errorResponse := ErrorResponse{
			Message:   "Failed to create post",
			RequestID: reqId,
			ErrorCode: http.StatusInternalServerError,
		}
		responseJson, _ := json.Marshal(errorResponse)

		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		_, writeErr := w.Write(responseJson)
		if writeErr != nil {
			return
		}
		return
	}

	go func() {
		messageBody, err := json.Marshal(post)
		if err != nil {
			return
		}

		err = s.broker.Publish("posts", string(messageBody))
		if err != nil {
			return
		}
	}()

	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte("Post created"))
	if err != nil {
		return
	}
}

func (s Server) PostUpdate(w http.ResponseWriter, r *http.Request) {
	post := &domain.Post{}

	ctx := r.Context()
	reqId := middleware.GetReqID(r.Context())

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&post); err != nil {
		errorResponse := ErrorResponse{
			Message:   "Failed to unmarshal post",
			RequestID: reqId,
			ErrorCode: http.StatusInternalServerError,
		}
		responseJson, _ := json.Marshal(errorResponse)

		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		_, writeErr := w.Write(responseJson)

		if writeErr != nil {
			return
		}
		return
	}

	err := s.postService.UpdatePost(ctx, *post)
	if err != nil {
		errorResponse := ErrorResponse{
			Message:   "Failed to update post",
			RequestID: reqId,
			ErrorCode: http.StatusInternalServerError,
		}

		responseJson, _ := json.Marshal(errorResponse)

		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		_, writeErr := w.Write(responseJson)
		if writeErr != nil {
			return
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte("Post updated"))
}

/*func (s Server) PostDelete(w http.ResponseWriter, r *http.Request) {

}

func (s Server) PostGet(w http.ResponseWriter, r *http.Request) {

}*/

func (s Server) GetFeed(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims, ok := ctx.Value("claims").(*UserClaims)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	userId := claims.UserID

	err, feed := s.postService.GetFeed(ctx, userId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	responseJson, _ := json.Marshal(feed)
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	_, writeErr := w.Write(responseJson)
	if writeErr != nil {
		return
	}
}

func (s Server) FriendSet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims, ok := ctx.Value("claims").(*UserClaims)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	userId := claims.UserID

	friendId := chi.URLParam(r, "id")
	friendIdInt, err := strconv.Atoi(friendId)

	err = s.friendService.SetFriend(ctx, userId, friendIdInt)
	if err != nil {
		errorResponse := ErrorResponse{
			Message:   "Failed to set friend",
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

func (s Server) FriendDelete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims, ok := ctx.Value("claims").(*UserClaims)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	userId := claims.UserID

	friendId := chi.URLParam(r, "id")
	friendIdInt, err := strconv.Atoi(friendId)

	err = s.friendService.DeleteFriend(ctx, userId, friendIdInt)
	if err != nil {
		errorResponse := ErrorResponse{
			Message:   "Failed to delete friend",
			RequestID: strconv.FormatUint(ctx.Value("request_id").(uint64), 10),
			ErrorCode: fasthttp.StatusInternalServerError,
		}
		responseJson, _ := json.Marshal(errorResponse)

		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		_, writeErr := w.Write(responseJson)
		if writeErr != nil {
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

func (s Server) GetUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := chi.URLParam(r, "id")
	reqId := middleware.GetReqID(r.Context())

	userIdInt, err := strconv.Atoi(userID)
	if err != nil {
		return
	}
	user, err := s.userService.GetUser(ctx, userIdInt)
	if err != nil {
		errorResponse := ErrorResponse{
			Message:   "Failed to fetch user details",
			RequestID: reqId,
			ErrorCode: http.StatusInternalServerError,
		}
		responseJson, _ := json.Marshal(errorResponse)

		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		_, writeErr := w.Write(responseJson)

		if writeErr != nil {
			return
		}

		return
	}

	userJson, _ := json.Marshal(user)

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	_, writeErr := w.Write(userJson)
	if writeErr != nil {
		return
	}
}

func (s Server) CreateUser(w http.ResponseWriter, r *http.Request) {
	user := &domain.User{}
	ctx := r.Context()
	reqId := middleware.GetReqID(r.Context())

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&user); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userId, err := s.userService.CreateUser(ctx, *user)
	if err != nil {
		errorResponse := ErrorResponse{
			Message:   "Internal server error",
			RequestID: reqId,
			ErrorCode: fasthttp.StatusInternalServerError,
		}
		responseJson, _ := json.Marshal(errorResponse)

		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		_, writeErr := w.Write(responseJson)

		if writeErr != nil {
			return
		}
		return
	}

	response := struct {
		UserId string `json:"user_id"`
	}{
		UserId: userId,
	}

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		errorResponse := ErrorResponse{
			Message:   "Internal server error",
			RequestID: reqId,
			ErrorCode: http.StatusInternalServerError,
		}
		responseJson, _ := json.Marshal(errorResponse)
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		_, writeErr := w.Write(responseJson)

		if writeErr != nil {
			return
		}
		return

	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, writeErr := w.Write(jsonResponse)
	if writeErr != nil {
		return
	}

	return
}

func (s Server) UserSearch(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userSearch := &domain.Search{}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&userSearch); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	users, err := s.userService.SearchUser(ctx, *userSearch)
	if err != nil {
		errorResponse := ErrorResponse{
			Message:   "Failed to fetch user details",
			RequestID: strconv.FormatUint(ctx.Value("request_id").(uint64), 10),
			ErrorCode: http.StatusInternalServerError,
		}
		responseJson, _ := json.Marshal(errorResponse)

		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		_, writeErr := w.Write(responseJson)
		if writeErr != nil {
			return
		}
		return
	}

	userJson, _ := json.Marshal(users)

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	_, writeErr := w.Write(userJson)
	if writeErr != nil {
		return
	}
	return
}

func (s Server) Login(w http.ResponseWriter, r *http.Request) {
	userLogin := &domain.Login{}
	ctx := r.Context()

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&userLogin); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	existedUser, err := s.userService.GetUser(ctx, userLogin.Id)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			return
		} else {
			errorResponse := ErrorResponse{
				Message:   "Failed to found user",
				RequestID: strconv.FormatUint(ctx.Value("request_id").(uint64), 10),
				ErrorCode: http.StatusInternalServerError,
			}
			responseJson, _ := json.Marshal(errorResponse)

			w.WriteHeader(http.StatusInternalServerError)
			w.Header().Set("Content-Type", "application/json")
			_, writeErr := w.Write(responseJson)
			if writeErr != nil {
				return
			}
			return
		}
	}

	err = service.CheckPassword(userLogin.Password, existedUser.Password)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	token, err := generateToken(existedUser.Id)
	if err != nil {
		errorResponse := ErrorResponse{
			Message:   "Failed to generate token",
			RequestID: strconv.FormatUint(ctx.Value("request_id").(uint64), 10),
			ErrorCode: http.StatusInternalServerError,
		}
		responseJson, _ := json.Marshal(errorResponse)

		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		_, writeErr := w.Write(responseJson)
		if writeErr != nil {
			return
		}
		return
	}

	w.Header().Set("Authorization", "Bearer "+token)

	response := struct {
		Token string `json:"token"`
	}{
		Token: token,
	}

	responseJson, err := json.Marshal(response)
	if err != nil {
		errorResponse := ErrorResponse{
			Message:   "Failed to encode response",
			RequestID: strconv.FormatUint(ctx.Value("request_id").(uint64), 10),
			ErrorCode: fasthttp.StatusInternalServerError,
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

	_, writeErr := w.Write(responseJson)
	if writeErr != nil {
		return
	}
	return
}

func (s Server) Start() error {
	fmt.Println("server started")
	return http.ListenAndServe(":8080", s.router)
}

func (s Server) DialogSend(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims, ok := ctx.Value("claims").(*UserClaims)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	reqId := middleware.GetReqID(ctx)
	userId := claims.UserID

	toUserId := chi.URLParam(r, "user_id")
	toUserIdInt, _ := strconv.Atoi(toUserId)

	dialogueText := &domain.DialogueMessage{}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&dialogueText); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	message := domain.DialogueMessage{
		UserID:   userId,
		ToUserID: toUserIdInt,
		Text:     dialogueText.Text,
	}

	messageBody, err := json.Marshal(message)
	if err != nil {
		errorResponse := ErrorResponse{
			Message:   "Failed to encode response",
			RequestID: reqId,
			ErrorCode: fasthttp.StatusInternalServerError,
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

	req, err := http.NewRequest("POST", "http://dialogues_app:8081/dialog/{user_id}/send", bytes.NewBuffer(messageBody))

	if err != nil {
		errorResponse := ErrorResponse{
			Message:   "Failed to encode response",
			RequestID: reqId,
			ErrorCode: fasthttp.StatusInternalServerError,
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
	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("X-Request-ID", reqId)

	// Send the HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		errorResponse := ErrorResponse{
			Message:   "Failed to encode response",
			RequestID: reqId,
			ErrorCode: fasthttp.StatusInternalServerError,
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
	defer resp.Body.Close()

	statusCodes := []int{http.StatusInternalServerError, http.StatusBadRequest}

	for _, code := range statusCodes {
		if resp.StatusCode == code {
			errorResponse := ErrorResponse{
				Message:   "Failed to perform request",
				RequestID: reqId,
				ErrorCode: resp.StatusCode,
			}
			responseJson, _ := json.Marshal(errorResponse)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(resp.StatusCode)
			_, errWrite := w.Write(responseJson)
			if errWrite != nil {
				return
			}

			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

func (s Server) DialogList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims, ok := ctx.Value("claims").(*UserClaims)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	userId := claims.UserID
	userIdStr := strconv.Itoa(userId)

	withUserId := chi.URLParam(r, "user_id")
	reqId := middleware.GetReqID(ctx)

	req, err := http.NewRequest("GET", "http://dialogues_app:8081/dialog/list/"+userIdStr+":"+withUserId, nil)
	req.Header.Add("X-Request-ID", reqId)
	if err != nil {
		errorResponse := ErrorResponse{
			Message:   "Failed to encode response",
			RequestID: reqId,
			ErrorCode: fasthttp.StatusInternalServerError,
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

	// Send the HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		errorResponse := ErrorResponse{
			Message:   "Failed to encode response",
			RequestID: reqId,
			ErrorCode: fasthttp.StatusInternalServerError,
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
	defer resp.Body.Close()

	statusCodes := []int{http.StatusInternalServerError, http.StatusBadRequest}

	for _, code := range statusCodes {
		if resp.StatusCode == code {
			errorResponse := ErrorResponse{
				Message:   "Failed to perform request",
				RequestID: reqId,
				ErrorCode: resp.StatusCode,
			}
			responseJson, _ := json.Marshal(errorResponse)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(resp.StatusCode)
			_, errWrite := w.Write(responseJson)
			if errWrite != nil {
				return
			}

			return
		}
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		errorResponse := ErrorResponse{
			Message:   "Failed to encode response",
			RequestID: reqId,
			ErrorCode: fasthttp.StatusInternalServerError,
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
	_, err = w.Write(body)
	if err != nil {
		return
	}
}
