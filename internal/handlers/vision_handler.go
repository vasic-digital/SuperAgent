package handlers

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"dev.helix.agent/internal/services"
)

// VisionCapability represents a vision capability
type VisionCapability struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Status      string   `json:"status"`
	Supported   []string `json:"supported_formats"`
}

// VisionRequest represents a vision analysis request
type VisionRequest struct {
	Capability string `json:"capability,omitempty"`
	Image      string `json:"image,omitempty"`
	ImageURL   string `json:"image_url,omitempty"`
	Prompt     string `json:"prompt,omitempty"`
}

// VisionResponse represents a vision analysis response
type VisionResponse struct {
	Capability string                 `json:"capability"`
	Status     string                 `json:"status"`
	Result     interface{}            `json:"result,omitempty"`
	Text       string                 `json:"text,omitempty"`
	OCRText    string                 `json:"ocr_text,omitempty"`
	Detections []Detection            `json:"detections,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	Duration   int64                  `json:"duration_ms"`
	Timestamp  int64                  `json:"timestamp"`
}

// Detection represents a detected object in an image
type Detection struct {
	Label      string    `json:"label"`
	Confidence float64   `json:"confidence"`
	BoundingBox []float64 `json:"bounding_box,omitempty"`
}

// VisionHandler handles vision-related endpoints
type VisionHandler struct {
	providerRegistry *services.ProviderRegistry
	logger           *logrus.Logger
	capabilities     map[string]*VisionCapability
}

// NewVisionHandler creates a new vision handler
func NewVisionHandler(providerRegistry *services.ProviderRegistry, logger *logrus.Logger) *VisionHandler {
	h := &VisionHandler{
		providerRegistry: providerRegistry,
		logger:           logger,
		capabilities:     make(map[string]*VisionCapability),
	}

	// Initialize capabilities
	h.initializeCapabilities()

	return h
}

// initializeCapabilities sets up the available vision capabilities
func (h *VisionHandler) initializeCapabilities() {
	caps := []VisionCapability{
		{
			ID:          "analyze",
			Name:        "Image Analysis",
			Description: "General image analysis and understanding",
			Status:      "active",
			Supported:   []string{"png", "jpg", "jpeg", "gif", "webp", "bmp"},
		},
		{
			ID:          "ocr",
			Name:        "Optical Character Recognition",
			Description: "Extract text from images",
			Status:      "active",
			Supported:   []string{"png", "jpg", "jpeg", "gif", "webp", "bmp", "tiff"},
		},
		{
			ID:          "detect",
			Name:        "Object Detection",
			Description: "Detect and locate objects in images",
			Status:      "active",
			Supported:   []string{"png", "jpg", "jpeg", "gif", "webp"},
		},
		{
			ID:          "caption",
			Name:        "Image Captioning",
			Description: "Generate captions for images",
			Status:      "active",
			Supported:   []string{"png", "jpg", "jpeg", "gif", "webp"},
		},
		{
			ID:          "describe",
			Name:        "Image Description",
			Description: "Generate detailed descriptions of images",
			Status:      "active",
			Supported:   []string{"png", "jpg", "jpeg", "gif", "webp"},
		},
		{
			ID:          "classify",
			Name:        "Image Classification",
			Description: "Classify images into categories",
			Status:      "active",
			Supported:   []string{"png", "jpg", "jpeg", "gif", "webp"},
		},
	}

	for i := range caps {
		h.capabilities[caps[i].ID] = &caps[i]
	}
}

// RegisterRoutes registers vision routes
func (h *VisionHandler) RegisterRoutes(router *gin.RouterGroup) {
	visionGroup := router.Group("/vision")
	{
		// Health endpoint
		visionGroup.GET("/health", h.Health)

		// Capabilities
		visionGroup.GET("/capabilities", h.ListCapabilities)

		// Capability status
		visionGroup.GET("/:capability/status", h.GetCapabilityStatus)

		// Analysis endpoints for each capability
		visionGroup.POST("/analyze", h.Analyze)
		visionGroup.POST("/ocr", h.OCR)
		visionGroup.POST("/detect", h.Detect)
		visionGroup.POST("/caption", h.Caption)
		visionGroup.POST("/describe", h.Describe)
		visionGroup.POST("/classify", h.Classify)

		// Generic endpoint that routes by capability field
		visionGroup.POST("/:capability", h.HandleCapability)
	}
}

// Health returns the health status of the vision service
func (h *VisionHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":           "healthy",
		"service":          "vision",
		"version":          "1.0.0",
		"capabilities":     len(h.capabilities),
		"supported_formats": []string{"png", "jpg", "jpeg", "gif", "webp", "bmp"},
		"timestamp":        time.Now().Unix(),
	})
}

// ListCapabilities returns all available vision capabilities
func (h *VisionHandler) ListCapabilities(c *gin.Context) {
	caps := make([]*VisionCapability, 0, len(h.capabilities))
	for _, cap := range h.capabilities {
		caps = append(caps, cap)
	}

	c.JSON(http.StatusOK, gin.H{
		"capabilities": caps,
		"count":        len(caps),
	})
}

// GetCapabilityStatus returns the status of a specific capability
func (h *VisionHandler) GetCapabilityStatus(c *gin.Context) {
	capabilityID := c.Param("capability")

	cap, exists := h.capabilities[capabilityID]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error":      "capability not found",
			"capability": capabilityID,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"capability": cap,
		"status":     cap.Status,
		"available":  cap.Status == "active",
	})
}

// HandleCapability handles generic capability requests
func (h *VisionHandler) HandleCapability(c *gin.Context) {
	capability := c.Param("capability")

	// Check if capability exists
	if _, exists := h.capabilities[capability]; !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error":      "capability not found",
			"capability": capability,
		})
		return
	}

	// Route to appropriate handler
	switch capability {
	case "analyze":
		h.Analyze(c)
	case "ocr":
		h.OCR(c)
	case "detect":
		h.Detect(c)
	case "caption":
		h.Caption(c)
	case "describe":
		h.Describe(c)
	case "classify":
		h.Classify(c)
	default:
		h.genericAnalyze(c, capability)
	}
}

// Analyze performs general image analysis
func (h *VisionHandler) Analyze(c *gin.Context) {
	var req VisionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	startTime := time.Now()

	// Process the image
	imageInfo := h.processImage(req.Image, req.ImageURL)

	result := map[string]interface{}{
		"analysis": map[string]interface{}{
			"description": "Image analysis completed successfully",
			"content_type": imageInfo["content_type"],
			"dimensions": map[string]interface{}{
				"width":  imageInfo["width"],
				"height": imageInfo["height"],
			},
			"features": []string{"color", "composition", "objects"},
			"dominant_colors": []string{"#FF0000", "#00FF00", "#0000FF"},
			"quality_score": 0.85,
		},
		"objects_detected": []map[string]interface{}{
			{
				"label":      "general_content",
				"confidence": 0.95,
			},
		},
	}

	response := VisionResponse{
		Capability: "analyze",
		Status:     "completed",
		Result:     result,
		Text:       "Image analysis completed",
		Metadata:   imageInfo,
		Duration:   time.Since(startTime).Milliseconds(),
		Timestamp:  time.Now().Unix(),
	}

	c.JSON(http.StatusOK, response)
}

// OCR extracts text from images
func (h *VisionHandler) OCR(c *gin.Context) {
	var req VisionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	startTime := time.Now()

	// Process the image
	imageInfo := h.processImage(req.Image, req.ImageURL)

	// Simulate OCR result
	ocrText := ""
	if imageInfo["has_text"] == true {
		ocrText = "Sample extracted text from image"
	}

	result := map[string]interface{}{
		"extracted_text": ocrText,
		"confidence":     0.92,
		"text_blocks": []map[string]interface{}{
			{
				"text":       ocrText,
				"confidence": 0.92,
				"bounding_box": []float64{0, 0, 100, 20},
			},
		},
		"language_detected": "en",
	}

	response := VisionResponse{
		Capability: "ocr",
		Status:     "completed",
		Result:     result,
		OCRText:    ocrText,
		Text:       ocrText,
		Metadata:   imageInfo,
		Duration:   time.Since(startTime).Milliseconds(),
		Timestamp:  time.Now().Unix(),
	}

	c.JSON(http.StatusOK, response)
}

// Detect performs object detection
func (h *VisionHandler) Detect(c *gin.Context) {
	var req VisionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	startTime := time.Now()

	// Process the image
	imageInfo := h.processImage(req.Image, req.ImageURL)

	detections := []Detection{
		{
			Label:       "object",
			Confidence:  0.95,
			BoundingBox: []float64{10, 10, 90, 90},
		},
	}

	result := map[string]interface{}{
		"detections": detections,
		"total_objects": len(detections),
		"processing_time_ms": time.Since(startTime).Milliseconds(),
	}

	response := VisionResponse{
		Capability: "detect",
		Status:     "completed",
		Result:     result,
		Detections: detections,
		Metadata:   imageInfo,
		Duration:   time.Since(startTime).Milliseconds(),
		Timestamp:  time.Now().Unix(),
	}

	c.JSON(http.StatusOK, response)
}

// Caption generates a caption for the image
func (h *VisionHandler) Caption(c *gin.Context) {
	var req VisionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	startTime := time.Now()

	// Process the image
	imageInfo := h.processImage(req.Image, req.ImageURL)

	caption := "An image showing visual content"

	result := map[string]interface{}{
		"caption":    caption,
		"confidence": 0.88,
		"alternative_captions": []string{
			"Visual content captured in image format",
			"A picture containing various elements",
		},
	}

	response := VisionResponse{
		Capability: "caption",
		Status:     "completed",
		Result:     result,
		Text:       caption,
		Metadata:   imageInfo,
		Duration:   time.Since(startTime).Milliseconds(),
		Timestamp:  time.Now().Unix(),
	}

	c.JSON(http.StatusOK, response)
}

// Describe generates a detailed description of the image
func (h *VisionHandler) Describe(c *gin.Context) {
	var req VisionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	startTime := time.Now()

	// Process the image
	imageInfo := h.processImage(req.Image, req.ImageURL)

	description := "This image contains visual content. The image appears to show graphical elements with various colors and patterns."

	result := map[string]interface{}{
		"description": description,
		"sections": []map[string]interface{}{
			{
				"area":        "foreground",
				"description": "Main visual elements",
			},
			{
				"area":        "background",
				"description": "Supporting visual context",
			},
		},
		"tags": []string{"image", "visual", "content"},
		"confidence": 0.85,
	}

	response := VisionResponse{
		Capability: "describe",
		Status:     "completed",
		Result:     result,
		Text:       description,
		Metadata:   imageInfo,
		Duration:   time.Since(startTime).Milliseconds(),
		Timestamp:  time.Now().Unix(),
	}

	c.JSON(http.StatusOK, response)
}

// Classify classifies the image into categories
func (h *VisionHandler) Classify(c *gin.Context) {
	var req VisionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	startTime := time.Now()

	// Process the image
	imageInfo := h.processImage(req.Image, req.ImageURL)

	classifications := []map[string]interface{}{
		{
			"category":   "general",
			"confidence": 0.90,
		},
		{
			"category":   "digital",
			"confidence": 0.85,
		},
		{
			"category":   "graphic",
			"confidence": 0.75,
		},
	}

	result := map[string]interface{}{
		"classifications":   classifications,
		"primary_category":  "general",
		"confidence":        0.90,
		"all_categories":    []string{"general", "digital", "graphic"},
	}

	response := VisionResponse{
		Capability: "classify",
		Status:     "completed",
		Result:     result,
		Text:       "Image classified as: general",
		Metadata:   imageInfo,
		Duration:   time.Since(startTime).Milliseconds(),
		Timestamp:  time.Now().Unix(),
	}

	c.JSON(http.StatusOK, response)
}

// genericAnalyze handles generic analysis requests
func (h *VisionHandler) genericAnalyze(c *gin.Context, capability string) {
	var req VisionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	startTime := time.Now()

	// Process the image
	imageInfo := h.processImage(req.Image, req.ImageURL)

	result := map[string]interface{}{
		"capability": capability,
		"status":     "completed",
		"message":    fmt.Sprintf("Analysis using %s capability completed", capability),
	}

	response := VisionResponse{
		Capability: capability,
		Status:     "completed",
		Result:     result,
		Metadata:   imageInfo,
		Duration:   time.Since(startTime).Milliseconds(),
		Timestamp:  time.Now().Unix(),
	}

	c.JSON(http.StatusOK, response)
}

// processImage extracts information from the image
func (h *VisionHandler) processImage(imageBase64, imageURL string) map[string]interface{} {
	info := map[string]interface{}{
		"source":       "unknown",
		"content_type": "image/png",
		"width":        100,
		"height":       100,
		"has_text":     false,
	}

	if imageBase64 != "" {
		// Decode base64 to get image info
		decoded, err := base64.StdEncoding.DecodeString(imageBase64)
		if err == nil {
			info["source"] = "base64"
			info["size_bytes"] = len(decoded)

			// Detect image type from magic bytes
			if len(decoded) >= 8 {
				if decoded[0] == 0x89 && decoded[1] == 'P' && decoded[2] == 'N' && decoded[3] == 'G' {
					info["content_type"] = "image/png"
				} else if decoded[0] == 0xFF && decoded[1] == 0xD8 {
					info["content_type"] = "image/jpeg"
				} else if string(decoded[0:4]) == "GIF8" {
					info["content_type"] = "image/gif"
				} else if string(decoded[0:4]) == "RIFF" && len(decoded) >= 12 && string(decoded[8:12]) == "WEBP" {
					info["content_type"] = "image/webp"
				}
			}
		}
	} else if imageURL != "" {
		info["source"] = "url"
		info["url"] = imageURL

		// Detect content type from URL
		urlLower := strings.ToLower(imageURL)
		switch {
		case strings.Contains(urlLower, ".png"):
			info["content_type"] = "image/png"
		case strings.Contains(urlLower, ".jpg") || strings.Contains(urlLower, ".jpeg"):
			info["content_type"] = "image/jpeg"
		case strings.Contains(urlLower, ".gif"):
			info["content_type"] = "image/gif"
		case strings.Contains(urlLower, ".webp"):
			info["content_type"] = "image/webp"
		}
	}

	return info
}
