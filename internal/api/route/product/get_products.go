package route

import (
	"github.com/gin-gonic/gin"
)

func (h *productRoute) GetProducts(ctx *gin.Context) {
	ctx.JSON(200, gin.H{
		"message": "GetProducts",
	})
}
