package storage

import "time"

const UserFieldUName = "user_name"
const UserFieldId = "id"
const RecipientId = "recipient_id"

type AuthType int

const (
	AuthLocalUsernamePassword AuthType = iota
	AuthMicroservicePasskey
)

type PdsUserStorage interface {
	CreatePdsUser(user *PdsUser) error
	UpdatePdsUser(user *PdsUser) error
}

type PdsUser struct {
	Id         uint64    `json:"id" gorm:"primarykey"`
	Handle     string    `json:"handle" gorm:"index:pds_user_handle_idx,unique"`
	Password   string    `json:"password"`
	Email      string    `json:"email"`
	Did        string    `json:"did"`
	InviteCode string    `json:"inviteCode"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}
type AuthClaims struct {
	Id          int64  `json:"id"`
	Username    string `json:"username"`
	LoginType   int    `json:"loginType"`
	Expire      int64  `json:"expire"`
	Role        int    `json:"role"`
	Createdt    int64  `json:"createdt"`
	LastLogindt int64  `json:"lastLogindt"`
}

func (p *psql) CreatePdsUser(user *PdsUser) error {
	return p.db.Create(user).Error
}

func (p *psql) UpdatePdsUser(user *PdsUser) error {
	return p.db.Save(user).Error
}
