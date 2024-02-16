package domain

import "time"

type User struct {
	Id       int64
	Email    string
	Nickname string
	Phone    string
	Password string

	WechatInfo WechatInfo
	Ctime      time.Time
}
