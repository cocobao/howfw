package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetHome(g *gin.Context) {
	g.HTML(http.StatusOK, "home.html", gin.H{})
}
