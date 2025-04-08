package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/application/usecase"
	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/interfaces/http/handlers"
)

func SetupRoutes(router *gin.Engine, registrationUseCase *usecase.RegistrationUseCase) {
	// Public routes
	authGroup := router.Group("/auth")
	{
		registrationHandler := handlers.NewRegistrationHandler(registrationUseCase)
		authGroup.POST("/register", registrationHandler.Register)

		// Other auth routes will be added in future phases
	}
}
