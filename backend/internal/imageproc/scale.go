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

func AreaAverageResize(img image.Image, targetWidth, targetHeight int) image.Image {
	bounds := img.Bounds()
	srcWidth := bounds.Dx()
	srcHeight := bounds.Dy()

	dst := image.NewRGBA(image.Rect(0, 0, targetWidth, targetHeight))

	xRatio := float64(srcWidth) / float64(targetWidth)
	yRatio := float64(srcHeight) / float64(targetHeight)

	for dstY := 0; dstY < targetHeight; dstY++ {
		for dstX := 0; dstX < targetWidth; dstX++ {
			srcXStart := int(math.Floor(float64(dstX) * xRatio))
			srcYStart := int(math.Floor(float64(dstY) * yRatio))
			srcXEnd := int(math.Floor(float64(dstX+1) * xRatio))
			srcYEnd := int(math.Floor(float64(dstY+1) * yRatio))

			srcXStart = min(max(srcXStart, 0), srcWidth-1)
			srcYStart = min(max(srcYStart, 0), srcHeight-1)
			srcXEnd = min(max(srcXEnd, srcXStart+1), srcWidth)
			srcYEnd = min(max(srcYEnd, srcYStart+1), srcHeight)

			var sumR, sumG, sumB, sumA uint64
			count := 0

			for sy := srcYStart; sy < srcYEnd; sy++ {
				for sx := srcXStart; sx < srcXEnd; sx++ {
					c := img.At(bounds.Min.X+sx, bounds.Min.Y+sy)
					r, g, b, a := c.RGBA()
					sumR += uint64(r >> 8)
					sumG += uint64(g >> 8)
					sumB += uint64(b >> 8)
					sumA += uint64(a >> 8)
					count++
				}
			}

			if count > 0 {
				dst.Set(dstX, dstY, color.RGBA{
					R: uint8(sumR / uint64(count)),
					G: uint8(sumG / uint64(count)),
					B: uint8(sumB / uint64(count)),
					A: uint8(sumA / uint64(count)),
				})
			}
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

	return AreaAverageResize(img, targetWidth, targetHeight)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
