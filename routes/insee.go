package routes

import (
	inseeHandler "pec2-backend/handlers/insee"

	"github.com/gin-gonic/gin"
)

func InseeRoutes(r *gin.Engine) {
	r.GET("/insee/:siret", inseeHandler.GetEntrepriseInfo)
}
