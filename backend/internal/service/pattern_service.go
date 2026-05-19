package service

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"sync"

	"perlerbeads/internal/imageproc"

	"github.com/google/uuid"
)

type PatternData struct {
	ID      string            `json:"id"`
	Grid    [][]int           `json:"grid"`
	Palette []imageproc.Color `json:"palette"`
}

type PatternService struct {
	uploadDir string
	outputDir string
	patterns  sync.Map
}

func NewPatternService(uploadDir, outputDir string) *PatternService {
	os.MkdirAll(uploadDir, 0755)
	os.MkdirAll(outputDir, 0755)

	return &PatternService{
		uploadDir: uploadDir,
		outputDir: outputDir,
		patterns:  sync.Map{},
	}
}

func (s *PatternService) ProcessImage(reader io.Reader, filename string, size, colorLimit int, forceCrop bool) (*PatternData, error) {
	id := uuid.New().String()

	ext := filepath.Ext(filename)
	uploadPath := filepath.Join(s.uploadDir, id+ext)

	file, err := os.Create(uploadPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create upload file: %w", err)
	}
	defer file.Close()

	_, err = io.Copy(file, reader)
	if err != nil {
		return nil, fmt.Errorf("failed to save upload file: %w", err)
	}

	file.Seek(0, 0)

	var img image.Image
	switch ext {
	case ".jpg", ".jpeg":
		img, err = jpeg.Decode(file)
	case ".png":
		img, err = png.Decode(file)
	default:
		return nil, fmt.Errorf("unsupported image format: %s", ext)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	opts := imageproc.GenerateOptions{
		Size:       size,
		ColorLimit: colorLimit,
		ForceCrop:  forceCrop,
	}

	result, err := imageproc.GenerateGrid(img, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to generate grid: %w", err)
	}

	pattern := &PatternData{
		ID:      id,
		Grid:    result.Grid,
		Palette: result.Palette,
	}

	s.patterns.Store(id, pattern)

	cellSize := 20
	exportImg := imageproc.GridToImageWithGrid(pattern.Grid, pattern.Palette, cellSize)
	outputPath := filepath.Join(s.outputDir, id+".png")
	err = imageproc.SaveImageToFile(exportImg, outputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to save export image: %w", err)
	}

	return pattern, nil
}

func (s *PatternService) GetPattern(id string) (*PatternData, bool) {
	if pattern, ok := s.patterns.Load(id); ok {
		return pattern.(*PatternData), true
	}
	return nil, false
}

func (s *PatternService) ExportImage(id string) (string, error) {
	outputPath := filepath.Join(s.outputDir, id+".png")

	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		return "", fmt.Errorf("export image not found: %s", id)
	}

	return outputPath, nil
}

func ValidateImageFormat(filename string) bool {
	ext := filepath.Ext(filename)
	switch ext {
	case ".jpg", ".jpeg", ".png":
		return true
	default:
		return false
	}
}

const MaxFileSize = 2 * 1024 * 1024

func ValidateFileSize(size int64) bool {
	return size <= MaxFileSize
}
