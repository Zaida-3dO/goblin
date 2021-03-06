package controllers

import (
	"net/http"

	"github.com/Zaida-3dO/goblin/internal/ports"
	"github.com/Zaida-3dO/goblin/internal/services"
	"github.com/Zaida-3dO/goblin/pkg/errs"
	"github.com/gin-gonic/gin"
)

type AuthController interface {
	Register(c *gin.Context)
	Login(c *gin.Context)
	ForgotPassword(c *gin.Context)
	ResetPassword(c *gin.Context)
}

type authController struct {
	authService services.AuthService
}

func NewAuthController(mode string) AuthController {
	var ac AuthController = &authController{
		authService: services.NewAuthService(mode),
	}
	return ac
}

func (ac *authController) Register(c *gin.Context) {
	var request ports.RegisterUserRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		httpErr := errs.NewBadRequestErr("invalid json body", err)
		c.JSON(httpErr.StatusCode, httpErr)
		return
	}

	user, at, rt, saveErr := ac.authService.RegisterUser(&request)
	if saveErr != nil {
		c.JSON(saveErr.StatusCode, saveErr)
		return
	}

	response := ports.LoginReply(user, at, rt, "account created successfully")

	c.JSON(http.StatusCreated, response)
}

func (ac *authController) Login(c *gin.Context) {
	var request ports.LoginRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		httpErr := errs.NewBadRequestErr("invalid json body", nil)
		c.JSON(httpErr.StatusCode, httpErr)
		return
	}

	user, at, rt, saveErr := ac.authService.LoginUser(&request)
	if saveErr != nil {
		c.JSON(saveErr.StatusCode, saveErr)
		return
	}

	response := ports.LoginReply(user, at, rt, "logged in successfully")

	c.JSON(http.StatusOK, response)
}

func (ac *authController) ForgotPassword(c *gin.Context) {
	var request ports.ForgotPasswordRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		httpErr := errs.NewBadRequestErr("invalid json body", nil)
		c.JSON(httpErr.StatusCode, httpErr)
		return
	}

	err := ac.authService.ForgotPassword(&request)
	if err != nil {
		c.JSON(err.StatusCode, err)
		return
	}

	response := ports.ForgotPasswordReply()
	c.JSON(http.StatusOK, response)
}

func (ac *authController) ResetPassword(c *gin.Context) {
	var request ports.ResetPasswordRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		httpErr := errs.NewBadRequestErr("invalid json body", nil)
		c.JSON(httpErr.StatusCode, httpErr)
		return
	}

	err := ac.authService.ResetPassword(&request)
	if err != nil {
		c.JSON(err.StatusCode, err)
		return
	}

	response := ports.ResetPasswordReply()
	c.JSON(http.StatusOK, response)
}
