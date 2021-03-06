package services

import (
	"fmt"
	"time"

	"github.com/Zaida-3dO/goblin/internal/dtos"
	"github.com/Zaida-3dO/goblin/internal/ports"
	"github.com/Zaida-3dO/goblin/internal/repositories"
	"github.com/Zaida-3dO/goblin/pkg/common"
	"github.com/Zaida-3dO/goblin/pkg/errs"
)

type AuthService interface {
	RegisterUser(req *ports.RegisterUserRequest) (*dtos.User, string, string, *errs.Err)
	LoginUser(req *ports.LoginRequest) (*dtos.User, string, string, *errs.Err)
	ForgotPassword(req *ports.ForgotPasswordRequest) *errs.Err
	ResetPassword(req *ports.ResetPasswordRequest) *errs.Err
}

type authService struct {
	userRepo      repositories.UserRepo
	userTokenRepo repositories.UserTokenRepo
	emailService  EmailServiceInterface
}

func NewAuthService(mode string) AuthService {
	var as AuthService = &authService{
		userRepo:      repositories.NewUserRepo(mode),
		userTokenRepo: repositories.NewUserTokenRepo(mode),
		emailService:  NewEmailService(),
	}
	return as
}

func (as *authService) RegisterUser(req *ports.RegisterUserRequest) (*dtos.User, string, string, *errs.Err) {
	err := req.ValidateRegisterUserRequest()
	if err != nil {
		return nil, "", "", err
	}

	var user = dtos.NewUser(req.FirstName, req.LastName, req.Email, req.PhoneNumber, req.Password)
	err = EnsureEmailNotTaken(as.userRepo, req.Email)
	if err != nil {
		return nil, "", "", err
	}

	colour, colourErr := common.UserDefaultProfileColour(req.FirstName, req.LastName)
	if colourErr != nil {
		return nil, "", "", errs.NewBadRequestErr(colourErr.Error(), colourErr)
	}

	user.Colour = colour
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	var hashErr error
	user.Password, hashErr = common.HashPassword(user.Password)
	if err != nil {
		return nil, "", "", errs.NewInternalServerErr(hashErr.Error(), hashErr)
	}

	err = as.userRepo.CreateUser(*user)
	if err != nil {
		return nil, "", "", err
	}

	var lr *ports.LoginRequest
	lr, err = ports.NewLoginRequest(req.Email, req.Password)
	if err != nil {
		return nil, "", "", err
	}

	return as.LoginUser(lr)
}

func (as *authService) LoginUser(req *ports.LoginRequest) (*dtos.User, string, string, *errs.Err) {
	var user dtos.User
	var ts = NewTokenService("psql")

	err := as.userRepo.FindUserByEmail(&user, req.Email)
	if err != nil {
		return nil, "", "", errs.NewBadRequestErr("invalid login credentials", nil)
	}

	compareErr := common.ComparePassword(user.Password, req.Password)
	if compareErr != nil {
		return nil, "", "", errs.NewBadRequestErr("invalid login credentials", nil)
	}

	var tokenPair *token
	tokenPair, err = ts.GenerateTokenPair(user.ID)
	if err != nil {
		return nil, "", "", err
	}

	userToken := dtos.NewUserToken(user.ID, tokenPair.accessToken, tokenPair.refreshToken, tokenPair.accessUUID, tokenPair.refreshUUID)

	err = as.userTokenRepo.CreateToken(*userToken)
	if err != nil {
		return nil, "", "", err
	}

	return &user, userToken.AccessToken, userToken.RefreshToken, nil
}

func (as authService) ForgotPassword(req *ports.ForgotPasswordRequest) *errs.Err {
	var ts = NewTokenService("psql")

	err := req.ValidateForgotPasswordRequest()
	if err != nil {
		return err
	}

	var user dtos.User
	if err = as.userRepo.FindUserByEmail(&user, req.Email); err != nil {
		return err
	}

	var passwordResetToken *string
	passwordResetToken, err = ts.GenerateEmailToken(user.Email)
	if err != nil {
		return err
	}

	if err = as.emailService.SendForgotPasswordEmail(user.FirstName, req.Email, *passwordResetToken, req.RedirectTo); err != nil {
		// log the error
		fmt.Printf("error sending email: %v\n", err)
	}

	return nil
}

func (as authService) ResetPassword(req *ports.ResetPasswordRequest) *errs.Err {
	var ts = NewTokenService("psql")

	err := req.ValidateResetPasswordRequest()
	if err != nil {
		return err
	}

	var (
		user  dtos.User
		email string
	)
	email, err = ts.GetEmailFromToken(req.Token)
	if err != nil {
		return err
	}

	if err = as.userRepo.FindUserByEmail(&user, email); err != nil {
		return err
	}

	var hashErr error
	user.Password, hashErr = common.HashPassword(req.Password)
	if err != nil {
		return errs.NewInternalServerErr(hashErr.Error(), hashErr)
	}

	if err = as.userRepo.SaveUser(&user); err != nil {
		return err
	}

	if err = as.emailService.SendPasswordResetEmail(user.FirstName, email); err != nil {
		// log the error
		fmt.Printf("error sending email: %v\n", err)
	}

	return nil
}
