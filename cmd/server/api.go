package main

import (
	"encoding/base64"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func startAPI() {
	r := gin.Default()
	r.GET("/frames", getFramesHandler)
	r.GET("/frame/:id", getFrameByID)
	r.Run(":8080")
}

func getFramesHandler(c *gin.Context) {
	fromStr := c.Query("from")
	toStr := c.Query("to")
	from, _ := strconv.ParseUint(fromStr, 10, 32)
	to, _ := strconv.ParseUint(toStr, 10, 32)
	frames := getFrames(uint32(from), uint32(to))
	// Преобразуем в JSON, данные могут быть бинарными – лучше base64 или hex
	type respFrame struct {
		FrameID   uint32 `json:"frame_id"`
		Timestamp uint64 `json:"timestamp"`
		Data      string `json:"data"` // base64
	}
	resp := make([]respFrame, len(frames))
	for i, f := range frames {
		resp[i] = respFrame{
			FrameID:   f.FrameID,
			Timestamp: f.Timestamp,
			Data:      base64.StdEncoding.EncodeToString(f.Data),
		}
	}
	c.JSON(http.StatusOK, resp)
}

func getFrameByID(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)
	framesMu.RLock()
	defer framesMu.RUnlock()
	for _, f := range frames {
		if f.FrameID == uint32(id) {
			c.JSON(http.StatusOK, gin.H{
				"frame_id":  f.FrameID,
				"timestamp": f.Timestamp,
				"data":      base64.StdEncoding.EncodeToString(f.Data),
			})
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "frame not found"})
}
