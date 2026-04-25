package handler

import (
	"time"

	_ "ride-hailing/services/rider/docs"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

func NewRouter(h *Handler, logger *zap.Logger) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(loggerMiddleware(logger))

	v1 := r.Group("/api/v1")
	riders := v1.Group("/riders")
	{
		riders.POST("", h.RegisterRider)
		riders.GET("/:id", h.GetRider)
		riders.PUT("/:id", h.UpdateProfile)
		riders.POST("/:id/wallet/topup", h.TopUpWallet)
		riders.POST("/:id/addresses", h.AddSavedAddress)
		riders.DELETE("/:id/addresses/:addressId", h.RemoveSavedAddress)
	}

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return r
}

func loggerMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		logger.Info("http",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", time.Since(start)),
		)
	}
}
