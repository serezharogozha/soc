package transport

import (
	"awesomeProject10/pkg/domain"
	"awesomeProject10/pkg/service"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/buaazp/fasthttprouter"
	"github.com/dgrijalva/jwt-go"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/valyala/fasthttp"
	"strconv"
	"time"
)

type Server struct {
	router      *fasthttprouter.Router
	userService service.UserService
	db          *pgxpool.Pool
}

type ErrorResponse struct {
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
	ErrorCode int    `json:"code"`
}

var signingKey = []byte("my-secret-key")
var expirationTime = 15 * time.Minute

func NewServer(db *pgxpool.Pool, service service.UserService) Server {
	fmt.Println("NewServer")
	s := Server{}
	s.router = fasthttprouter.New()
	s.db = db
	s.userService = service

	s.router.GET("/user/get/:id", s.GetUser)
	s.router.POST("/login", s.Login)
	s.router.POST("/user/register", s.CreateUser)

	return s
}

func (s Server) GetUser(ctx *fasthttp.RequestCtx) {
	userID := ctx.UserValue("id").(string)
	userIdInt, err := strconv.Atoi(userID)
	if err != nil {
		return
	}
	user, err := s.userService.GetUser(ctx, userIdInt)
	if err != nil {
		errorResponse := ErrorResponse{
			Message:   "Failed to fetch user details",
			RequestID: strconv.FormatUint(ctx.ID(), 10),
			ErrorCode: fasthttp.StatusInternalServerError,
		}
		responseJson, _ := json.Marshal(errorResponse)
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.Write(responseJson)
		return
	}

	userJson, _ := json.Marshal(user)
	ctx.Response.Header.Set("Content-Type", "application/json")
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.Write(userJson)

}

func (s Server) CreateUser(ctx *fasthttp.RequestCtx) {
	user := &domain.User{}

	err := json.Unmarshal(ctx.PostBody(), user)
	if err != nil {
		fmt.Println(err)
		ctx.Error("Bad Request", fasthttp.StatusBadRequest)
		return
	}

	userId, err := s.userService.CreateUser(ctx, *user)
	if err != nil {
		errorResponse := ErrorResponse{
			Message:   "Internal server error",
			RequestID: strconv.FormatUint(ctx.ID(), 10),
			ErrorCode: fasthttp.StatusInternalServerError,
		}
		responseJson, _ := json.Marshal(errorResponse)
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.Write(responseJson)
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
			RequestID: strconv.FormatUint(ctx.ID(), 10),
			ErrorCode: fasthttp.StatusInternalServerError,
		}
		responseJson, _ := json.Marshal(errorResponse)
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.Write(responseJson)
		return
	}

	ctx.SetContentType("application/json")
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.Write(jsonResponse)
}

func (s Server) Login(ctx *fasthttp.RequestCtx) {
	fmt.Println("login")
	userLogin := &domain.Login{}

	err := json.Unmarshal(ctx.PostBody(), userLogin)
	if err != nil {
		ctx.Error("Bad Request", fasthttp.StatusBadRequest)
		return
	}

	existedUser, err := s.userService.GetUser(ctx, userLogin.Id)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.Error("User not found", fasthttp.StatusNotFound)
			return
		} else {
			errorResponse := ErrorResponse{
				Message:   "Failed to found user",
				RequestID: strconv.FormatUint(ctx.ID(), 10),
				ErrorCode: fasthttp.StatusInternalServerError,
			}
			responseJson, _ := json.Marshal(errorResponse)
			ctx.Response.Header.Set("Content-Type", "application/json")
			ctx.SetStatusCode(fasthttp.StatusInternalServerError)
			ctx.Write(responseJson)
			return
		}
	}

	// Check if the password is correct
	if userLogin.Password != existedUser.Password {
		ctx.Error("Unauthorized", fasthttp.StatusUnauthorized)
		return
	}

	token, err := generateToken(existedUser.Id)
	if err != nil {
		errorResponse := ErrorResponse{
			Message:   "Failed to generate token",
			RequestID: strconv.FormatUint(ctx.ID(), 10),
			ErrorCode: fasthttp.StatusInternalServerError,
		}
		responseJson, _ := json.Marshal(errorResponse)
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.Write(responseJson)
		return
	}

	ctx.Response.Header.Set("Authorization", "Bearer "+token)

	response := struct {
		Token string `json:"token"`
	}{
		Token: token,
	}

	responseJson, err := json.Marshal(response)
	if err != nil {
		errorResponse := ErrorResponse{
			Message:   "Failed to encode response",
			RequestID: strconv.FormatUint(ctx.ID(), 10),
			ErrorCode: fasthttp.StatusInternalServerError,
		}
		responseJson, _ := json.Marshal(errorResponse)
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.Write(responseJson)
		return
	}

	ctx.SetContentType("application/json")
	ctx.SetStatusCode(fasthttp.StatusOK)

	ctx.Write(responseJson)
}

func generateToken(username int) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["username"] = username
	claims["exp"] = time.Now().Add(expirationTime).Unix()

	signedToken, err := token.SignedString(signingKey)
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

func (s Server) Start() error {
	fmt.Println("server started")
	return fasthttp.ListenAndServe(":8080", s.router.Handler)
}
