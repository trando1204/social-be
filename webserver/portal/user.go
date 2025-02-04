package portal

import (
	"socialat/be/utils"

	"gorm.io/gorm"
)

type RegisterForm struct {
	UserName    string `validate:"required,alphanum,gte=4,lte=32"`
	DisplayName string
	Password    string `validate:"required"`
	Email       string `validate:"omitempty,email"`
}

type LoginForm struct {
	UserName  string `validate:"required,alphanum,gte=4,lte=32"`
	Password  string `validate:"required"`
	LoginType int    `validate:"required"`
}

type PasskeyRegisterInfo struct {
	DisplayName string `json:"displayName"`
	Email       string `json:"email"`
	SessionKey  string `json:"sessionKey"`
}

type UserSelection struct {
	Id          uint64 `json:"id"`
	UserName    string `json:"userName"`
	DisplayName string `json:"displayName"`
}

type UpdateUserRequest struct {
	UserName    string         `json:"userName"`
	DisplayName string         `json:"displayName"`
	Password    string         `json:"password"`
	Email       string         `validate:"omitempty,email" json:"email"`
	UserId      int            `json:"userId"`
	Role        utils.UserRole `json:"role"`
}

type UserWithList struct {
	List []uint64
}

func (a UserWithList) RequestedSort() string {
	return ""
}
func (a UserWithList) BindQuery(db *gorm.DB) *gorm.DB {
	return db.Where("id IN ?", a.List)
}
func (a UserWithList) BindFirst(db *gorm.DB) *gorm.DB {
	return db
}
func (a UserWithList) BindCount(db *gorm.DB) *gorm.DB {
	return db
}
func (a UserWithList) Sortable() map[string]bool {
	return map[string]bool{}
}
