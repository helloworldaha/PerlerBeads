package imageproc

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"sort"
)

type RGB struct {
	R, G, B uint32
}

type LAB struct {
	L, A, B float64
}

type Color struct {
	R   uint8  `json:"r"`
	G   uint8  `json:"g"`
	B   uint8  `json:"b"`
	Hex string `json:"hex"`
}

type ColorCount struct {
	Color Color
	Count int
}

type ColorWithCount struct {
	Color RGB
	Count int
}

func NewRGB(c color.Color) RGB {
	r, g, b, _ := c.RGBA()
	return RGB{R: r >> 8, G: g >> 8, B: b >> 8}
}

func (r RGB) ToColor() Color {
	return Color{
		R:   uint8(r.R),
		G:   uint8(r.G),
		B:   uint8(r.B),
		Hex: fmt.Sprintf("#%02X%02X%02X", r.R, r.G, r.B),
	}
}

func RGBToLAB(rgb RGB) LAB {
	r := float64(rgb.R) / 255.0
	g := float64(rgb.G) / 255.0
	b := float64(rgb.B) / 255.0

	if r > 0.04045 {
		r = math.Pow((r+0.055)/1.055, 2.4)
	} else {
		r = r / 12.92
	}
	if g > 0.04045 {
		g = math.Pow((g+0.055)/1.055, 2.4)
	} else {
		g = g / 12.92
	}
	if b > 0.04045 {
		b = math.Pow((b+0.055)/1.055, 2.4)
	} else {
		b = b / 12.92
	}

	x := r*0.4124 + g*0.3576 + b*0.1805
	y := r*0.2126 + g*0.7152 + b*0.0722
	z := r*0.0193 + g*0.1192 + b*0.9505

	x = x / 0.95047
	y = y / 1.00000
	z = z / 1.08883

	if x > 0.008856 {
		x = math.Pow(x, 1.0/3.0)
	} else {
		x = (7.787 * x) + (16.0 / 116.0)
	}
	if y > 0.008856 {
		y = math.Pow(y, 1.0/3.0)
	} else {
		y = (7.787 * y) + (16.0 / 116.0)
	}
	if z > 0.008856 {
		z = math.Pow(z, 1.0/3.0)
	} else {
		z = (7.787 * z) + (16.0 / 116.0)
	}

	return LAB{
		L: (116.0 * y) - 16.0,
		A: 500.0 * (x - y),
		B: 200.0 * (y - z),
	}
}

func RGBDistance(a, b RGB) float64 {
	dr := float64(a.R) - float64(b.R)
	dg := float64(a.G) - float64(b.G)
	db := float64(a.B) - float64(b.B)

	rMean := (float64(a.R) + float64(b.R)) / 2.0

	if rMean < 128 {
		return 2*dr*dr + 4*dg*dg + 3*db*db
	}
	return 3*dr*dr + 4*dg*dg + 2*db*db
}

func LABDistance(a, b LAB) float64 {
	dL := a.L - b.L
	dA := a.A - b.A
	dB := a.B - b.B
	return dL*dL + dA*dA + dB*dB
}

func FindNearestColor(target RGB, palette []RGB) int {
	minDist := float64(1e18)
	minIndex := 0

	targetLAB := RGBToLAB(target)

	for i, color := range palette {
		colorLAB := RGBToLAB(color)
		dist := LABDistance(targetLAB, colorLAB)
		if dist < minDist {
			minDist = dist
			minIndex = i
		}
	}

	return minIndex
}

func ExtractColorsWithCount(img image.Image) []ColorWithCount {
	bounds := img.Bounds()
	colorMap := make(map[RGB]int)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := NewRGB(img.At(x, y))
			colorMap[c]++
		}
	}

	colors := make([]ColorWithCount, 0, len(colorMap))
	for c, count := range colorMap {
		colors = append(colors, ColorWithCount{Color: c, Count: count})
	}

	sort.Slice(colors, func(i, j int) bool {
		return colors[i].Count > colors[j].Count
	})

	return colors
}

func ExtractColors(img image.Image) []RGB {
	colorsWithCount := ExtractColorsWithCount(img)
	colors := make([]RGB, len(colorsWithCount))
	for i, c := range colorsWithCount {
		colors[i] = c.Color
	}
	return colors
}

func ReduceColors(img image.Image, targetCount int) ([]RGB, [][]int) {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	grid := make([][]int, height)
	for i := range grid {
		grid[i] = make([]int, width)
	}

	colorsWithCount := ExtractColorsWithCount(img)

	if len(colorsWithCount) <= targetCount {
		colors := make([]RGB, len(colorsWithCount))
		colorToIndex := make(map[RGB]int)
		for i, c := range colorsWithCount {
			colors[i] = c.Color
			colorToIndex[c.Color] = i
		}

		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				c := NewRGB(img.At(x, y))
				grid[y][x] = colorToIndex[c]
			}
		}
		return colors, grid
	}

	return kMeansPlusPlusQuantize(img, colorsWithCount, targetCount)
}

func kMeansPlusPlusQuantize(img image.Image, colorsWithCount []ColorWithCount, k int) ([]RGB, [][]int) {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	totalPixels := width * height
	sampleRate := 1
	if totalPixels > 5000 {
		sampleRate = totalPixels / 5000
		if sampleRate < 1 {
			sampleRate = 1
		}
	}

	samples := make([]RGB, 0, totalPixels/sampleRate+1)
	for y := 0; y < height; y += sampleRate {
		for x := 0; x < width; x += sampleRate {
			samples = append(samples, NewRGB(img.At(x, y)))
		}
	}

	centroids := initializeCentroidsKMeansPlusPlus(samples, k)

	maxIterations := 50
	for iter := 0; iter < maxIterations; iter++ {
		clusters := make([][]RGB, k)
		clusterCounts := make([]int, k)

		for _, c := range samples {
			nearest := FindNearestColor(c, centroids)
			clusters[nearest] = append(clusters[nearest], c)
			clusterCounts[nearest]++
		}

		newCentroids := make([]RGB, k)
		changed := false

		for i := 0; i < k; i++ {
			if len(clusters[i]) == 0 {
				farthestDist := float64(-1)
				farthestColor := RGB{}
				for _, c := range samples {
					minDist := float64(1e18)
					for j := 0; j < k; j++ {
						if j == i {
							continue
						}
						dist := RGBDistance(c, newCentroids[j])
						if dist < minDist {
							minDist = dist
						}
					}
					if minDist > farthestDist {
						farthestDist = minDist
						farthestColor = c
					}
				}
				newCentroids[i] = farthestColor
				changed = true
				continue
			}

			var sumR, sumG, sumB uint64
			for _, c := range clusters[i] {
				sumR += uint64(c.R)
				sumG += uint64(c.G)
				sumB += uint64(c.B)
			}

			count := uint64(len(clusters[i]))
			newCentroids[i] = RGB{
				R: uint32(sumR / count),
				G: uint32(sumG / count),
				B: uint32(sumB / count),
			}

			if newCentroids[i] != centroids[i] {
				changed = true
			}
		}

		centroids = newCentroids
		if !changed {
			break
		}
	}

	grid := make([][]int, height)
	for i := range grid {
		grid[i] = make([]int, width)
	}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			c := NewRGB(img.At(x, y))
			grid[y][x] = FindNearestColor(c, centroids)
		}
	}

	return centroids, grid
}

func initializeCentroidsKMeansPlusPlus(samples []RGB, k int) []RGB {
	if len(samples) == 0 || k == 0 {
		return nil
	}

	if len(samples) < k {
		result := make([]RGB, len(samples))
		copy(result, samples)
		for len(result) < k {
			result = append(result, samples[0])
		}
		return result
	}

	centroids := make([]RGB, 0, k)
	used := make(map[RGB]bool)

	firstIndex := len(samples) / 2
	centroids = append(centroids, samples[firstIndex])
	used[samples[firstIndex]] = true

	for i := 1; i < k; i++ {
		distances := make([]float64, len(samples))
		totalDist := 0.0

		for j, c := range samples {
			if used[c] {
				distances[j] = 0
				continue
			}

			minDist := float64(1e18)
			for _, centroid := range centroids {
				dist := RGBDistance(c, centroid)
				if dist < minDist {
					minDist = dist
				}
			}
			distances[j] = minDist
			totalDist += minDist
		}

		if totalDist == 0 {
			for _, c := range samples {
				if !used[c] {
					centroids = append(centroids, c)
					used[c] = true
					break
				}
			}
			continue
		}

		r := totalDist * (0.1 + 0.8*float64(i)/float64(k))
		currentSum := 0.0
		selectedIndex := 0

		for j := range samples {
			currentSum += distances[j]
			if currentSum >= r {
				selectedIndex = j
				break
			}
		}

		for j := range samples {
			if distances[j] > distances[selectedIndex] {
				selectedIndex = j
			}
		}

		centroids = append(centroids, samples[selectedIndex])
		used[samples[selectedIndex]] = true
	}

	return centroids
}

func CountColors(grid [][]int, paletteSize int) []int {
	counts := make([]int, paletteSize)

	for _, row := range grid {
		for _, colorIndex := range row {
			if colorIndex >= 0 && colorIndex < paletteSize {
				counts[colorIndex]++
			}
		}
	}

	return counts
}
