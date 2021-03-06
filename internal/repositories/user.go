package repositories

import (
	"errors"

	"github.com/Zaida-3dO/goblin/adapters/driven/dbs"
	"github.com/Zaida-3dO/goblin/internal/dtos"
	"github.com/Zaida-3dO/goblin/pkg/errs"
	"gorm.io/gorm"
)

type UserRepo interface {
	CreateUser(user dtos.User) *errs.Err
	DeleteUser(user dtos.User) *errs.Err
	GetUser(user *dtos.User) *errs.Err
	FindUserByEmail(user *dtos.User, email string) *errs.Err
	SaveUser(user *dtos.User) *errs.Err
}

type userRepo struct {
	psql *gorm.DB
}

func NewUserRepo(mode string) UserRepo {
	if mode == "test" {
		return &UserRepoMock{}
	}
	return &userRepo{
		psql: dbs.GetInstance(mode),
	}
}

func (ur *userRepo) CreateUser(user dtos.User) *errs.Err {
	err := ur.psql.Create(&user).Error
	if err != nil {
		return errs.NewInternalServerErr(err.Error(), err)
	}
	return nil
}

func (ur *userRepo) GetUser(user *dtos.User) *errs.Err {
	err := ur.psql.First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errs.NewNotFoundErr("user not found!", nil)
	}
	if err != nil {
		return errs.NewInternalServerErr(err.Error(), err)
	}
	return nil
}

func (ur *userRepo) FindUserByEmail(user *dtos.User, email string) *errs.Err {
	err := ur.psql.Where("email = ?", email).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errs.NewNotFoundErr("user not found!", nil)
	}
	if err != nil {
		return errs.NewInternalServerErr(err.Error(), err)
	}
	return nil
}

func (ur *userRepo) SaveUser(user *dtos.User) *errs.Err {
	err := ur.psql.Save(user).Error
	if err != nil {
		return errs.NewInternalServerErr(err.Error(), err)
	}
	return nil
}

func (ur *userRepo) DeleteUser(user dtos.User) *errs.Err {
	if err := ur.psql.Delete(&user).Error; err != nil {
		return errs.NewInternalServerErr("error deleting user", err)
	}

	return nil
}
