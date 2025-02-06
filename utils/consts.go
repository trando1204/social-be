package utils

const (
	SortByCreated  = 1
	SortByLastSeen = 2

	SortASC              = 1
	SortDESC             = 2
	LimitOfFetchTimeline = 50
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

type ResponseData struct {
	IsError bool        `json:"error"`
	Msg     string      `json:"msg"`
	Data    interface{} `json:"data"`
}

type UserRole int

const (
	UserRoleNone UserRole = iota
	UserRoleAdmin
)
