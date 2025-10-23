package controllers

import (
	"errors"
	"net/http"
	"notification/services/user"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	svc *user.Service
}

func NewUserController(svc *user.Service) *UserController { return &UserController{svc: svc} }

// @Summary Create user
// @Description Register a new user
// @Tags auth
// @Accept json
// @Produce json
// @Param data body user.SignupRequest true "Registration data"
// @Success 201 {object} models.MessageResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /signup [post]
func (uc *UserController) Signup(c *gin.Context) {
	var body struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if c.BindJSON(&body) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user data"})
		return
	}

	if err := uc.svc.Signup(c.Request.Context(), user.SignupRequest{
		Name:     body.Name,
		Email:    body.Email,
		Password: body.Password,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User created successfully"})
}

// @Summary Login
// @Description Authenticate a user and return a JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param data body user.LoginRequest true "Credentials"
// @Success 200 {object} models.TokenResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /login [post]
func (uc *UserController) Login(c *gin.Context) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if c.BindJSON(&body) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	token, err := uc.svc.Login(c.Request.Context(), user.LoginRequest{
		Email:    body.Email,
		Password: body.Password,
	})
	if err != nil {
		if errors.Is(err, user.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token})
}
