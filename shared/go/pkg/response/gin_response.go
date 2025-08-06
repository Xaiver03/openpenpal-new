// Package response provides shared Gin response utilities
// Safe to import - doesn't affect existing implementations
package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GinResponse provides shared response utilities for Gin framework
type GinResponse struct{}

// NewGinResponse creates a new GinResponse instance
func NewGinResponse() *GinResponse {
	return &GinResponse{}
}

// Success sends a success JSON response
func (r *GinResponse) Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
	})
}

// SuccessWithMessage sends a success JSON response with custom message
func (r *GinResponse) SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": message,
		"data":    data,
	})
}

// Created sends a created JSON response
func (r *GinResponse) Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Resource created successfully",
		"data":    data,
	})
}

// CreatedWithMessage sends a created JSON response with custom message
func (r *GinResponse) CreatedWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": message,
		"data":    data,
	})
}

// Error sends an error JSON response
func (r *GinResponse) Error(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{
		"success": false,
		"error":   message,
	})
}

// BadRequest sends a bad request error response
func (r *GinResponse) BadRequest(c *gin.Context, message string) {
	r.Error(c, http.StatusBadRequest, message)
}

// Unauthorized sends an unauthorized error response
func (r *GinResponse) Unauthorized(c *gin.Context, message string) {
	r.Error(c, http.StatusUnauthorized, message)
}

// InternalServerError sends an internal server error response
func (r *GinResponse) InternalServerError(c *gin.Context, message string) {
	r.Error(c, http.StatusInternalServerError, message)
}

// NotFound sends a not found error response
func (r *GinResponse) NotFound(c *gin.Context, message string) {
	r.Error(c, http.StatusNotFound, message)
}

// ValidationError sends a validation error response
func (r *GinResponse) ValidationError(c *gin.Context, message string) {
	r.Error(c, http.StatusUnprocessableEntity, message)
}

// OK sends a simple OK response
func (r *GinResponse) OK(c *gin.Context, message string) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": message,
	})
}