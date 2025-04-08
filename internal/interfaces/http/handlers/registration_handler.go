package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/application/dto"
	"github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/application/usecase"
	domainErrors "github.com/mohamedfawas/qubool-kallyanam/auth-service-qubool-kallyaanam/internal/domain/errors"
)

type RegistrationHandler struct {
	registrationUseCase *usecase.RegistrationUseCase
}

func NewRegistrationHandler(registrationUseCase *usecase.RegistrationUseCase) *RegistrationHandler {
	return &RegistrationHandler{
		registrationUseCase: registrationUseCase,
	}
}

func (h *RegistrationHandler) Register(c *gin.Context) {
	var request dto.RegistrationRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "Invalid request format",
			"error":   err.Error(),
		})
		return
	}

	// Input validation - can be extended with more sophisticated validation
	if request.Email == "" || request.Phone == "" || request.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "Missing required fields",
			"error":   domainErrors.ErrRequiredFieldMissing.Error(),
		})
		return
	}

	// Basic password strength check
	if len(request.Password) < 8 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "Password is too weak",
			"error":   domainErrors.ErrWeakPassword.Error(),
		})
		return
	}

	// Call use case
	response, err := h.registrationUseCase.Register(c.Request.Context(), &request)
	if err != nil {
		// Handle specific errors
		switch err {
		case domainErrors.ErrEmailAlreadyExists, domainErrors.ErrPhoneAlreadyExists:
			c.JSON(http.StatusConflict, gin.H{
				"status":  false,
				"message": "Registration failed",
				"error":   err.Error(),
			})
		case domainErrors.ErrInvalidEmail, domainErrors.ErrInvalidPhone:
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  false,
				"message": "Invalid input",
				"error":   err.Error(),
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  false,
				"message": "Registration failed",
				"error":   "Internal server error",
			})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  true,
		"message": "Registration initiated successfully",
		"data":    response,
	})
}
