package api

import (
	db "github.com/gaggudeep/bank_go/db/sqlc"
	"github.com/gaggudeep/bank_go/util"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	validator2 "github.com/go-playground/validator/v10"
	"log"
)

type Server struct {
	store  db.Store
	router *gin.Engine
}

func NewServer(store db.Store, config *util.Config) *Server {
	server := &Server{store: store}
	router := gin.Default()

	for _, validator := range config.CustomValidators {
		if v, ok := binding.Validator.Engine().(*validator2.Validate); ok {
			err := v.RegisterValidation(validator.Name, validator.Func)
			if err != nil {
				log.Fatal("cannot register validator: ", err)
			}
		}
	}

	router.POST("/users", server.createUser)

	router.POST("/accounts", server.createAccount)
	router.GET("/accounts/:id", server.getAccount)
	router.GET("/accounts", server.getAccounts)

	router.POST("/transfers", server.createTransfer)

	server.router = router
	return server
}

func (server *Server) Start(addr string) error {
	return server.router.Run(addr)
}

func parseErrorResp(err error) gin.H {
	return gin.H{"error": err.Error()}
}
