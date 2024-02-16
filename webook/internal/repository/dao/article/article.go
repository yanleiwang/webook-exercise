package article

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

var (
	ErrArticleDuplicate        = gorm.ErrDuplicatedKey
	ErrPossibleIncorrectAuthor = errors.New("用户在尝试操作非本人数据")
)

type ArticleDao interface {
	Insert(ctx context.Context, entity Article) (int64, error)
	Update(ctx context.Context, entity Article) error
	Sync(ctx context.Context, entity Article) (int64, error)
	SyncStatus(ctx context.Context, uid int64, id int64, status uint8) (int64, error)
	GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]Article, error)
	GetById(ctx context.Context, id int64) (Article, error)
	GetPubById(ctx context.Context, id int64) (PublishedArticle, error)
}

type articleDaoGORM struct {
	db *gorm.DB
}

func (d *articleDaoGORM) GetPubById(ctx context.Context, id int64) (PublishedArticle, error) {
	var res PublishedArticle
	err := d.db.WithContext(ctx).Where("id = ?", id).First(&res).Error
	return res, err
}

func (d *articleDaoGORM) GetById(ctx context.Context, id int64) (Article, error) {
	var res Article
	err := d.db.WithContext(ctx).Where("id = ?", id).First(&res).Error
	return res, err
}

func (d *articleDaoGORM) GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]Article, error) {
	var res []Article
	err := d.db.WithContext(ctx).Where("author_id = ?", uid).
		Order("utime DESC").
		Offset(offset).
		Limit(limit).
		Find(&res).Error
	return res, err
}

func (d *articleDaoGORM) SyncStatus(ctx context.Context, uid int64, id int64, status uint8) (int64, error) {

	err := d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now().UnixMilli()
		res := tx.Model(&Article{}).
			Where("id = ? and author_id = ?", id, uid).
			Updates(map[string]any{
				"status": status,
				"Utime":  now,
			})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return ErrPossibleIncorrectAuthor
		}
		res = tx.Model(&PublishedArticle{}).
			Where("id = ? and author_id = ?", id, uid).
			Updates(map[string]any{
				"status": status,
				"utime":  now,
			})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return ErrPossibleIncorrectAuthor
		}
		return nil
	})
	return id, err

}

func (d *articleDaoGORM) Sync(ctx context.Context, entity Article) (int64, error) {
	var (
		id  = entity.ID
		err error
	)
	err = d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txDAO := NewArticleDaoGORM(tx)

		if id == 0 {
			id, err = txDAO.Insert(ctx, entity)
		} else {
			err = txDAO.Update(ctx, entity)
		}
		if err != nil {
			return err
		}

		entity.ID = id
		publishArt := PublishedArticle(entity)
		now := time.Now().UnixMilli()
		publishArt.Utime = now
		publishArt.Ctime = now
		return tx.Clauses(clause.OnConflict{
			// ID 冲突的时候。实际上，在 MYSQL 里面你写不写都可以
			Columns: []clause.Column{{Name: "id"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"title":   publishArt.Title,
				"content": publishArt.Content,
				"status":  publishArt.Status,
				"utime":   publishArt.Utime,
			}),
		}).Create(&publishArt).Error
	})
	return id, err
}

func (d *articleDaoGORM) Update(ctx context.Context, entity Article) error {
	now := time.Now().UnixMilli()
	entity.Utime = now
	res := d.db.WithContext(ctx).Model(&entity).
		Where("id = ? and author_id = ?", entity.ID, entity.AuthorID).
		Updates(entity)
	err := res.Error
	if err != nil {
		return err
	}
	if res.RowsAffected == 0 {
		return ErrPossibleIncorrectAuthor
	}

	return nil
}

func (d *articleDaoGORM) Insert(ctx context.Context, entity Article) (int64, error) {
	now := time.Now().UnixMilli()
	entity.Utime = now
	entity.Ctime = now
	err := d.db.WithContext(ctx).Create(&entity).Error
	return entity.ID, err
}

func NewArticleDaoGORM(db *gorm.DB) ArticleDao {
	return &articleDaoGORM{db: db}
}
