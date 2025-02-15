package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	ErrInvalidPayload = "Invalid request payload"
	ErrServer         = "the server encountered a problem and could not process your request"
	ErrNotFound       = "the requested resource could not be found"
)

func errorResponse(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"error": message})
}

func serverErrorResponse(c *gin.Context, err error) {
	log.Printf("Server error: %v", err)
	errorResponse(c, http.StatusInternalServerError, ErrServer)
}

func notFoundResponse(c *gin.Context) {
	errorResponse(c, http.StatusNotFound, ErrNotFound)
}
