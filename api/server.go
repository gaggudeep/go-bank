package api

import (
	"fmt"
	db "github.com/gaggudeep/bank_go/db/sqlc"
	"github.com/gaggudeep/bank_go/token"
	"github.com/gaggudeep/bank_go/util"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	validator2 "github.com/go-playground/validator/v10"
	"log"
)

type Server struct {
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

	server.setupValidators()
	server.setupRouter()

	return server, nil
}

func (server *Server) setupValidators() {
	for _, validator := range server.config.CustomValidators {
		if v, ok := binding.Validator.Engine().(*validator2.Validate); ok {
			err := v.RegisterValidation(validator.Name, validator.Func)
			if err != nil {
				log.Fatal("cannot register validator: ", err)
			}
		}
	}
}

func (server *Server) setupRouter() {
	router := gin.Default()

	router.POST("/users", server.createUser)
	router.POST("/users/login", server.loginUser)
	router.POST("/tokens/renew_access", server.renewAccessToken)

	authRoutes := router.Group("/").Use(authMiddleware(server.tokenMaker))

	authRoutes.POST("/accounts", server.createAccount)
	authRoutes.GET("/accounts/:id", server.getAccount)
	authRoutes.GET("/accounts", server.getAccounts)

	authRoutes.POST("/transfers", server.Transfer)

	server.router = router
}

func (server *Server) Start(addr string) error {
	return server.router.Run(addr)
}

func parseErrorResp(err error) gin.H {
	return gin.H{"error": err.Error()}
}
