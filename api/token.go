package api

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type renewAccessTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type renewAccessTokenResponse struct {
	AccessToken          string    `json:"access_token"`
	AccessTokenExpiresAt time.Time `json:"access_token_expires_at"`
}

func (server *Server) renewAccessToken(ctx *gin.Context) {
	var req renewAccessTokenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		log.Printf("error binding JSON: %v\n", err)
		return
	}

	refreshTokenPayload, err := server.tokenMaker.VerifyToken(req.RefreshToken)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		log.Printf("error verifying refresh token: %v\n", err)
		return
	}

	session, err := server.store.GetSession(ctx, refreshTokenPayload.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			log.Printf("session not found: %v\n", err)
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		log.Printf("error getting session: %v\n", err)
		return
	}

	log.Printf("found user with id: %v\n", session.Username)

	if session.IsBlocked {
		err := errors.New("blocked session")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		log.Printf("blocked session: %v\n", err)
		return
	}

	accessToken, accessPayload, err := server.tokenMaker.CreateToken(
		refreshTokenPayload.Username,
		server.config.AccessTokenDuration,
	)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		log.Printf("error creating access token: %v\n", err)
		return
	}

	rsp := renewAccessTokenResponse{
		AccessToken:          accessToken,
		AccessTokenExpiresAt: accessPayload.ExpiredAt,
	}

	ctx.JSON(http.StatusOK, rsp)
}
