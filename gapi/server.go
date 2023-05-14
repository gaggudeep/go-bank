package gapi

import (
	"fmt"
	db "github.com/gaggudeep/bank_go/db/sqlc"
	"github.com/gaggudeep/bank_go/pb"
	"github.com/gaggudeep/bank_go/token"
	"github.com/gaggudeep/bank_go/util"
	"github.com/gaggudeep/bank_go/worker"
	"github.com/gin-gonic/gin"
)

type Server struct {
	pb.UnimplementedBankServer
	config          util.Config
	store           db.Store
	tokenMaker      token.Maker
	router          *gin.Engine
	taskDistributor worker.TaskDistributor
}

func NewServer(store db.Store, config *util.Config, distributor worker.TaskDistributor) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %v", err)
	}

	server := &Server{
		config:          *config,
		store:           store,
		tokenMaker:      tokenMaker,
		taskDistributor: distributor,
	}

	return server, nil
}
