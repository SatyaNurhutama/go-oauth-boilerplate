package utils

import "github.com/gin-gonic/gin"

type Response struct {
	Error   bool        `json:"error"`
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// SendResponse sends a standardized JSON response
func SendResponse(c *gin.Context, code int, message string, data interface{}, err bool) {
	response := Response{
		Code:    code,
		Message: message,
		Data:    data,
		Error:   err,
	}

	c.JSON(code, response)
}
