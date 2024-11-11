package route

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func addSwaggerRouters(r *gin.Engine) {
	r.GET("/swagger/v1/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}
