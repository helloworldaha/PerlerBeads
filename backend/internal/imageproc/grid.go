package imageproc

import (
	"image"
	"image/color"
	"image/png"
	"os"
)

type GridResult struct {
	Grid      [][]int `json:"grid"`
	Palette   []Color `json:"palette"`
	GridWidth int     `json:"grid_width"`
	GridHeight int    `json:"grid_height"`
}

type GenerateOptions struct {
	Size       int
	ColorLimit int
	ForceCrop  bool
}

func GenerateGrid(img image.Image, opts GenerateOptions) (*GridResult, error) {
	mode := ScaleModeKeepRatio
	if opts.ForceCrop {
		mode = ScaleModeForceCrop
	}

	scaleOpts := ScaleOptions{
		TargetSize: opts.Size,
		Mode:       mode,
	}

	resizedImg := ResizeWithOptions(img, scaleOpts)

	palette, grid := ReduceColors(resizedImg, opts.ColorLimit)

	colorPalette := make([]Color, len(palette))
	for i, c := range palette {
		colorPalette[i] = c.ToColor()
	}

	bounds := resizedImg.Bounds()

	return &GridResult{
		Grid:       grid,
		Palette:    colorPalette,
		GridWidth:  bounds.Dx(),
		GridHeight: bounds.Dy(),
	}, nil
}

func GridToImage(grid [][]int, palette []Color, cellSize int) image.Image {
	height := len(grid)
	if height == 0 {
		return nil
	}

	width := len(grid[0])

	imgWidth := width * cellSize
	imgHeight := height * cellSize

	img := image.NewRGBA(image.Rect(0, 0, imgWidth, imgHeight))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			colorIndex := grid[y][x]
			if colorIndex >= 0 && colorIndex < len(palette) {
				c := palette[colorIndex]
				pixelColor := color.RGBA{
					R: c.R,
					G: c.G,
					B: c.B,
					A: 255,
				}

				for py := 0; py < cellSize; py++ {
					for px := 0; px < cellSize; px++ {
						img.Set(x*cellSize+px, y*cellSize+py, pixelColor)
					}
				}
			}
		}
	}

	return img
}

func GridToImageWithGrid(grid [][]int, palette []Color, cellSize int) image.Image {
	height := len(grid)
	if height == 0 {
		return nil
	}

	width := len(grid[0])

	gridLineWidth := 1

	imgWidth := width*cellSize + (width+1)*gridLineWidth
	imgHeight := height*cellSize + (height+1)*gridLineWidth

	img := image.NewRGBA(image.Rect(0, 0, imgWidth, imgHeight))

	gridColor := color.RGBA{R: 200, G: 200, B: 200, A: 255}

	for y := 0; y < imgHeight; y++ {
		for x := 0; x < imgWidth; x++ {
			img.Set(x, y, gridColor)
		}
	}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			colorIndex := grid[y][x]
			if colorIndex >= 0 && colorIndex < len(palette) {
				c := palette[colorIndex]
				pixelColor := color.RGBA{
					R: c.R,
					G: c.G,
					B: c.B,
					A: 255,
				}

				startX := x*(cellSize+gridLineWidth) + gridLineWidth
				startY := y*(cellSize+gridLineWidth) + gridLineWidth

				for py := 0; py < cellSize; py++ {
					for px := 0; px < cellSize; px++ {
						img.Set(startX+px, startY+py, pixelColor)
					}
				}
			}
		}
	}

	return img
}

func SaveImageToFile(img image.Image, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return png.Encode(file, img)
}
