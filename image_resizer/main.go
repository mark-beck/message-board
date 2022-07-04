package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"net/http"
	"os"

	"github.com/chai2010/webp"
	"github.com/google/uuid"
	"github.com/labstack/echo-contrib/jaegertracing"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/nfnt/resize"
	"github.com/opentracing/opentracing-go"
)

type resizeRequest struct {
	Data   string `json:"data" form:"data" query:"data" validate:"required"`
	X      uint   `json:"x" form:"x" query:"x" validate:"required"`
	Y      uint   `json:"y" form:"y" query:"y" validate:"required"`
	Format string `json:"format" form:"format" query:"format" validate:"required"`
}

type ResizeResponse struct {
	Id string `json:"id"`
	X  uint   `json:"x"`
	Y  uint   `json:"y"`
}

func main() {
	// get port from environment variable
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	e := echo.New()

	// Hide banner
	e.HideBanner = true

	c := jaegertracing.New(e, nil)
	defer c.Close()

	// Setup logging
	e.Logger.SetLevel(log.INFO)
	e.Logger.SetOutput(os.Stdout)
	e.Use(middleware.Logger())
	e.HTTPErrorHandler = customHTTPErrorHandler

	e.GET("/scale", scaleImage)
	e.POST("/scale", scaleImage)
	e.GET("/limit", limitImage)
	e.POST("/limit", limitImage)
	e.Static("/image", "res")
	e.Logger.Fatal(e.Start(fmt.Sprint(":", port)))
}

func scaleImage(c echo.Context) error {
	sp := jaegertracing.CreateChildSpan(c, "scaleImage")
	defer sp.Finish()
	return resizeImage(c, "scale")
}

func limitImage(c echo.Context) error {
	sp := jaegertracing.CreateChildSpan(c, "limitImage")
	defer sp.Finish()
	return resizeImage(c, "limit")
}

// this function takes a post request with an image and returns the image
func resizeImage(c echo.Context, operation string) error {
	sp := jaegertracing.CreateChildSpan(c, "resizeImage")
	defer sp.Finish()

	// bind request to resizeRequest struct
	req := new(resizeRequest)
	if err := c.Bind(req); err != nil {
		Error(sp, "Error binding request", err)
		return echo.ErrBadRequest
	}

	Info(sp, "Request received", req)

	// decode image from base64
	decoded, err := base64.StdEncoding.DecodeString(req.Data)
	if err != nil {
		Error(sp, "Error decoding image", err)
		return echo.ErrBadRequest
	}

	reader := bytes.NewReader(decoded)

	image, format, err := image.Decode(reader)
	if err != nil {
		Error(sp, "Error decoding image", err)
		return echo.ErrBadRequest
	}

	Info(sp, "Image format", format)

	var resized_image = image

	if operation == "scale" {
		resized_image = resize.Resize(req.X, req.Y, image, resize.Lanczos3)
	} else if operation == "limit" {
		resized_image = resize.Thumbnail(req.X, req.Y, image, resize.Lanczos3)
	} else {
		Error(sp, "Invalid operation", operation)
		return echo.ErrInternalServerError
	}

	result_array := bytes.NewBuffer(make([]byte, 0))

	switch req.Format {
	case "webp":
		err = webp.Encode(result_array, resized_image, &webp.Options{Lossless: true})
	case "png":
		err = png.Encode(result_array, resized_image)
	case "jpg", "jpeg":
		err = jpeg.Encode(result_array, resized_image, &jpeg.Options{Quality: 100})
	case "gif":
		err = gif.Encode(result_array, resized_image, &gif.Options{})
	default:
		Error(sp, "Invalid format", req.Format)
		return echo.ErrBadRequest
	}

	if err != nil {
		Error(sp, "Error encoding image", err)
		return echo.ErrInternalServerError
	}

	new_x := uint(resized_image.Bounds().Dx())
	new_y := uint(resized_image.Bounds().Dy())

	id := uuid.New().String()

	// save image to disk
	file, err := os.Create("res/" + id)
	if err != nil {
		Error(sp, "Error creating file", err)
		return echo.ErrInternalServerError
	}

	_, err = file.Write(result_array.Bytes())
	if err != nil {
		Error(sp, "Error writing file", err)
		return echo.ErrInternalServerError
	}

	return c.JSON(http.StatusOK, ResizeResponse{Id: id, X: new_x, Y: new_y})
}

func customHTTPErrorHandler(err error, c echo.Context) {
	if herror, ok := err.(*echo.HTTPError); ok {
		if herror.Code == http.StatusNotFound {
			c.Echo().DefaultHTTPErrorHandler(err, c)
			return
		}
	}
	c.Logger().Error(err)
	c.Echo().DefaultHTTPErrorHandler(err, c)
}

func Error(sp opentracing.Span, message string, err interface{}) {
	sp.SetTag("error", true)
	sp.LogKV("level", "ERROR", message, err)
}

func Info(sp opentracing.Span, message string, info interface{}) {
	sp.LogKV("level", "INFO", message, info)
}
