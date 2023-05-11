package gapi

import (
	"fmt"
	db "github.com/gaggudeep/bank_go/db/sqlc"
	"github.com/gaggudeep/bank_go/pb"
	"github.com/gaggudeep/bank_go/token"
	"github.com/gaggudeep/bank_go/util"
	"github.com/gin-gonic/gin"
)

type Server struct {
	pb.UnimplementedBankServer
	config     util.Config
	store      db.Store
	tokenMaker token.Maker
	router     *gin.Engine
}

func NewServer(store db.Store, config *util.Config) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}
	server := &Server{
		config:     *config,
		store:      store,
		tokenMaker: tokenMaker,
	}

	return server, nil
}
