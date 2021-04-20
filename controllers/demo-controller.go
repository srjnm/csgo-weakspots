package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/srjnm/csgo-weakspots/services"
)

type DemoController interface {
	PlayerSpots(cxt *gin.Context) error
}

type demoController struct {
	demoParseService services.DemoParseService
}

func NewDemoController(demoParseService services.DemoParseService) DemoController {
	return &demoController{
		demoParseService: demoParseService,
	}
}

func (controller *demoController) PlayerSpots(cxt *gin.Context) error {
	header, err := cxt.FormFile("demo")
	if err != nil {
		return err
	}

	file, err := header.Open()
	if err != nil {
		return err
	}

	vFile, err := header.Open()
	if err != nil {
		return err
	}

	eFile, err := header.Open()
	if err != nil {
		return err
	}

	name := cxt.Request.FormValue("player")

	err = controller.demoParseService.ParsePlayerSpots(cxt, &file, &vFile, &eFile, name)
	if err != nil {
		return err
	}

	file.Close()

	return nil
}
