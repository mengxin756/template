package router

import (
	"github.com/gin-gonic/gin"
)

type UserHandler interface {
	Get(c *gin.Context)
	Create(c *gin.Context)
}

type UserDTO struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}
