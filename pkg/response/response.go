// Package response provides standardized JSON response helpers for the API.
//
// WHY STANDARDIZE RESPONSES?
// Every API endpoint should return responses in the SAME format. This:
//   1. Makes the frontend predictable — it always knows where to find data vs errors
//   2. Makes debugging easier — error codes and messages are consistent
//   3. Is a hallmark of production APIs (look at Stripe, GitHub, Twilio APIs)
//
// OUR FORMAT:
//   Success: { "success": true, "data": {...} }
//   Error:   { "success": false, "error": { "code": "...", "message": "..." } }
//
// WHY NOT JUST gin.H{}?
//   gin.H{} is a shortcut for map[string]any — convenient but unstructured.
//   With typed structs, the compiler catches mistakes, and every response
//   follows the same shape automatically.

package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// APIResponse is the standard envelope for all API responses.
// The frontend can always check `success` first, then read `data` or `error`.
type APIResponse struct {
	Success bool      `json:"success"`
	Data    any       `json:"data,omitempty"`
	Error   *APIError `json:"error,omitempty"`
}

// APIError represents a structured error in API responses.
type APIError struct {
	Code    string `json:"code"`    // Machine-readable error code (e.g., "UNAUTHORIZED")
	Message string `json:"message"` // Human-readable error message
}

// PaginatedData wraps a list of items with pagination metadata.
// Example: GET /documents?page=2&per_page=10
type PaginatedData struct {
	Items      any   `json:"items"`
	TotalCount int64 `json:"total_count"`
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	TotalPages int   `json:"total_pages"`
}

// OK sends a 200 response with data.
// Use for: successful GET, PUT, PATCH requests.
func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    data,
	})
}

func Created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, APIResponse{
		Success: true,
		Data:    data,
	})
}

// NoContent sends a 204 response with no body.
// Use for: successful DELETE requests.
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// Paginated sends a 200 response with paginated data.
func Paginated(c *gin.Context, items any, totalCount int64, page, perPage int) {
	totalPages := int(totalCount) / perPage
	if int(totalCount)%perPage != 0 {
		totalPages++
	}

	OK(c, PaginatedData{
		Items:      items,
		TotalCount: totalCount,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	})
}

// BadRequest sends a 400 response.
// Use when: the client sent invalid data (bad JSON, missing fields, wrong file type).
func BadRequest(c *gin.Context, code, message string) {
	c.JSON(http.StatusBadRequest, APIResponse{
		Success: false,
		Error:   &APIError{Code: code, Message: message},
	})
}

// Unauthorized sends a 401 response.
//
//	401 = "I don't know who you are" (authentication failed)
//	403 = "I know who you are, but you can't do this" (authorization failed)
func Unauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, APIResponse{
		Success: false,
		Error:   &APIError{Code: "UNAUTHORIZED", Message: message},
	})
}

// Forbidden sends a 403 response
func Forbidden(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, APIResponse{
		Success: false,
		Error:   &APIError{Code: "FORBIDDEN", Message: message},
	})
}

// NotFound sends a 404 response.
func NotFound(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, APIResponse{
		Success: false,
		Error:   &APIError{Code: "NOT_FOUND", Message: message},
	})
}

// Conflict sends a 409 response
func Conflict(c *gin.Context, message string) {
	c.JSON(http.StatusConflict, APIResponse{
		Success: false,
		Error:   &APIError{Code: "CONFLICT", Message: message},
	})
}

// TooManyRequests sends a 429 response.
func TooManyRequests(c *gin.Context, message string) {
	c.JSON(http.StatusTooManyRequests, APIResponse{
		Success: false,
		Error:   &APIError{Code: "RATE_LIMIT_EXCEEDED", Message: message},
	})
}

// InternalError sends a 500 response.
func InternalError(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, APIResponse{
		Success: false,
		Error:   &APIError{Code: "INTERNAL_ERROR", Message: message},
	})
}