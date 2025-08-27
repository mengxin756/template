package router

import (
    "net/http"
    "strconv"

    "github.com/gin-gonic/gin"

    appuser "example.com/classic/internal/app/user"
)

type userHandler struct {
    svc appuser.Service
}

func NewUserHandler(svc appuser.Service) UserHandler { return &userHandler{svc: svc} }

func (h *userHandler) Get(c *gin.Context) {
    id, _ := strconv.Atoi(c.Param("id"))
    u, err := h.svc.Get(c.Request.Context(), id)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": err.Error()})
        return
    }
    c.JSON(http.StatusOK, u)
}

func (h *userHandler) Create(c *gin.Context) {
    var in appuser.CreateInput
    if err := c.ShouldBindJSON(&in); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
        return
    }
    u, err := h.svc.Create(c.Request.Context(), in)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
        return
    }
    c.JSON(http.StatusOK, u)
}


