package api

import (
	"database/sql"
	db "github.com/gaggudeep/bank_go/db/sqlc"
	"github.com/gaggudeep/bank_go/util"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"net/http"
	"time"
)

type CreateUserRequest struct {
	Username string `json:"username" binding:"required,alphanum"`
	Password string `json:"password" binding:"required,min=6"`
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
}

type UserResponse struct {
	Username          string    `json:"username"`
	Name              string    `json:"name"`
	Email             string    `json:"email"`
	PasswordChangedAt time.Time `json:"password_changed_at"`
	CreatedAt         time.Time `json:"created_at"`
}

func (server *Server) createUser(ctx *gin.Context) {
	var req CreateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, parseErrorResp(err))
		return
	}

	hashedPwd, err := util.HashPassword(req.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, parseErrorResp(err))
		return
	}

	arg := db.CreateUserParams{
		Username:       req.Username,
		HashedPassword: hashedPwd,
		Name:           req.Name,
		Email:          req.Email,
	}

	user, err := server.store.CreateUser(ctx, arg)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "unique_violation":
				ctx.JSON(http.StatusForbidden, parseErrorResp(err))
				return
			}
		}

		ctx.JSON(http.StatusInternalServerError, parseErrorResp(err))
		return
	}

	resp := newUserResponse(&user)

	ctx.JSON(http.StatusOK, resp)
}

type LoginRequest struct {
	Username string `json:"username" binding:"required,alphanum"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginResponse struct {
	AccessToken string       `json:"access_token"`
	User        UserResponse `json:"user"`
}

func (server *Server) loginUser(ctx *gin.Context) {
	var req LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, parseErrorResp(err))
		return
	}

	user, err := server.store.GetUser(ctx, req.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, parseErrorResp(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, parseErrorResp(err))
		return
	}

	err = util.ValidatePassword(req.Password, user.HashedPassword)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, parseErrorResp(err))
		return
	}

	accessToken, err := server.tokenMaker.CreateToken(req.Username,
		server.config.SecurityConfig.TokenConfig.AccessDuration)

	resp := LoginResponse{
		AccessToken: accessToken,
		User:        newUserResponse(&user),
	}

	ctx.JSON(http.StatusOK, resp)
}

func newUserResponse(user *db.User) UserResponse {
	return UserResponse{
		Username:          user.Username,
		Name:              user.Name,
		Email:             user.Email,
		PasswordChangedAt: user.PasswordChangedAt,
		CreatedAt:         user.CreatedAt,
	}
}
