package api

import (
	"database/sql"
	"errors"
	db "github.com/gaggudeep/bank_go/db/sqlc"
	"github.com/gaggudeep/bank_go/token"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"net/http"
)

type CreateAccountRequest struct {
	Currency string `json:"currency" binding:"required,currency"`
}

type GetAccountRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

type GetAccountsRequest struct {
	Page int32 `form:"page" binding:"min=1"`
	Size int32 `form:"page_size" binding:"required,min=1,max=100"`
}

func (server *Server) createAccount(ctx *gin.Context) {
	var req CreateAccountRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, parseErrorResp(err))
		return
	}

	authorizationPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)

	arg := db.CreateAccountParams{
		OwnerName: authorizationPayload.Username,
		Currency:  req.Currency,
		Balance:   "0",
	}

	acc, err := server.store.CreateAccount(ctx, arg)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "foreign_key_violation", "unique_violation":
				ctx.JSON(http.StatusForbidden, parseErrorResp(err))
				return
			}
		}

		ctx.JSON(http.StatusInternalServerError, parseErrorResp(err))
		return
	}

	ctx.JSON(http.StatusOK, acc)
}

func (server *Server) getAccount(ctx *gin.Context) {
	var req GetAccountRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, parseErrorResp(err))
		return
	}

	acc, err := server.store.GetAccount(ctx, req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, parseErrorResp(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, parseErrorResp(err))
		return
	}

	authorizationPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	if acc.OwnerName != authorizationPayload.Username {
		err := errors.New("account doesn't belong to authenticated user")
		ctx.JSON(http.StatusUnauthorized, parseErrorResp(err))
		return
	}

	ctx.JSON(http.StatusOK, acc)
}

func (server *Server) getAccounts(ctx *gin.Context) {
	var req GetAccountsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, parseErrorResp(err))
		return
	}

	authorizationPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	arg := db.GetAccountsParams{
		OwnerName: authorizationPayload.Username,
		Limit:     req.Size,
		Offset:    (req.Page - 1) * req.Size,
	}

	accounts, err := server.store.GetAccounts(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, parseErrorResp(err))
		return
	}

	ctx.JSON(http.StatusOK, accounts)
}
