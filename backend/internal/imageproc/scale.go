package imageproc

import (
	"image"
	"image/color"
	"math"
)

type ScaleMode int

const (
	ScaleModeKeepRatio ScaleMode = iota
	ScaleModeForceCrop
)

type ScaleOptions struct {
	TargetSize int
	Mode       ScaleMode
}

func NearestNeighborResize(img image.Image, targetWidth, targetHeight int) image.Image {
	bounds := img.Bounds()
	srcWidth := bounds.Dx()
	srcHeight := bounds.Dy()

	dst := image.NewRGBA(image.Rect(0, 0, targetWidth, targetHeight))

	xRatio := float64(srcWidth) / float64(targetWidth)
	yRatio := float64(srcHeight) / float64(targetHeight)

	for dstY := 0; dstY < targetHeight; dstY++ {
		for dstX := 0; dstX < targetWidth; dstX++ {
			srcX := int(math.Floor(float64(dstX) * xRatio))
			srcY := int(math.Floor(float64(dstY) * yRatio))

			srcX = min(srcX, srcWidth-1)
			srcY = min(srcY, srcHeight-1)

			srcX += bounds.Min.X
			srcY += bounds.Min.Y

			c := img.At(srcX, srcY)
			r, g, b, a := c.RGBA()

			dst.Set(dstX, dstY, color.RGBA{
				R: uint8(r >> 8),
				G: uint8(g >> 8),
				B: uint8(b >> 8),
				A: uint8(a >> 8),
			})
		}
	}

	return dst
}

func CalculateTargetSize(srcWidth, srcHeight int, targetSize int, mode ScaleMode) (int, int) {
	if mode == ScaleModeForceCrop {
		return targetSize, targetSize
	}

	aspectRatio := float64(srcWidth) / float64(srcHeight)

	if aspectRatio >= 1 {
		return targetSize, int(float64(targetSize) / aspectRatio)
	}

	return int(float64(targetSize) * aspectRatio), targetSize
}

func CropToSquare(img image.Image) image.Image {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	if width == height {
		return img
	}

	var cropSize, offsetX, offsetY int

	if width > height {
		cropSize = height
		offsetX = (width - height) / 2
		offsetY = 0
	} else {
		cropSize = width
		offsetX = 0
		offsetY = (height - width) / 2
	}

	cropped := image.NewRGBA(image.Rect(0, 0, cropSize, cropSize))

	for y := 0; y < cropSize; y++ {
		for x := 0; x < cropSize; x++ {
			srcX := bounds.Min.X + x + offsetX
			srcY := bounds.Min.Y + y + offsetY
			cropped.Set(x, y, img.At(srcX, srcY))
		}
	}

	return cropped
}

func ResizeWithOptions(img image.Image, opts ScaleOptions) image.Image {
	bounds := img.Bounds()
	srcWidth := bounds.Dx()
	srcHeight := bounds.Dy()

	var targetWidth, targetHeight int

	if opts.Mode == ScaleModeForceCrop {
		img = CropToSquare(img)
		targetWidth = opts.TargetSize
		targetHeight = opts.TargetSize
	} else {
		targetWidth, targetHeight = CalculateTargetSize(srcWidth, srcHeight, opts.TargetSize, opts.Mode)
	}

	return NearestNeighborResize(img, targetWidth, targetHeight)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
