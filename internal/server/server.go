package server

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/aisa-it/aiplan-mem/internal/config"
	"github.com/aisa-it/aiplan-mem/internal/db"
	"github.com/gofrs/uuid/v5"
	"github.com/labstack/echo/v4"
)

type Server struct {
	DataStore *db.DataStore
}

func RunServer(cfg *config.Config, ds *db.DataStore) {
	s := Server{DataStore: ds}

	e := echo.New()
	e.HideBanner = true

	blacklistGroup := e.Group("/blacklist")
	{
		blacklistGroup.GET("/:signatureBase64", s.isSessionsBlacklisted)
		blacklistGroup.POST("/:signatureBase64", s.sessionBlacklist)
	}

	lastSeenGroup := e.Group("/lastSeen")
	{
		lastSeenGroup.GET("/:userId", s.getUserLastSeen)
		lastSeenGroup.POST("/:userId", s.postUserLastSeen)
	}

	emailCodeGroup := e.Group("/lastSeen")
	{
		emailCodeGroup.GET("/:userId", s.getEmailCode)
		emailCodeGroup.POST("/:userId", s.saveEmailCode)
	}

	if err := e.Start(cfg.ListenAddr); err != nil {
		slog.Error("Start http server", "err", err)
	}
	ds.Close()
}

func sendError(c echo.Context, err error) error {
	slog.Error("Fail in handler", "path", c.Path(), "err", err)
	return c.JSON(http.StatusInternalServerError, map[string]string{
		"error": err.Error(),
	})
}

func (s Server) isSessionsBlacklisted(c echo.Context) error {
	signature, err := base64.StdEncoding.DecodeString(c.Param("signatureBase64"))
	if err != nil {
		sendError(c, err)
	}
	blk, err := s.DataStore.Sessions.IsTokenBlacklisted(signature)
	if err != nil {
		sendError(c, err)
	}
	c.Response().Header().Set("blacklisted", fmt.Sprint(blk))
	return c.NoContent(http.StatusOK)
}

func (s Server) sessionBlacklist(c echo.Context) error {
	signature, err := base64.StdEncoding.DecodeString(c.Param("signatureBase64"))
	if err != nil {
		sendError(c, err)
	}
	if err := s.DataStore.Sessions.BlacklistToken(signature); err != nil {
		sendError(c, err)
	}
	return c.NoContent(http.StatusOK)
}

func (s Server) getUserLastSeen(c echo.Context) error {
	userId := uuid.FromStringOrNil(c.Param("userId"))

	lastSeen, err := s.DataStore.Sessions.GetUserLastSeenTime(userId)
	if err != nil {
		return sendError(c, err)
	}
	c.Response().Header().Set("LastSeen", fmt.Sprint(lastSeen.Unix()))
	return c.NoContent(http.StatusOK)
}

func (s *Server) postUserLastSeen(c echo.Context) error {
	userId := uuid.FromStringOrNil(c.Param("userId"))

	if err := s.DataStore.Sessions.SaveUserLastSeenTime(userId); err != nil {
		return sendError(c, err)
	}
	return c.NoContent(http.StatusOK)
}

// EmailCodes handlers
type SaveEmailCodeRequest struct {
	NewEmail  string        `json:"new_email"`
	Code      string        `json:"code"`
	ExpiresIn time.Duration `json:"expires_in"`
}

func (s *Server) saveEmailCode(c echo.Context) error {
	userId := uuid.FromStringOrNil(c.Param("userId"))

	var req SaveEmailCodeRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return sendError(c, err)
	}

	if err := s.DataStore.EmailCodes.SaveCode(userId, req.NewEmail, req.Code, req.ExpiresIn); err != nil {
		return sendError(c, err)
	}

	return c.NoContent(http.StatusOK)
}

func (s *Server) getEmailCode(c echo.Context) error {
	userId := uuid.FromStringOrNil(c.Param("userId"))

	codeData, err := s.DataStore.EmailCodes.GetCode(userId)
	if err != nil {
		return sendError(c, err)
	}

	if codeData == nil {
		return c.NoContent(http.StatusNotFound)
	}

	return c.JSON(http.StatusOK, codeData)

}
