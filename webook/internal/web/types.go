package web

import "github.com/gin-gonic/gin"

type handler interface {
	RegisterHandlers(engine *gin.Engine)
}
