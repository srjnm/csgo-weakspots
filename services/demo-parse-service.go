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
	"strings"

	heatmap "github.com/dustin/go-heatmap"
	schemes "github.com/dustin/go-heatmap/schemes"
	"github.com/fogleman/gg"
	"github.com/gin-gonic/gin"
	"github.com/golang/geo/r2"
	uuid "github.com/satori/go.uuid"
	demo "github.com/srjnm/demoinfocs-golang/pkg/demoinfocs"
	events "github.com/srjnm/demoinfocs-golang/pkg/demoinfocs/events"
	metadata "github.com/srjnm/demoinfocs-golang/pkg/demoinfocs/metadata"
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

	var players []string

	parser.RegisterEventHandler(
		func(e events.PlayerConnect) {
			players = append(players, e.Player.Name)
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
		return errors.New("Player Name Invalid! Connected Players: " + strings.Join(players, ", "))
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
	draw.Draw(outImg, mapImg.Bounds(), mapImg, image.Point{}, draw.Over)

	// Add Connection Lines
	// var lDx, lDy int
	// if vBoundingRect.Dx() > eBoundingRect.Dx() {
	// 	lDx = vBoundingRect.Dx()
	// } else {
	// 	lDx = eBoundingRect.Dx()
	// }
	// if vBoundingRect.Dy() > eBoundingRect.Dy() {
	// 	lDy = vBoundingRect.Dy()
	// } else {
	// 	lDy = eBoundingRect.Dy()
	// }

	// var lMaxX, lMaxY int
	// if vBoundingRect.Max.X > eBoundingRect.Max.X {
	// 	lMaxX = vBoundingRect.Max.X
	// } else {
	// 	lMaxX = eBoundingRect.Max.X
	// }
	// if vBoundingRect.Max.Y > eBoundingRect.Max.Y {
	// 	lMaxY = vBoundingRect.Max.Y
	// } else {
	// 	lMaxY = eBoundingRect.Max.Y
	// }

	// var lMinX, lMinY int
	// if vBoundingRect.Min.X > eBoundingRect.Min.X {
	// 	lMinX = vBoundingRect.Min.X
	// } else {
	// 	lMinX = eBoundingRect.Min.X
	// }
	// if vBoundingRect.Min.Y > eBoundingRect.Min.Y {
	// 	lMinY = vBoundingRect.Min.Y
	// } else {
	// 	lMinY = eBoundingRect.Min.Y
	// }

	lineCxt := gg.NewContext(1024, 1024)

	lineCxt.SetLineWidth(3)
	lineCxt.SetColor(color.White)
	lineCxt.SetLineCapSquare()
	// var scaleAbout float64
	// var scalingFactor float64
	// if mapName == "de_dust2" {
	// 	scaleAbout = 0.0
	// 	scalingFactor = 0.98
	// } else {
	// 	scaleAbout = 528.0
	// 	scalingFactor = 0.96
	// }
	// scalingFactor := 0.98
	// lineCxt.ScaleAbout(scalingFactor, scalingFactor, float64(0.0), float64(0.0))
	for index := range vData {
		if index < len(vData) && index < len(eData) {
			lineCxt.DrawLine(vData[index].X(), vData[index].Y()*-1, eData[index].X(), eData[index].Y()*-1)
		}
	}
	lineCxt.Stroke()

	lineName := "temp/temp-" + uuid.NewV1().String() + ".png"

	lineCxt.SavePNG(lineName)

	// Load Lines Image
	lnImage, err := os.Open(lineName)
	if err != nil {
		return err
	}

	// Decode Lines Image
	conLineImg, _, err := image.Decode(lnImage)
	if err != nil {
		return err
	}

	// Apply Connection Lines
	draw.Draw(outImg, lineCxt.Image().Bounds(), conLineImg, image.Point{}, draw.Over)

	// Genrate Victim Heatmap
	vHeatmapImg := heatmap.Heatmap(image.Rect(0, 0, vBoundingRect.Dx(), vBoundingRect.Dy()), vData, 30, 230, vScheme)

	// Genrate Enemy Heatmap
	eHeatmapImg := heatmap.Heatmap(image.Rect(0, 0, eBoundingRect.Dx(), eBoundingRect.Dy()), eData, 30, 230, eScheme)

	// Apply Victim Heatmap over BG
	draw.Draw(outImg, vBoundingRect, vHeatmapImg, image.Point{}, draw.Over)

	// Apply Enemy Heatmap over BG
	draw.Draw(outImg, eBoundingRect, eHeatmapImg, image.Point{}, draw.Over)

	// Apply Overlay
	draw.Draw(outImg, ovrImg.Bounds(), ovrImg, image.Point{}, draw.Over)

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

	err = os.Remove(lineName)
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
				// Making sure Victim or Killer is not nil(which might be caused if the file was corrupted)
				if playerType == 0 && e.Victim != nil {
					if e.Victim.Name == name && e.Weapon.Type.String() != "Knife" {
						var x, y float64

						// Convert In-Game coordinates to map coordinates
						x, y = mapMetaData.TranslateScale(e.Victim.LastAlivePosition.X, e.Victim.LastAlivePosition.Y)

						points = append(points, r2.Point{X: x, Y: y})
					}
				} else if playerType == 1 && e.Victim != nil && e.Killer != nil {
					if e.Victim.Name == name && e.Weapon.Type.String() != "Knife" {
						var x, y float64

						// Convert In-Game coordinates to map coordinates
						x, y = mapMetaData.TranslateScale(e.Killer.Position().X, e.Killer.Position().Y)

						points = append(points, r2.Point{X: x, Y: y})
					}
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
		// Expanded bounding rectangle to fix shrinkage
		Min: image.Point{X: int(r2bound.X.Lo - 16), Y: int(r2bound.Y.Lo - 16)},
		Max: image.Point{X: int(r2bound.X.Hi + 16), Y: int(r2bound.Y.Hi + 16)},
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
