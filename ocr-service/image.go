package main

import (
	"bufio"
	"context"
	"encoding/binary"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
)

// Images smaller than this on the longer side are enlarged: Tesseract reads small screenshot text badly.
const minLongSide = 1400

// recognizeImage passes an ordinary PNG/JPEG straight to Tesseract. Large, rotated (EXIF) or
// other formats are first converted to a grayscale PNG of a sensible size.
func (s *server) recognizeImage(ctx context.Context, path, work string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	cfg, format, err := image.DecodeConfig(bufio.NewReader(f))
	_ = f.Close()
	if err != nil {
		return "", &httpError{http.StatusUnprocessableEntity, "unsupported image: " + err.Error()}
	}
	pixels := cfg.Width * cfg.Height
	if pixels <= 0 || pixels > 80_000_000 {
		return "", &httpError{http.StatusUnprocessableEntity, fmt.Sprintf("image is too large: %dx%d", cfg.Width, cfg.Height)}
	}
	orientation := 1
	if format == "jpeg" {
		orientation = jpegOrientation(path)
	}
	longSide := max(cfg.Width, cfg.Height)
	direct := (format == "png" || format == "jpeg") && orientation == 1 &&
		pixels <= s.cfg.MaxPixels && longSide >= minLongSide
	pageCtx, cancel := context.WithTimeout(ctx, s.cfg.PageTimeout)
	defer cancel()
	if direct {
		return s.tesseract(pageCtx, path)
	}
	prepared := filepath.Join(work, "prepared.png")
	if err := prepareImage(path, prepared, orientation, s.cfg.MaxPixels); err != nil {
		return "", &httpError{http.StatusUnprocessableEntity, "cannot read image: " + err.Error()}
	}
	defer os.Remove(prepared)
	return s.tesseract(pageCtx, prepared)
}

func prepareImage(src, dst string, orientation, maxPixels int) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	img, _, err := image.Decode(bufio.NewReader(f))
	_ = f.Close()
	if err != nil {
		return err
	}
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if orientation >= 5 { // 90° turns swap the sides
		w, h = h, w
	}
	tw, th := targetSize(w, h, maxPixels)
	gray := resampleGray(img, orientation, tw, th)
	img = nil
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	enc := png.Encoder{CompressionLevel: png.BestSpeed}
	if err := enc.Encode(out, gray); err != nil {
		_ = out.Close()
		return err
	}
	return out.Close()
}

// targetSize keeps the picture within maxPixels and enlarges small pictures up to minLongSide (at most 3x).
func targetSize(w, h, maxPixels int) (int, int) {
	scale := 1.0
	if w*h > maxPixels {
		scale = math.Sqrt(float64(maxPixels) / float64(w*h))
	} else if long := max(w, h); long < minLongSide {
		scale = min(float64(minLongSide)/float64(long), 3)
	}
	tw, th := int(float64(w)*scale+0.5), int(float64(h)*scale+0.5)
	return max(tw, 1), max(th, 1)
}

// resampleGray writes a grayscale copy of img, turned per the EXIF orientation, at tw x th.
// Downscaling averages the source pixels of each target pixel (box filter); enlarging
// uses bilinear interpolation. Only the target buffer is allocated.
func resampleGray(img image.Image, orientation, tw, th int) *image.Gray {
	b := img.Bounds()
	sw, sh := b.Dx(), b.Dy()
	ow, oh := sw, sh // oriented source size
	if orientation >= 5 {
		ow, oh = sh, sw
	}
	lum := func(x, y int) float64 { // x, y in oriented coordinates
		sx, sy := orientSource(x, y, sw, sh, orientation)
		r, g, bl, _ := img.At(b.Min.X+sx, b.Min.Y+sy).RGBA()
		return (0.299*float64(r) + 0.587*float64(g) + 0.114*float64(bl)) / 257
	}
	dst := image.NewGray(image.Rect(0, 0, tw, th))
	fx := float64(ow) / float64(tw)
	fy := float64(oh) / float64(th)
	for y := 0; y < th; y++ {
		for x := 0; x < tw; x++ {
			var v float64
			if fx >= 1 && fy >= 1 {
				x0, x1 := int(float64(x)*fx), int(float64(x+1)*fx)
				y0, y1 := int(float64(y)*fy), int(float64(y+1)*fy)
				x1, y1 = max(x1, x0+1), max(y1, y0+1)
				x1, y1 = min(x1, ow), min(y1, oh)
				// Sample at most 4x4 source pixels per target pixel: enough for text, and fast.
				stepX, stepY := max((x1-x0)/4, 1), max((y1-y0)/4, 1)
				sum, n := 0.0, 0
				for yy := y0; yy < y1; yy += stepY {
					for xx := x0; xx < x1; xx += stepX {
						sum += lum(xx, yy)
						n++
					}
				}
				v = sum / float64(n)
			} else {
				gx := (float64(x)+0.5)*fx - 0.5
				gy := (float64(y)+0.5)*fy - 0.5
				gx, gy = clampF(gx, 0, float64(ow-1)), clampF(gy, 0, float64(oh-1))
				x0, y0 := int(gx), int(gy)
				x1, y1 := min(x0+1, ow-1), min(y0+1, oh-1)
				dx, dy := gx-float64(x0), gy-float64(y0)
				top := lum(x0, y0)*(1-dx) + lum(x1, y0)*dx
				bottom := lum(x0, y1)*(1-dx) + lum(x1, y1)*dx
				v = top*(1-dy) + bottom*dy
			}
			dst.Pix[y*dst.Stride+x] = uint8(clampF(v+0.5, 0, 255))
		}
	}
	return dst
}

func clampF(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// orientSource maps a pixel of the upright picture back to the stored pixel for EXIF orientation 1-8.
func orientSource(x, y, sw, sh, orientation int) (int, int) {
	switch orientation {
	case 2:
		return sw - 1 - x, y
	case 3:
		return sw - 1 - x, sh - 1 - y
	case 4:
		return x, sh - 1 - y
	case 5:
		return y, x
	case 6:
		return y, sh - 1 - x
	case 7:
		return sw - 1 - y, sh - 1 - x
	case 8:
		return sw - 1 - y, x
	default:
		return x, y
	}
}

// jpegOrientation reads the EXIF orientation tag (1-8) of a JPEG file; 1 when absent.
func jpegOrientation(path string) int {
	f, err := os.Open(path)
	if err != nil {
		return 1
	}
	defer f.Close()
	r := bufio.NewReader(f)
	var soi [2]byte
	if _, err := io.ReadFull(r, soi[:]); err != nil || soi != [2]byte{0xFF, 0xD8} {
		return 1
	}
	for i := 0; i < 32; i++ {
		var hdr [4]byte
		if _, err := io.ReadFull(r, hdr[:]); err != nil || hdr[0] != 0xFF {
			return 1
		}
		marker := hdr[1]
		size := int(binary.BigEndian.Uint16(hdr[2:])) - 2
		if marker == 0xDA || size < 0 { // image data starts: no EXIF before it
			return 1
		}
		if marker != 0xE1 {
			if _, err := r.Discard(size); err != nil {
				return 1
			}
			continue
		}
		data := make([]byte, size)
		if _, err := io.ReadFull(r, data); err != nil {
			return 1
		}
		if o := exifOrientation(data); o != 0 {
			return o
		}
	}
	return 1
}

func exifOrientation(app1 []byte) int {
	if len(app1) < 14 || string(app1[:6]) != "Exif\x00\x00" {
		return 0
	}
	tiff := app1[6:]
	var order binary.ByteOrder
	switch string(tiff[:2]) {
	case "II":
		order = binary.LittleEndian
	case "MM":
		order = binary.BigEndian
	default:
		return 0
	}
	ifd := int(order.Uint32(tiff[4:8]))
	if ifd+2 > len(tiff) {
		return 0
	}
	count := int(order.Uint16(tiff[ifd:]))
	for i := 0; i < count; i++ {
		entry := ifd + 2 + i*12
		if entry+12 > len(tiff) {
			return 0
		}
		if order.Uint16(tiff[entry:]) == 0x0112 {
			v := int(order.Uint16(tiff[entry+8:]))
			if v >= 1 && v <= 8 {
				return v
			}
			return 0
		}
	}
	return 0
}
