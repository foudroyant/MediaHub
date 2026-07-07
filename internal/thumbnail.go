package internal

import (
	"bytes"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type ThumbnailCache struct {
	mu     sync.RWMutex
	cache  map[string][]byte
	maxDim int
}

func NewThumbnailCache(maxDim int) *ThumbnailCache {
	return &ThumbnailCache{
		cache:  make(map[string][]byte),
		maxDim: maxDim,
	}
}

func (tc *ThumbnailCache) Get(mediaID string, path string) ([]byte, string, error) {
	tc.mu.RLock()
	data, ok := tc.cache[mediaID]
	tc.mu.RUnlock()
	if ok {
		return data, "image/jpeg", nil
	}

	data, mime, err := generateThumbnail(path, tc.maxDim)
	if err != nil {
		return nil, "", err
	}

	tc.mu.Lock()
	tc.cache[mediaID] = data
	tc.mu.Unlock()

	return data, mime, nil
}

func (tc *ThumbnailCache) Clear() {
	tc.mu.Lock()
	tc.cache = make(map[string][]byte)
	tc.mu.Unlock()
}

func generateThumbnail(path string, maxDim int) ([]byte, string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, "", err
	}
	defer f.Close()

	ext := strings.ToLower(filepath.Ext(path))
	var src image.Image

	switch ext {
	case ".jpg", ".jpeg":
		src, err = jpeg.Decode(f)
	case ".png":
		src, err = png.Decode(f)
	case ".gif":
		src, err = gif.Decode(f)
	case ".webp", ".bmp", ".svg":
		return nil, "", nil
	default:
		return nil, "", nil
	}
	if err != nil {
		return nil, "", err
	}

	bounds := src.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()

	if w <= maxDim && h <= maxDim {
		var buf bytes.Buffer
		if err := jpeg.Encode(&buf, src, &jpeg.Options{Quality: 80}); err != nil {
			return nil, "", err
		}
		return buf.Bytes(), "image/jpeg", nil
	}

	if w > h {
		h = h * maxDim / w
		w = maxDim
	} else {
		w = w * maxDim / h
		h = maxDim
	}
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}

	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	sw := float64(bounds.Dx()) / float64(w)
	sh := float64(bounds.Dy()) / float64(h)

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			sx := int(float64(x) * sw)
			sy := int(float64(y) * sh)
			dst.Set(x, y, src.At(sx, sy))
		}
	}

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, dst, &jpeg.Options{Quality: 75}); err != nil {
		return nil, "", err
	}
	return buf.Bytes(), "image/jpeg", nil
}


