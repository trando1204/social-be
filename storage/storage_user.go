package storage

const UserFieldUName = "user_name"
const UserFieldId = "id"
const RecipientId = "recipient_id"

type AuthType int

const (
	AuthLocalUsernamePassword AuthType = iota
	AuthMicroservicePasskey
)

type AuthClaims struct {
	Id          int64  `json:"id"`
	Username    string `json:"username"`
	LoginType   int    `json:"loginType"`
	Expire      int64  `json:"expire"`
	Role        int    `json:"role"`
	Createdt    int64  `json:"createdt"`
	LastLogindt int64  `json:"lastLogindt"`
}
