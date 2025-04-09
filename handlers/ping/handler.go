package ping

import (
	"pec2-backend/utils"

	"github.com/gin-gonic/gin"
)

type Handler struct{}

func New() *Handler {
	return &Handler{}
}

// HandlePing gère la logique de l'endpoint ping
// @Summary Ping test
// @Description Endpoint de test qui répond pong
// @Tags test
// @Produce json
// @Success 200 {object} utils.Response
// @Router /ping [get]
func (h *Handler) HandlePing(c *gin.Context) {
	utils.SendSuccess(c, 200, "Ping successful", gin.H{
		"message": "pong",
	})
}
