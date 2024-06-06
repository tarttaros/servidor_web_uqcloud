package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func HelpCenterPage(c *gin.Context) {
	c.HTML(http.StatusOK, "helpCenter.html", nil)
}
