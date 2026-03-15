package route

import (
	"github.com/gin-gonic/gin"
)

func (h *productRoute) GetProduct(ctx *gin.Context) {
	ctx.JSON(200, gin.H{
		"message": "GetProduct",
	})
}
