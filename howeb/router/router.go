package router

import (
	"net/http"

	"github.com/cocobao/howfw/howeb/controller"
	"github.com/gin-gonic/gin"
)

func LoadRouter() http.Handler {
	gin.SetMode(gin.DebugMode)

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())
	router.LoadHTMLGlob("views/*")
	router.Static("/static", "./static")
	router.GET("/", controller.GetHome)

	return router
}
