package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func AboutUsPage(c *gin.Context) {

	c.HTML(http.StatusOK, "aboutUs.html", nil)
}
