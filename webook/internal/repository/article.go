package repository

import (
	"context"
	"gitee.com/geekbang/basic-go/webook/internal/domain"
	"gitee.com/geekbang/basic-go/webook/internal/repository/cache"
	"gitee.com/geekbang/basic-go/webook/internal/repository/dao/article"
	"gitee.com/geekbang/basic-go/webook/pkg/logger"
	"time"
)

var (
	ErrArticleDuplicate        = article.ErrArticleDuplicate
	ErrPossibleIncorrectAuthor = article.ErrPossibleIncorrectAuthor
)

type ArticleRepository interface {
	Create(ctx context.Context, art domain.Article) (int64, error)
	Update(ctx context.Context, art domain.Article) error
	Sync(ctx context.Context, art domain.Article) (int64, error)
	SyncStatus(ctx context.Context, uid int64, id int64, status domain.ArticleStatus) (int64, error)
	List(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error)
	GetById(ctx context.Context, id int64) (domain.Article, error)
	GetPubById(ctx context.Context, id int64) (domain.Article, error)
}

type articleRepository struct {
	dao      article.ArticleDao
	cache    cache.ArticleCache
	l        logger.Logger
	userRepo UserRepo
}

func (repo *articleRepository) GetPubById(ctx context.Context, id int64) (domain.Article, error) {
	cachedArt, err := repo.cache.GetPub(ctx, id)
	if err == nil {
		return cachedArt, nil
	}
	art, err := repo.dao.GetPubById(ctx, id)
	if err != nil {
		return domain.Article{}, err
	}
	user, err := repo.userRepo.FindById(ctx, art.AuthorID)
	if err != nil {
		return domain.Article{}, err
	}

	res := domain.Article{
		Id:      art.ID,
		Title:   art.Title,
		Status:  domain.ArticleStatus(art.Status),
		Content: art.Content,
		Author: domain.Author{
			Id:   user.Id,
			Name: user.Nickname,
		},
	}
	// 也可以同步
	go func() {
		if err = repo.cache.SetPub(ctx, res); err != nil {
			repo.l.Error("缓存已发表文章失败",
				logger.Error(err), logger.Int64("aid", res.Id))
		}
	}()
	return res, nil
}

func (repo *articleRepository) GetById(ctx context.Context, id int64) (domain.Article, error) {
	cachedArt, err := repo.cache.Get(ctx, id)
	if err == nil {
		return cachedArt, nil
	}
	art, err := repo.dao.GetById(ctx, id)
	if err != nil {
		return domain.Article{}, err
	}
	return repo.toDomain(art), nil
}

func (repo *articleRepository) List(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error) {
	if offset == 0 && limit == 100 {
		data, err := repo.cache.GetFirstPage(ctx, uid)
		if err == nil {
			return data, nil
		}

	}
	data, err := repo.dao.GetByAuthor(ctx, uid, offset, limit)
	if err != nil {
		return nil, err
	}
	res := make([]domain.Article, 0, len(data))
	for _, art := range data {
		res = append(res, repo.toDomain(art))
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		repo.preCache(ctx, res)
	}()

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		err2 := repo.cache.SetFirstPage(ctx, uid, res)
		if err2 != nil {
			repo.l.Error("设置缓存失败", logger.Error(err2))
		}
	}()
	return res, nil
}

func (repo *articleRepository) SyncStatus(ctx context.Context, uid int64, id int64, status domain.ArticleStatus) (int64, error) {
	return repo.dao.SyncStatus(ctx, uid, id, status.ToUint8())
}

func (repo *articleRepository) Sync(ctx context.Context, art domain.Article) (int64, error) {
	id, err := repo.dao.Sync(ctx, repo.toEntity(art))
	if err != nil {
		return 0, err
	}
	go func() {
		author := art.Author.Id
		err = repo.cache.DelFirstPage(ctx, author)
		if err != nil {
			repo.l.Error("删除第一页缓存失败",
				logger.Int64("author", author), logger.Error(err))
		}
		user, err := repo.userRepo.FindById(ctx, author)
		if err != nil {
			repo.l.Error("提前设置缓存准备用户信息失败",
				logger.Int64("uid", author), logger.Error(err))
		}
		art.Author = domain.Author{
			Id:   user.Id,
			Name: user.Nickname,
		}
		err = repo.cache.SetPub(ctx, art)
		if err != nil {
			repo.l.Error("提前设置缓存失败",
				logger.Int64("author", author), logger.Error(err))
		}
	}()
	return id, nil
}

func (repo *articleRepository) Update(ctx context.Context, art domain.Article) error {
	err := repo.dao.Update(ctx, repo.toEntity(art))
	if err != nil {
		return err
	}
	author := art.Author.Id
	err = repo.cache.DelFirstPage(ctx, author)
	if err != nil {
		repo.l.Error("删除缓存失败",
			logger.Int64("author", author), logger.Error(err))
	}
	return nil
}

func (repo *articleRepository) Create(ctx context.Context, art domain.Article) (int64, error) {
	id, err := repo.dao.Insert(ctx, repo.toEntity(art))
	if err != nil {
		return 0, err
	}
	author := art.Author.Id
	err = repo.cache.DelFirstPage(ctx, author)
	if err != nil {
		repo.l.Error("删除缓存失败",
			logger.Int64("author", author), logger.Error(err))
	}
	return id, nil
}

func (repo *articleRepository) toEntity(art domain.Article) article.Article {
	return article.Article{
		ID:       art.Id,
		Title:    art.Title,
		Content:  art.Content,
		AuthorID: art.Author.Id,
		Status:   art.Status.ToUint8(),
	}
}

func (repo *articleRepository) toDomain(art article.Article) domain.Article {
	return domain.Article{
		Id:      art.ID,
		Title:   art.Title,
		Content: art.Content,
		Status:  domain.ArticleStatus(art.Status),
		Author: domain.Author{
			Id: art.AuthorID,
		},
		Ctime: time.UnixMilli(art.Ctime),
		Utime: time.UnixMilli(art.Utime),
	}
}

func (repo *articleRepository) preCache(ctx context.Context, arts []domain.Article) {
	// 1MB
	const contentSizeThreshold = 1024 * 1024
	if len(arts) > 0 && len(arts[0].Content) <= contentSizeThreshold {
		// 你也可以记录日志
		if err := repo.cache.Set(ctx, arts[0]); err != nil {
			repo.l.Error("提前准备缓存失败", logger.Error(err))
		}
	}
}

func NewArticleRepository(dao article.ArticleDao, c cache.ArticleCache, log logger.Logger,
	repo UserRepo) ArticleRepository {
	return &articleRepository{
		dao:      dao,
		cache:    c,
		l:        log,
		userRepo: repo,
	}
}
