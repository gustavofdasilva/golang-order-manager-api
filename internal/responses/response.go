package responses

import "github.com/labstack/echo/v4"

type SuccessResponse struct {
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func Success(c echo.Context, status int, message string, data interface{}) error {

	return c.JSON(status, SuccessResponse{
		Message: message,
		Data:    data,
	})
}

func Error(c echo.Context, status int, error error) error {

	return c.JSON(status, ErrorResponse{
		Error: error.Error(),
	})
}
