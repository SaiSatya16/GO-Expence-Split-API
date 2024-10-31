// pkg/response/json.go

package response

import (
	"encoding/json"
	"net/http"
)

// Response represents a standardized API response structure
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

// Meta contains metadata about the response, useful for pagination
type Meta struct {
	Total      int `json:"total,omitempty"`
	Page       int `json:"page,omitempty"`
	PerPage    int `json:"per_page,omitempty"`
	TotalPages int `json:"total_pages,omitempty"`
}

// JSON sends a success response with the provided data
func JSON(w http.ResponseWriter, code int, data interface{}) {
	response := Response{
		Success: code >= 200 && code < 300,
		Data:    data,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(response)
}

// Error sends an error response with the provided message
func Error(w http.ResponseWriter, code int, message string) {
	response := Response{
		Success: false,
		Error:   message,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(response)
}

// JSONWithMeta sends a success response with data and metadata
func JSONWithMeta(w http.ResponseWriter, code int, data interface{}, meta Meta) {
	response := Response{
		Success: code >= 200 && code < 300,
		Data:    data,
		Meta:    &meta,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(response)
}

// Download sends a file download response
func Download(w http.ResponseWriter, filename string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename="+filename)

	response := Response{
		Success: true,
		Data:    data,
	}

	json.NewEncoder(w).Encode(response)
}

// ValidationError sends a response for validation errors
func ValidationError(w http.ResponseWriter, errors map[string]string) {
	response := Response{
		Success: false,
		Error:   "Validation failed",
		Data:    errors,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(response)
}

// Example usage:
// Success response:
// {
//     "success": true,
//     "data": { ... }
// }

// Error response:
// {
//     "success": false,
//     "error": "Error message"
// }

// Paginated response:
// {
//     "success": true,
//     "data": [ ... ],
//     "meta": {
//         "total": 100,
//         "page": 1,
//         "per_page": 10,
//         "total_pages": 10
//     }
// }

// Validation error response:
// {
//     "success": false,
//     "error": "Validation failed",
//     "data": {
//         "email": "Invalid email format",
//         "password": "Password must be at least 8 characters"
//     }
// }
