// auth-service-qubool-kallyaanam/internal/handler/interfaces.go
package handler

import "github.com/gin-gonic/gin"

// Handler defines the interface for HTTP handlers
type Handler interface {
	// RegisterRoutes registers the handler's routes with the given router
	RegisterRoutes(router gin.IRouter)
}
