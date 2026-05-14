package responses

import "github.com/labstack/echo/v4"

type SuccessResponse struct {
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type Pagination struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

type PaginatedResponse struct {
	Message    string      `json:"message,omitempty"`
	Data       interface{} `json:"data,omitempty"`
	Pagination Pagination  `json:"pagination"`
}

func Success(c echo.Context, status int, message string, data interface{}) error {
	return c.JSON(status, SuccessResponse{
		Message: message,
		Data:    data,
	})
}

func Paginated(c echo.Context, status int, message string, data interface{}, pagination Pagination) error {
	return c.JSON(status, PaginatedResponse{
		Message:    message,
		Data:       data,
		Pagination: pagination,
	})
}

func Error(c echo.Context, status int, error error) error {
	return c.JSON(status, ErrorResponse{
		Error: error.Error(),
	})
}
