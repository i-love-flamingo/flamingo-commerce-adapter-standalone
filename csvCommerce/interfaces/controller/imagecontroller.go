package controller

import (
	"bytes"
	"context"
	"image"
	"image/jpeg"
	"net/http"
	"os"
	"strconv"
	"strings"

	"path/filepath"

	"flamingo.me/flamingo/framework/flamingo"
	"flamingo.me/flamingo/framework/web"
	"flamingo.me/flamingo/framework/web/responder"
	"github.com/disintegration/imaging"
)

type (
	ImageController struct {
		responder.ErrorAware    `inject:""`
		Logger                  flamingo.Logger `inject:""`
		ProductCsvPath          string          `inject:"config:flamingo-commerce-adapter-standalone.csvCommerce.productCsvPath"`
		AllowedResizeParamaters string          `inject:"config:flamingo-commerce-adapter-standalone.csvCommerce.allowedImageResizeParamaters"`
	}
)

//renderChan - channel that is limited to 5 - used to block amount of parallel requests
var renderChan = make(chan struct{}, 5)

// Get Response for Images
func (vc *ImageController) Get(c context.Context, r *web.Request) web.Response {
	//block if buffered channel size is reached
	renderChan <- struct{}{}
	defer func() {
		//release one entry from channel (will release one block)
		<-renderChan
	}()

	filename := r.MustParam1("filename")
	size := r.MustParam1("size")
	if !inSlice(size, strings.Split(vc.AllowedResizeParamaters, ",")) {
		vc.Logger.Warn("Imagesize " + size + " not allowed!")
		size = "200x"
	}
	sizeParts := strings.SplitN(size, "x", 2)
	height := 0
	width := 0
	if w := sizeParts[0]; w != "" {
		width, _ = strconv.Atoi(w)
	}
	if h := sizeParts[1]; h != "" {
		height, _ = strconv.Atoi(h)
	}

	reader, err := os.Open(filepath.Join(filepath.Dir(vc.ProductCsvPath), filename))
	if err != nil {
		vc.Logger.Error(err)
		return vc.ErrorWithCode(c, err, 404)
	}

	defer reader.Close()
	im, _, err := image.Decode(reader)
	if err != nil {
		vc.Logger.Error(err)
		return vc.ErrorWithCode(c, err, 500)
	}

	im = imaging.Resize(im, width, height, imaging.Lanczos)

	buf := new(bytes.Buffer)
	err = jpeg.Encode(buf, im, &jpeg.Options{Quality: 90})
	if err != nil {
		vc.Logger.Error(err)
		return vc.ErrorWithCode(c, err, 500)
	}
	return &web.ContentResponse{
		ContentType: "image/png",
		BasicResponse: web.BasicResponse{
			Status: http.StatusOK,
		},
		Body: buf,
	}
}

func inSlice(search string, slice []string) bool {
	for _, v := range slice {
		if search == v {
			return true
		}
	}
	return false
}
