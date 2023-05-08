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
	router        *fasthttprouter.Router
	userService   service.UserService
	friendService service.FriendService
	postService   service.PostService
	db            *pgxpool.Pool
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

var signingKey = []byte("my-secret-key")
var expirationTime = 15 * time.Minute

func NewServer(db *pgxpool.Pool, service service.UserService, friendService service.FriendService, postService service.PostService) Server {
	fmt.Println("NewServer")
	s := Server{}
	s.router = fasthttprouter.New()
	s.db = db
	s.userService = service
	s.friendService = friendService
	s.postService = postService

	s.router.GET("/user/get/:id", s.GetUser)
	s.router.POST("/login", s.Login)
	s.router.POST("/user/register", s.CreateUser)
	s.router.POST("/user/search", s.UserSearch)

	s.router.PUT("/friend/set/:id", AuthMiddleware(s.FriendSet))
	s.router.PUT("/friend/delete/:id", AuthMiddleware(s.FriendDelete))

	s.router.POST("/post/create", s.PostCreate)
	s.router.PUT("/post/update", s.PostUpdate)
	s.router.PUT("/post/delete", s.PostDelete)
	s.router.GET("/post/get", s.PostGet)
	s.router.GET("/post/get_feed", AuthMiddleware(s.GetFeed))

	return s
}

func (s Server) PostCreate(ctx *fasthttp.RequestCtx) {
	post := &domain.Post{}

	err := json.Unmarshal(ctx.PostBody(), &post)
	if err != nil {
		fmt.Println(err)
		errorResponse := ErrorResponse{
			Message:   "Failed to unmarshal post",
			RequestID: strconv.FormatUint(ctx.ID(), 10),
			ErrorCode: fasthttp.StatusInternalServerError,
		}
		responseJson, _ := json.Marshal(errorResponse)
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetBody(responseJson)
		return
	}

	err = s.postService.CreatePost(ctx, *post)
	if err != nil {
		fmt.Println(err)
		errorResponse := ErrorResponse{
			Message:   "Failed to create post",
			RequestID: strconv.FormatUint(ctx.ID(), 10),
			ErrorCode: fasthttp.StatusInternalServerError,
		}
		responseJson, _ := json.Marshal(errorResponse)
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetBody(responseJson)
		return
	}

	ctx.SetStatusCode(fasthttp.StatusOK)
}

func (s Server) PostUpdate(ctx *fasthttp.RequestCtx) {
	post := &domain.Post{}

	err := json.Unmarshal(ctx.PostBody(), &post)
	if err != nil {
		fmt.Println(err)
		errorResponse := ErrorResponse{
			Message:   "Failed to unmarshal post",
			RequestID: strconv.FormatUint(ctx.ID(), 10),
			ErrorCode: fasthttp.StatusInternalServerError,
		}
		responseJson, _ := json.Marshal(errorResponse)
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetBody(responseJson)
		return
	}

	err = s.postService.UpdatePost(ctx, *post)
	if err != nil {
		fmt.Println(err)
		errorResponse := ErrorResponse{
			Message:   "Failed to update post",
			RequestID: strconv.FormatUint(ctx.ID(), 10),
			ErrorCode: fasthttp.StatusInternalServerError,
		}
		responseJson, _ := json.Marshal(errorResponse)
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.SetBody(responseJson)
		return
	}

	ctx.SetStatusCode(fasthttp.StatusOK)
}

func (s Server) PostDelete(ctx *fasthttp.RequestCtx) {

}

func (s Server) PostGet(ctx *fasthttp.RequestCtx) {

}

func (s Server) GetFeed(ctx *fasthttp.RequestCtx) {
	claims, ok := ctx.UserValue("claims").(*UserClaims)
	if !ok {
		fmt.Println("Failed to get claims")
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		return
	}

	userId := claims.UserID

	err, feed := s.postService.GetFeed(ctx, userId)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		return
	}

	responseJson, _ := json.Marshal(feed)
	ctx.Response.Header.Set("Content-Type", "application/json")
	ctx.Write(responseJson)
	ctx.SetStatusCode(fasthttp.StatusOK)
}

func (s Server) FriendSet(ctx *fasthttp.RequestCtx) {
	claims, ok := ctx.UserValue("claims").(*UserClaims)
	if !ok {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		return
	}

	userId := claims.UserID

	friendId := ctx.UserValue("id").(string)
	fmt.Println("friendId: ", friendId)
	friendIdInt, err := strconv.Atoi(friendId)

	err = s.friendService.SetFriend(ctx, userId, friendIdInt)
	if err != nil {
		fmt.Println(err)
		errorResponse := ErrorResponse{
			Message:   "Failed to set friend",
			RequestID: strconv.FormatUint(ctx.ID(), 10),
			ErrorCode: fasthttp.StatusInternalServerError,
		}
		responseJson, _ := json.Marshal(errorResponse)
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.Write(responseJson)
		return
	}

	ctx.SetStatusCode(fasthttp.StatusOK)
}

func (s Server) FriendDelete(ctx *fasthttp.RequestCtx) {
	claims, ok := ctx.UserValue("claims").(*UserClaims)
	if !ok {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		return
	}

	userId := claims.UserID

	friendId := ctx.UserValue("id").(string)
	friendIdInt, err := strconv.Atoi(friendId)

	err = s.friendService.DeleteFriend(ctx, userId, friendIdInt)
	if err != nil {
		errorResponse := ErrorResponse{
			Message:   "Failed to delete friend",
			RequestID: strconv.FormatUint(ctx.ID(), 10),
			ErrorCode: fasthttp.StatusInternalServerError,
		}
		responseJson, _ := json.Marshal(errorResponse)
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.Write(responseJson)
		return
	}

	ctx.SetStatusCode(fasthttp.StatusOK)
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

func (s Server) UserSearch(ctx *fasthttp.RequestCtx) {
	userSearch := &domain.Search{}
	err := json.Unmarshal(ctx.PostBody(), userSearch)
	if err != nil {
		fmt.Println(err)
		ctx.Error("Bad Request", fasthttp.StatusBadRequest)
		return
	}

	users, err := s.userService.SearchUser(ctx, *userSearch)
	if err != nil {
		fmt.Println(err)
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

	userJson, _ := json.Marshal(users)
	ctx.Response.Header.Set("Content-Type", "application/json")
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.Write(userJson)
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

	err = service.CheckPassword(userLogin.Password, existedUser.Password)
	if err != nil {
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
	claims["user_id"] = username
	claims["exp"] = time.Now().Add(expirationTime).Unix()

	signedToken, err := token.SignedString(signingKey)
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

func AuthMiddleware(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		authHeader := string(ctx.Request.Header.Peek("Authorization"))
		if authHeader == "" {
			fmt.Println("no auth header")
			ctx.SetStatusCode(fasthttp.StatusUnauthorized)
			return
		}

		tokenString := authHeader[7:]
		token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return signingKey, nil
		})

		if err != nil {
			fmt.Println(err)
			ctx.SetStatusCode(fasthttp.StatusUnauthorized)
			return
		}

		if !token.Valid {
			fmt.Println(err)
			ctx.SetStatusCode(fasthttp.StatusUnauthorized)
			return
		}
		fmt.Println("get claims")
		claims := token.Claims.(*UserClaims)
		ctx.SetUserValue("claims", claims)
		fmt.Println("next")
		next(ctx)
	}
}

func (s Server) Start() error {
	fmt.Println("server started")
	return fasthttp.ListenAndServe(":8080", s.router.Handler)
}
