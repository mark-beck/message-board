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
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/nfnt/resize"
)

type resizeRequest struct {
	Data   string `json:"data" form:"data" query:"data" validate:"required"`
	X      uint   `json:"x" form:"x" query:"x" validate:"required"`
	Y      uint   `json:"y" form:"y" query:"y" validate:"required"`
	Format string `json:"format" form:"format" query:"format" validate:"required"`
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
	return resizeImage(c, "scale")
}

func limitImage(c echo.Context) error {
	return resizeImage(c, "limit")
}

// this function takes a post request with an image and returns the image
func resizeImage(c echo.Context, operation string) error {

	// bind request to resizeRequest struct
	req := new(resizeRequest)
	if err := c.Bind(req); err != nil {
		return err
	}

	c.Logger().Debugf("%+v", req)

	// decode image from base64
	decoded, err := base64.StdEncoding.DecodeString(req.Data)
	if err != nil {
		return err
	}

	reader := bytes.NewReader(decoded)

	image, format, err := image.Decode(reader)
	if err != nil {
		return err
	}

	c.Logger().Info("Decoded image with format: " + format)

	var resized_image = image

	if operation == "scale" {
		resized_image = resize.Resize(req.X, req.Y, image, resize.Lanczos3)
	} else if operation == "limit" {
		resized_image = resize.Thumbnail(req.X, req.Y, image, resize.Lanczos3)
	} else {
		panic("Invalid operation, this cannot happen")
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
		return fmt.Errorf("invalid format: %s", req.Format)
	}

	if err != nil {
		return fmt.Errorf("error encoding image: %v", err.Error())
	}

	b64_array := base64.StdEncoding.EncodeToString(result_array.Bytes())
	new_x := uint(resized_image.Bounds().Dx())
	new_y := uint(resized_image.Bounds().Dy())

	return c.JSON(http.StatusOK, resizeRequest{Data: b64_array, Format: req.Format, X: new_x, Y: new_y})
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
