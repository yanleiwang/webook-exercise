package service

import (
	"context"
	"gitee.com/geekbang/basic-go/webook/internal/domain"
	"gitee.com/geekbang/basic-go/webook/internal/repository"
	"gitee.com/geekbang/basic-go/webook/pkg/logger"
)

var (
	ErrArticleDuplicate        = repository.ErrArticleDuplicate
	ErrPossibleIncorrectAuthor = repository.ErrPossibleIncorrectAuthor
)

type ArticleService interface {
	Save(ctx context.Context, art domain.Article) (int64, error)
	Publish(ctx context.Context, art domain.Article) (int64, error)
	Withdraw(ctx context.Context, art domain.Article) (int64, error)
	List(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error)
	GetById(ctx context.Context, id int64) (domain.Article, error)
	GetPubById(ctx context.Context, id int64) (domain.Article, error)
}

type articleService struct {
	repo repository.ArticleRepository
	log  logger.Logger
}

func (s *articleService) GetPubById(ctx context.Context, id int64) (domain.Article, error) {
	return s.repo.GetPubById(ctx, id)
}

func (s *articleService) GetById(ctx context.Context, id int64) (domain.Article, error) {
	return s.repo.GetById(ctx, id)
}

func NewArticleService(repo repository.ArticleRepository, log logger.Logger) ArticleService {
	return &articleService{repo: repo, log: log}
}

func (s *articleService) List(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error) {
	return s.repo.List(ctx, uid, offset, limit)
}

func (s *articleService) Withdraw(ctx context.Context, art domain.Article) (int64, error) {

	return s.repo.SyncStatus(ctx, art.Author.Id, art.Id, domain.ArticleStatusPrivate)
}

func (s *articleService) Publish(ctx context.Context, art domain.Article) (int64, error) {
	art.Status = domain.ArticleStatusPublished
	return s.repo.Sync(ctx, art)
}

func (s *articleService) Save(ctx context.Context, art domain.Article) (int64, error) {
	art.Status = domain.ArticleStatusUnpublished
	if art.Id == 0 {
		return s.repo.Create(ctx, art)
	}
	err := s.repo.Update(ctx, art)
	return art.Id, err
}
