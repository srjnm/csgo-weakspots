package services

import (
	"encoding/base64"
	"errors"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"

	heatmap "github.com/dustin/go-heatmap"
	schemes "github.com/dustin/go-heatmap/schemes"
	"github.com/gin-gonic/gin"
	"github.com/golang/geo/r2"
	demo "github.com/markus-wa/demoinfocs-golang/v2/pkg/demoinfocs"
	events "github.com/markus-wa/demoinfocs-golang/v2/pkg/demoinfocs/events"
	metadata "github.com/markus-wa/demoinfocs-golang/v2/pkg/demoinfocs/metadata"
	uuid "github.com/satori/go.uuid"
)

type DemoParseService interface {
	ParsePlayerSpots(cxt *gin.Context, demoFile *multipart.File, demoFileV *multipart.File, demoFileE *multipart.File, name string) error
}

type demoParseService struct {
}

func NewDemoParseService() DemoParseService {
	return &demoParseService{}
}

func (service *demoParseService) ParsePlayerSpots(cxt *gin.Context, demoFile *multipart.File, demoFileV *multipart.File, demoFileE *multipart.File, name string) error {
	// Check if player exists
	parser := demo.NewParser(*demoFile)

	flag := 0

	parser.RegisterEventHandler(
		func(e events.PlayerConnect) {
			if e.Player.Name == name {
				flag = 1
			}
		},
	)

	err := parser.ParseToEnd()
	if err != nil {
		return err
	}

	parser.Close()

	if flag == 0 {
		return errors.New("Player Name Invalid!")
	}

	// Get Victim Data
	mapName, vBoundingRect, vData, vScheme, err := generateHeatMapPointsData(demoFileV, name, 0)
	if err != nil {
		return err
	}

	// Get Enemy Data
	_, eBoundingRect, eData, eScheme, err := generateHeatMapPointsData(demoFileE, name, 1)
	if err != nil {
		return err
	}

	// Load Map Image
	bgImage, err := os.Open("assets/maps/" + mapName + ".jpg")
	if err != nil {
		return err
	}

	// Decode Map Image
	mapImg, _, err := image.Decode(bgImage)
	if err != nil {
		return err
	}

	// Load Overlay Image
	olImage, err := os.Open("assets/overlays/weakspot.png")
	if err != nil {
		return err
	}

	// Decode Overlay Image
	ovrImg, _, err := image.Decode(olImage)
	if err != nil {
		return err
	}

	// Create Output Canvas
	outImg := image.NewRGBA(mapImg.Bounds())

	// Apply Map as BG
	draw.Draw(outImg, mapImg.Bounds(), mapImg, image.ZP, draw.Over)

	// Genrate Victim Heatmap
	vHeatmapImg := heatmap.Heatmap(image.Rect(0, 0, vBoundingRect.Dx(), vBoundingRect.Dy()), vData, 30, 192, vScheme)

	// Genrate Enemy Heatmap
	eHeatmapImg := heatmap.Heatmap(image.Rect(0, 0, eBoundingRect.Dx(), eBoundingRect.Dy()), eData, 30, 192, eScheme)

	// Apply Victim Heatmap over BG
	draw.Draw(outImg, vBoundingRect, vHeatmapImg, image.ZP, draw.Over)

	// Apply Enemy Heatmap over BG
	draw.Draw(outImg, eBoundingRect, eHeatmapImg, image.ZP, draw.Over)

	// Apply Overlay
	draw.Draw(outImg, ovrImg.Bounds(), ovrImg, image.ZP, draw.Over)

	// Encode Image to JPEG
	jpegImg, err := os.Create("temp/temp-" + uuid.NewV1().String() + ".jpeg")
	if err != nil {
		return err
	}
	err = jpeg.Encode(jpegImg, outImg, &jpeg.Options{Quality: 100})
	if err != nil {
		return err
	}

	// Convert JPEG to Base64 for displaying in HTML
	imgByte, err := ioutil.ReadFile(jpegImg.Name())
	if err != nil {
		return err
	}

	b64 := base64.StdEncoding.EncodeToString(imgByte)

	err = os.Remove(jpegImg.Name())
	if err != nil {
		return err
	}

	cxt.HTML(http.StatusOK, "image.html", gin.H{
		"image": b64,
	})

	return nil
}

func generateHeatMapPointsData(demoFile *multipart.File, name string, playerType int) (string, image.Rectangle, []heatmap.DataPoint, []color.Color, error) {
	// Create Parser
	parser := demo.NewParser(*demoFile)

	// Get Header and Extract map meta data
	header, err := parser.ParseHeader()
	if err != nil {
		return "", image.Rectangle{}, nil, nil, err
	}

	mapMetaData := metadata.MapNameToMap[header.MapName]

	// Store Points
	var points []r2.Point

	parser.RegisterEventHandler(
		func(e events.Kill) {
			if parser.GameState().IsMatchStarted() {
				if e.Victim.Name == name && e.Killer.ActiveWeapon().Type.String() != "Knife" {
					var x, y float64

					// Convert In-Game coordinates to map coordinates
					if playerType == 0 {
						x, y = mapMetaData.TranslateScale(e.Victim.LastAlivePosition.X, e.Victim.LastAlivePosition.Y)
					} else {
						x, y = mapMetaData.TranslateScale(e.Killer.Position().X, e.Killer.Position().Y)
					}

					points = append(points, r2.Point{X: x, Y: y})
				}
			}
		},
	)

	err = parser.ParseToEnd()
	if err != nil {
		return "", image.Rectangle{}, nil, nil, err
	}

	parser.Close()

	if len(points) <= 0 {
		return "", image.Rectangle{}, nil, nil, errors.New("Failed to parse the deaths!")
	}

	// Get bounding rectangle of points to be mapped
	r2bound := r2.RectFromPoints(points...)
	boundingRect := image.Rectangle{
		Min: image.Point{X: int(r2bound.X.Lo), Y: int(r2bound.Y.Lo)},
		Max: image.Point{X: int(r2bound.X.Hi), Y: int(r2bound.Y.Hi)},
	}

	// Convert Points into heatmap Datapoints
	var data []heatmap.DataPoint
	for _, p := range points {
		data = append(data, heatmap.P(p.X, p.Y*-1))
	}

	var scheme []color.Color

	if playerType == 0 {
		scheme, err = schemes.FromImage("assets/schemes/victim.jpg")
		if err != nil {
			return "", image.Rectangle{}, nil, nil, err
		}
	} else {
		scheme, err = schemes.FromImage("assets/schemes/enemy.jpg")
		if err != nil {
			return "", image.Rectangle{}, nil, nil, err
		}
	}

	return header.MapName, boundingRect, data, scheme, nil
}
