// Package repository hides storage details behind interfaces so services can
// be unit-tested with fakes and the store can be swapped freely.
package repository

import (
	"context"
	"errors"

	"github.com/Wi1low/go-web-template/internal/model"
	"gorm.io/gorm"
)

// ErrNotFound signals a missing row, translated from gorm's sentinel so
// callers never import gorm for control flow.
var ErrNotFound = errors.New("record not found")

// UserRepository is the persistence port for users.
type UserRepository interface {
	Create(ctx context.Context, u *model.User) error
	GetByID(ctx context.Context, id uint64) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	List(ctx context.Context, offset, limit int) ([]model.User, int64, error)
	Update(ctx context.Context, u *model.User) error
	Delete(ctx context.Context, id uint64) error
}

type userRepo struct {
	db *gorm.DB
}

// NewUserRepository returns the GORM-backed implementation.
func NewUserRepository(db *gorm.DB) UserRepository { return &userRepo{db: db} }

func (r *userRepo) Create(ctx context.Context, u *model.User) error {
	if err := r.db.WithContext(ctx).Create(u).Error; err != nil {
		return translate(err)
	}
	return nil
}

func (r *userRepo) GetByID(ctx context.Context, id uint64) (*model.User, error) {
	var u model.User
	if err := r.db.WithContext(ctx).First(&u, "id = ?", id).Error; err != nil {
		return nil, translate(err)
	}
	return &u, nil
}

func (r *userRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	var u model.User
	if err := r.db.WithContext(ctx).First(&u, "email = ?", email).Error; err != nil {
		return nil, translate(err)
	}
	return &u, nil
}

func (r *userRepo) List(ctx context.Context, offset, limit int) ([]model.User, int64, error) {
	var (
		users []model.User
		total int64
	)
	if err := r.db.WithContext(ctx).Model(&model.User{}).Count(&total).Error; err != nil {
		return nil, 0, translate(err)
	}
	if err := r.db.WithContext(ctx).Offset(offset).Limit(limit).Order("id asc").Find(&users).Error; err != nil {
		return nil, 0, translate(err)
	}
	return users, total, nil
}

func (r *userRepo) Update(ctx context.Context, u *model.User) error {
	if err := r.db.WithContext(ctx).Save(u).Error; err != nil {
		return translate(err)
	}
	return nil
}

func (r *userRepo) Delete(ctx context.Context, id uint64) error {
	res := r.db.WithContext(ctx).Delete(&model.User{}, "id = ?", id)
	if res.Error != nil {
		return translate(res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func translate(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	return err
}
