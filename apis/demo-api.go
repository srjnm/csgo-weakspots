package apis

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/srjnm/csgo-weakspots/controllers"
)

type DemoAPI struct {
	demoController controllers.DemoController
}

func NewDemoAPI(demoController controllers.DemoController) *DemoAPI {
	return &DemoAPI{
		demoController: demoController,
	}
}

func (api *DemoAPI) SpotMapPostHandler(cxt *gin.Context) {
	err := api.demoController.PlayerSpots(cxt)
	if err != nil {
		cxt.HTML(http.StatusConflict, "error.html", gin.H{
			"message": err.Error(),
		})
		return
	}
}

func (api *DemoAPI) WeakSpotGetHandler(cxt *gin.Context) {
	cxt.HTML(http.StatusOK, "index.html", nil)
}

func (api *DemoAPI) NoRouteHandler(cxt *gin.Context) {
	cxt.HTML(http.StatusNotFound, "404.html", nil)
}
