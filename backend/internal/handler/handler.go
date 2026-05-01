package handler

import (
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"perlerbeads/internal/service"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type GenerateRequest struct {
	Size      int  `form:"size"`
	ColorLimit int `form:"color_limit"`
	ForceCrop bool `form:"force_crop"`
}

type Handler struct {
	patternService *service.PatternService
	outputDir      string
}

func NewHandler(patternService *service.PatternService, outputDir string) *Handler {
	return &Handler{
		patternService: patternService,
		outputDir:      outputDir,
	}
}

func (h *Handler) Generate(c *gin.Context) {
	file, header, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "请上传图片文件",
		})
		return
	}
	defer file.Close()

	if !service.ValidateImageFormat(header.Filename) {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "仅支持 jpg 和 png 格式的图片",
		})
		return
	}

	if !service.ValidateFileSize(header.Size) {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "图片大小不能超过 2MB",
		})
		return
	}

	sizeStr := c.DefaultPostForm("size", "32")
	size, err := strconv.Atoi(sizeStr)
	if err != nil {
		size = 32
	}
	if size != 16 && size != 32 && size != 64 {
		size = 32
	}

	colorLimitStr := c.DefaultPostForm("color_limit", "16")
	colorLimit, err := strconv.Atoi(colorLimitStr)
	if err != nil {
		colorLimit = 16
	}
	if colorLimit < 2 || colorLimit > 64 {
		colorLimit = 16
	}

	forceCropStr := c.DefaultPostForm("force_crop", "false")
	forceCrop := forceCropStr == "true"

	pattern, err := h.patternService.ProcessImage(file, header.Filename, size, colorLimit, forceCrop)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: "处理图片失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data:    pattern,
	})
}

func (h *Handler) Export(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "缺少参数 id",
		})
		return
	}

	outputPath, err := h.patternService.ExportImage(id)
	if err != nil {
		c.JSON(http.StatusNotFound, Response{
			Code:    404,
			Message: "图纸不存在",
		})
		return
	}

	filename := filepath.Base(outputPath)
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "image/png")
	c.File(outputPath)
}

func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "ok",
	})
}
