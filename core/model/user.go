package model

import (
	"errors"
	"strings"
	"time"

	"github.com/labring/aiproxy/core/common"
)

const (
	ErrUserNotFound = "user"
)

const (
	UserStatusEnabled  = 1
	UserStatusDisabled = 2
)

type User struct {
	ID        int       `json:"id" gorm:"primaryKey"`
	Username  string    `json:"username" gorm:"unique;not null"`
	Email     string    `json:"email" gorm:"unique"`
	GroupID   string    `json:"group_id" gorm:"index"`
	Group     *Group    `json:"group,omitempty" gorm:"foreignKey:GroupID"`
	Status    int       `json:"status" gorm:"default:1;index"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func GetUserByID(id int) (*User, error) {
	if id == 0 {
		return nil, errors.New("user id is empty")
	}

	user := User{}
	err := DB.Where("id = ?", id).First(&user).Error

	return &user, HandleNotFound(err, ErrUserNotFound)
}

func GetUserByUsername(username string) (*User, error) {
	if username == "" {
		return nil, errors.New("username is empty")
	}

	user := User{}
	err := DB.Where("username = ?", username).First(&user).Error

	return &user, HandleNotFound(err, ErrUserNotFound)
}

func CreateUser(user *User) error {
	return DB.Create(user).Error
}

func UpdateUserStatus(id, status int) error {
	result := DB.Model(&User{}).Where("id = ?", id).Update("status", status)
	return HandleUpdateResult(result, ErrUserNotFound)
}

func DeleteUser(id int) error {
	if id == 0 {
		return errors.New("user id is empty")
	}

	result := DB.Delete(&User{ID: id})
	return HandleUpdateResult(result, ErrUserNotFound)
}

func GetUsers(page, perPage int, order string, status int) (users []*User, total int64, err error) {
	tx := DB.Model(&User{})
	if status != 0 {
		tx = tx.Where("status = ?", status)
	}

	err = tx.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	if total <= 0 {
		return nil, 0, nil
	}

	limit, offset := toLimitOffset(page, perPage)
	err = tx.Order(getUserOrder(order)).Limit(limit).Offset(offset).Find(&users).Error

	return users, total, err
}

func getUserOrder(order string) string {
	prefix, suffix, _ := strings.Cut(order, "-")
	switch prefix {
	case "id", "username", "email", "status", "created_at":
		switch suffix {
		case "asc":
			return prefix + " asc"
		default:
			return prefix + " desc"
		}
	default:
		return "id desc"
	}
}

func SearchUsers(keyword string, page, perPage int, order string, status int) (users []*User, total int64, err error) {
	tx := DB.Model(&User{})
	if status != 0 {
		tx = tx.Where("status = ?", status)
	}

	if keyword != "" {
		var conditions []string
		var values []any

		if common.UsingPostgreSQL {
			conditions = append(conditions, "username ILIKE ?", "email ILIKE ?")
		} else {
			conditions = append(conditions, "username LIKE ?", "email LIKE ?")
		}

		values = append(values, "%"+keyword+"%", "%"+keyword+"%")

		if len(conditions) > 0 {
			tx = tx.Where("("+strings.Join(conditions, " OR ")+")", values...)
		}
	}

	err = tx.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	if total <= 0 {
		return nil, 0, nil
	}

	limit, offset := toLimitOffset(page, perPage)
	err = tx.Order(getUserOrder(order)).Limit(limit).Offset(offset).Find(&users).Error

	return users, total, err
}