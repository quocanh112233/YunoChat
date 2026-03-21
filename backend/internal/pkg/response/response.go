package response

import (
	"encoding/json"
	"net/http"
)

// Response represents the standard envelope wrapper for API responses
type Response struct {
	Success bool         `json:"success"`
	Data    interface{}  `json:"data,omitempty"`
	Meta    *Meta        `json:"meta,omitempty"`
	Error   *ErrorDetail `json:"error,omitempty"`
}

// Meta provides pagination or standard metadata if needed
type Meta struct {
	Cursor  string `json:"cursor,omitempty"`
	HasMore bool   `json:"has_more"`
}

// ErrorDetail defines the standard error payload
type ErrorDetail struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

// OK sends a successful JSON response
func OK(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	res := Response{
		Success: true,
		Data:    data,
	}

	_ = json.NewEncoder(w).Encode(res)
}

// OKWithMeta sends a successful JSON response with pagination metadata
func OKWithMeta(w http.ResponseWriter, statusCode int, data interface{}, meta *Meta) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	res := Response{
		Success: true,
		Data:    data,
		Meta:    meta,
	}

	_ = json.NewEncoder(w).Encode(res)
}

// Err sends a failure JSON response
func Err(w http.ResponseWriter, statusCode int, code string, message string, details ...map[string]string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	errDetail := &ErrorDetail{
		Code:    code,
		Message: message,
	}

	if len(details) > 0 && details[0] != nil {
		errDetail.Details = details[0]
	}

	res := Response{
		Success: false,
		Error:   errDetail,
	}

	_ = json.NewEncoder(w).Encode(res)
}
