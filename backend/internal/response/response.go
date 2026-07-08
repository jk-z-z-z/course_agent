package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Envelope struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Envelope{Code: 0, Message: "success", Data: data})
}

func Fail(c *gin.Context, status int, code int, message string) {
	c.JSON(status, Envelope{Code: code, Message: message, Data: nil})
}
