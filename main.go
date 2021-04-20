package main

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/srjnm/csgo-weakspots/apis"
	"github.com/srjnm/csgo-weakspots/controllers"
	"github.com/srjnm/csgo-weakspots/services"
)

func main() {
	var port string
	err := godotenv.Load()
	if err != nil {
		port = os.Getenv("PORT")
	}
	if port == "" {
		port = "8080"
	}

	gin.SetMode(gin.ReleaseMode)

	demoService := services.NewDemoParseService()
	demoController := controllers.NewDemoController(demoService)
	demoAPI := apis.NewDemoAPI(demoController)

	server := gin.Default()

	server.LoadHTMLGlob("html/*.html")
	server.Static("/assets", "./html/assets")
	server.StaticFile("/favicon.png", "assets/favicon.png")

	server.NoRoute(demoAPI.NoRouteHandler)

	// Front-End
	server.GET("/", demoAPI.WeakSpotGetHandler)

	// Back-End
	server.POST("/spotmap", demoAPI.SpotMapPostHandler)

	// Serve
	server.Run(":" + port)
}
