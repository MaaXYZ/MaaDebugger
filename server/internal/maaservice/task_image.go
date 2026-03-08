package maaservice

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"net/http"
	"sync"
	"time"
)

const taskImageJPEGQuality = 80

// ImageRef 是提供给前端的图片引用信息。
type ImageRef struct {
	ID          string `json:"id"`
	URL         string `json:"url"`
	MIME        string `json:"mime"`
	Width       int    `json:"width,omitempty"`
	Height      int    `json:"height,omitempty"`
	ContentSize int    `json:"content_size,omitempty"`
}

// taskImageItem 是缓存中的图片二进制数据。
type taskImageItem struct {
	ContentType string
	Data        []byte
	Width       int
	Height      int
	CreatedAt   time.Time
}

func buildTaskImageURL(id string) string {
	return "/api/task/image/" + id
}

func buildTaskImageRef(id string, item *taskImageItem) *ImageRef {
	if item == nil {
		return nil
	}
	return &ImageRef{
		ID:          id,
		URL:         buildTaskImageURL(id),
		MIME:        item.ContentType,
		Width:       item.Width,
		Height:      item.Height,
		ContentSize: len(item.Data),
	}
}

func imageBoundsSize(img image.Image) (int, int) {
	if img == nil {
		return 0, 0
	}
	bounds := img.Bounds()
	return bounds.Dx(), bounds.Dy()
}

func flattenImageToOpaque(img image.Image) image.Image {
	if img == nil {
		return nil
	}
	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, &image.Uniform{C: color.White}, image.Point{}, draw.Src)
	draw.Draw(rgba, bounds, img, bounds.Min, draw.Over)
	return rgba
}

func encodeJPEGImage(img image.Image) (*taskImageItem, error) {
	if img == nil {
		return nil, fmt.Errorf("image is nil")
	}
	flattened := flattenImageToOpaque(img)
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, flattened, &jpeg.Options{Quality: taskImageJPEGQuality}); err != nil {
		return nil, err
	}
	width, height := imageBoundsSize(flattened)
	return &taskImageItem{
		ContentType: "image/jpeg",
		Data:        buf.Bytes(),
		Width:       width,
		Height:      height,
		CreatedAt:   time.Now(),
	}, nil
}

func taskImageETag(item *taskImageItem) string {
	if item == nil {
		return ""
	}
	return fmt.Sprintf("W/\"%d-%d\"", len(item.Data), item.CreatedAt.UnixNano())
}

func WriteTaskImageResponse(w http.ResponseWriter, req *http.Request, item *taskImageItem) {
	if item == nil {
		http.NotFound(w, req)
		return
	}
	etag := taskImageETag(item)
	if etag != "" && req.Header.Get("If-None-Match") == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	w.Header().Set("Content-Type", item.ContentType)
	w.Header().Set("Cache-Control", "private, max-age=300")
	if etag != "" {
		w.Header().Set("ETag", etag)
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(item.Data)
}

func storeTaskImage(m *sync.Map, id string, img image.Image) *ImageRef {
	item, err := encodeJPEGImage(img)
	if err != nil {
		return nil
	}
	m.Store(id, item)
	return buildTaskImageRef(id, item)
}
