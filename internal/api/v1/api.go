package api

import (
	"analitycsService/internal/api/dto"
	"analitycsService/internal/repository"
	"analitycsService/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"time"
)

const (
	incorrectRequest = "Некорректный формат запроса"
	internalError    = "Непредвиденная ошибка. Повторите операцию позже"
	headKey          = "X-Tantum-Authorization"
)

type api struct {
	logger  *logrus.Logger
	service service.Service
	repo    repository.Repo
}

type Api interface {
	RegisterPublicHandlers(group *gin.RouterGroup)
	RegisterPrivateHandlers(group *gin.RouterGroup)
	RegisterInternalHandlers(group *gin.RouterGroup)
}

func New(logger *logrus.Logger, service service.Service) Api {
	return &api{
		logger:  logger,
		service: service,
	}
}

func (api *api) RegisterInternalHandlers(_ *gin.RouterGroup) {
}

func (api *api) RegisterPublicHandlers(_ *gin.RouterGroup) {
}

func (api *api) RegisterPrivateHandlers(group *gin.RouterGroup) {
	group.POST("/analytics", api.analytics)
}

func (api *api) analytics(ctx *gin.Context) {
	if len(ctx.GetHeader(headKey)) == 0 {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": incorrectRequest})
		return
	}

	bodyBytes, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": incorrectRequest})
		return
	}

	if err := api.service.SaveRowData(ctx.Request.Context(), ctx.Request.Header, bodyBytes, time.Now()); err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": internalError})
		return
	}

	ctx.JSON(http.StatusAccepted, dto.Response{Status: "OK"})
}
