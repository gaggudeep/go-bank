package api

import (
	"errors"
	"github.com/gaggudeep/bank_go/token"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

const (
	authorizationHeaderKey    = "authorization"
	authorizationSchemeBearer = "bearer"
	authorizationPayloadKey   = "authorization_payload"
)

func authMiddleware(maker token.Maker) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authorizationHeader := ctx.GetHeader(authorizationHeaderKey)
		if len(authorizationHeader) == 0 {
			err := errors.New("authorization header is not provided")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, parseErrorResp(err))
			return
		}

		fields := strings.Fields(authorizationHeader)
		if len(fields) < 2 {
			err := errors.New("invalid authorization header format")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, parseErrorResp(err))
			return
		}

		authorizationScheme := strings.ToLower(fields[0])
		if authorizationScheme != authorizationSchemeBearer {
			err := errors.New("unsupported authorization scheme")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, parseErrorResp(err))
			return
		}

		token := fields[1]
		payload, err := maker.VerifyToken(token)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, parseErrorResp(err))
			return
		}

		ctx.Set(authorizationPayloadKey, payload)
		ctx.Next()
	}
}
