package domain

import "time"

type Article struct {
	Id      int64
	Title   string
	Content string
	Status  ArticleStatus
	Author  Author
	Ctime   time.Time
	Utime   time.Time
}

func (a Article) Abstract() string {
	res := []rune(a.Content)
	if len(res) > 100 {
		res = res[:100]
	}
	return string(res)
}

type Author struct {
	Id   int64
	Name string
}

//go:generate stringer -type=ArticleStatus
type ArticleStatus uint8

const (
	// ArticleStatusUnknown 未知状态
	ArticleStatusUnknown ArticleStatus = iota
	// ArticleStatusUnpublished 未发表
	ArticleStatusUnpublished
	// ArticleStatusPublished 已发表
	ArticleStatusPublished
	// ArticleStatusPrivate 仅自己可见
	ArticleStatusPrivate
)

//go:inline
func (s ArticleStatus) ToUint8() uint8 { return uint8(s) }
