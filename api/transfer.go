package api

import (
	"database/sql"
	"fmt"
	db "github.com/gaggudeep/bank_go/db/sqlc"
	"github.com/gin-gonic/gin"
	"net/http"
)

type TransferRequest struct {
	FromAccountID int64  `json:"from_account_id" binding:"required,min=1"`
	ToAccountID   int64  `json:"to_account_id" binding:"required,min=1"`
	Amount        string `json:"amount" binding:"required,amount"`
	Currency      string `json:"currency" binding:"required,currency"`
}

func (server *Server) Transfer(ctx *gin.Context) {
	var req TransferRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, parseErrorResp(err))
		return
	}

	if !server.validAccount(ctx, req.FromAccountID, req.Currency) ||
		!server.validAccount(ctx, req.ToAccountID, req.Currency) {
		return
	}

	arg := db.TransferTxParams{
		FromAccountID: req.FromAccountID,
		ToAccountID:   req.ToAccountID,
		Amount:        req.Amount,
	}

	res, err := server.store.TransferTxPreventingCircularWait(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, parseErrorResp(err))
		return
	}

	ctx.JSON(http.StatusOK, res)
}

func (server *Server) validAccount(ctx *gin.Context, accId int64, currency string) bool {
	acc, err := server.store.GetAccount(ctx, accId)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, parseErrorResp(err))
			return false
		}

		ctx.JSON(http.StatusInternalServerError, parseErrorResp(err))
		return false
	}

	if acc.Currency != currency {
		err := fmt.Errorf("account [%d] currency mismatch, request: %s DB: %s",
			accId, currency, acc.Currency)
		ctx.JSON(http.StatusBadRequest, parseErrorResp(err))
		return false
	}

	return true
}
