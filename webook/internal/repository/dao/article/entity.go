package article

type Article struct {
	ID       int64  `gorm:"primaryKey,autoIncrement"`
	Title    string `gorm:"type:varchar(255),size:255"`
	Content  string `gorm:"type:BLOB"`
	AuthorID int64  `gorm:"index:idx_authorID_utime, priority:1"`
	Status   uint8
	Utime    int64 `gorm:"index:idx_authorID_utime, priority:2"`
	Ctime    int64
}

// PublishedArticle 衍生类型，偷个懒
type PublishedArticle Article
