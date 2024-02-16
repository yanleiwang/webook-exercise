package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"gitee.com/geekbang/basic-go/webook/internal/domain"
	"gitee.com/geekbang/basic-go/webook/internal/integration/startup"
	"gitee.com/geekbang/basic-go/webook/internal/repository/dao/article"
	"gitee.com/geekbang/basic-go/webook/internal/web"
	"gitee.com/geekbang/basic-go/webook/internal/web/jwt"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type ArticleHandlerTestSuit struct {
	suite.Suite
	server *gin.Engine
	db     *gorm.DB
}

func (s *ArticleHandlerTestSuit) SetupSuite() {
	s.server = gin.Default()
	s.server.Use(func(ctx *gin.Context) {
		// 直接设置好
		ctx.Set(jwt.KeyAccessClaims, &jwt.AccessClaims{
			Uid: 123,
		})
		ctx.Next()
	})
	startup.InitViper()
	s.db = startup.InitDB()
	hdl := startup.InitArticleHandler()
	hdl.RegisterHandlers(s.server)
}

//func (s *ArticleHandlerTestSuit) SetupTest() {
//	err := s.db.Exec("TRUNCATE TABLE `articles`").Error
//	assert.NoError(s.T(), err)
//	err = s.db.Exec("TRUNCATE TABLE `published_articles`").Error
//	assert.NoError(s.T(), err)
//}

func (s *ArticleHandlerTestSuit) SetupSubTest() {
	err := s.db.Exec("TRUNCATE TABLE `articles`").Error
	assert.NoError(s.T(), err)
	err = s.db.Exec("TRUNCATE TABLE `published_articles`").Error
	assert.NoError(s.T(), err)
}

func (s *ArticleHandlerTestSuit) Test_ArticleHandler_Edit() {

	testCases := []struct {
		name     string
		before   func(t *testing.T)
		after    func(t *testing.T)
		req      web.ArticleReq
		wantCode int
		wantRes  web.Result
	}{
		{
			name: "新建帖子",
			after: func(t *testing.T) {
				var art article.Article
				err := s.db.Where("author_id = ?", 123).First(&art).Error
				assert.NoError(t, err)
				assert.True(t, art.Utime > 0)
				assert.True(t, art.Ctime > 0)
				wantArt := article.Article{
					ID:       art.ID,
					Title:    "这是新建帖子",
					Content:  "123123",
					AuthorID: 123,
					Status:   domain.ArticleStatusUnpublished.ToUint8(),
					Utime:    art.Utime,
					Ctime:    art.Ctime,
				}
				assert.Equal(t, wantArt, art)
			},
			req: web.ArticleReq{
				Title:   "这是新建帖子",
				Content: "123123",
			},
			wantCode: 200,
			wantRes: web.Result{
				Code: 2,
				Msg:  "保存成功",
				Data: float64(1),
			},
		},
		{
			name: "修改帖子",
			before: func(t *testing.T) {
				now := time.Now().UnixMilli()
				art := article.Article{
					ID:       2,
					Title:    "原来的标题",
					Content:  "原来的内容",
					AuthorID: 123,
					Status:   domain.ArticleStatusUnpublished.ToUint8(),
					Utime:    now,
					Ctime:    now,
				}
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				err := s.db.WithContext(ctx).Create(&art).Error
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				var art article.Article
				err := s.db.Where("id = ?", 2).First(&art).Error
				assert.NoError(t, err)
				assert.True(t, art.Utime > 0)
				assert.True(t, art.Ctime > 0)
				assert.True(t, art.Utime > art.Ctime)
				wantArt := article.Article{
					ID:       2,
					Title:    "修改帖子",
					Content:  "hahaha",
					AuthorID: 123,
					Status:   domain.ArticleStatusUnpublished.ToUint8(),
					Utime:    art.Utime,
					Ctime:    art.Ctime,
				}
				assert.Equal(t, wantArt, art)
			},
			req: web.ArticleReq{
				Id:      2,
				Title:   "修改帖子",
				Content: "hahaha",
			},
			wantCode: 200,
			wantRes: web.Result{
				Code: 2,
				Msg:  "保存成功",
				Data: float64(2),
			},
		},
		{
			name: "更新别人的帖子",
			before: func(t *testing.T) {
				// 模拟已经存在的帖子
				s.db.Create(&article.Article{
					ID:      3,
					Title:   "我的标题",
					Content: "我的内容",
					Ctime:   456,
					Utime:   234,
					// 注意。这个 AuthorID 我们设置为另外一个人的ID
					AuthorID: 789,
					Status:   domain.ArticleStatusPublished.ToUint8(),
				})
			},
			after: func(t *testing.T) {
				// 更新应该是失败了，数据没有发生变化
				var art article.Article
				s.db.Where("id = ?", 3).First(&art)
				assert.Equal(t, article.Article{
					ID:       3,
					Title:    "我的标题",
					Content:  "我的内容",
					Ctime:    456,
					Utime:    234,
					AuthorID: 789,
					Status:   domain.ArticleStatusPublished.ToUint8(),
				}, art)
			},
			req: web.ArticleReq{
				Id:      3,
				Title:   "新的标题",
				Content: "新的内容",
			},
			wantCode: 200,
			wantRes: web.Result{
				Code: 5,
				Msg:  "系统错误",
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			t := s.T()
			if tc.before != nil {
				tc.before(t)
			}
			data, err := json.Marshal(tc.req)
			assert.NoError(t, err)
			req, err := http.NewRequest(http.MethodPost, "/articles/edit", bytes.NewReader(data))
			assert.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			recorder := httptest.NewRecorder()
			s.server.ServeHTTP(recorder, req)
			assert.Equal(t, tc.wantCode, recorder.Code)
			if recorder.Code != http.StatusOK {
				return
			}

			var res web.Result
			err = json.Unmarshal(recorder.Body.Bytes(), &res)
			assert.NoError(t, err)
			assert.Equal(t, tc.wantRes, res)
			if tc.after != nil {
				tc.after(t)
			}

		})
	}

}

func (s *ArticleHandlerTestSuit) Test_ArticleHandler_Publish() {
	testCases := []struct {
		name     string
		before   func(t *testing.T)
		after    func(t *testing.T)
		req      web.ArticleReq
		wantCode int
		wantRes  web.Result
	}{
		{
			name: "新建帖子 并发表",
			after: func(t *testing.T) {
				var art article.Article
				err := s.db.Where("id = ?", 1).First(&art).Error
				assert.NoError(t, err)
				assert.True(t, art.Utime > 0)
				assert.True(t, art.Ctime > 0)
				wantArt := article.Article{
					ID:       1,
					Title:    "这是新建帖子",
					Content:  "123123",
					AuthorID: 123,
					Status:   domain.ArticleStatusPublished.ToUint8(),
					Utime:    art.Utime,
					Ctime:    art.Ctime,
				}
				assert.Equal(t, wantArt, art)

				var pubArt article.PublishedArticle
				err = s.db.Where("author_id = ?", 123).First(&pubArt).Error
				assert.NoError(t, err)
				assert.True(t, pubArt.Utime > 0)
				assert.True(t, pubArt.Ctime > 0)
				wantPubArt := article.PublishedArticle{
					ID:       1,
					Title:    "这是新建帖子",
					Content:  "123123",
					AuthorID: 123,
					Status:   domain.ArticleStatusPublished.ToUint8(),
					Utime:    pubArt.Utime,
					Ctime:    pubArt.Ctime,
				}
				assert.Equal(t, wantPubArt, pubArt)
			},
			req: web.ArticleReq{
				Title:   "这是新建帖子",
				Content: "123123",
			},
			wantCode: 200,
			wantRes: web.Result{
				Code: 2,
				Msg:  "发表帖子成功",
				Data: float64(1),
			},
		},
		{
			// 制作库有，但是线上库没有
			name: "更新帖子并新发表",
			before: func(t *testing.T) {
				art := article.Article{
					ID:       2,
					Title:    "原来的标题",
					Content:  "原来的内容",
					AuthorID: 123,
					Status:   domain.ArticleStatusUnpublished.ToUint8(),
					Utime:    123,
					Ctime:    456,
				}
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				err := s.db.WithContext(ctx).Create(&art).Error
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				var art article.Article
				err := s.db.Where("id = ?", 2).First(&art).Error
				assert.NoError(t, err)
				assert.Equal(t, "新的标题", art.Title)
				assert.Equal(t, "新的内容", art.Content)
				assert.Equal(t, int64(123), art.AuthorID)
				// 创建时间没变
				assert.Equal(t, int64(456), art.Ctime)
				// 更新时间变了
				assert.True(t, art.Utime > 234)
				var publishedArt article.PublishedArticle
				s.db.Where("id = ?", 2).First(&publishedArt)
				assert.Equal(t, "新的标题", art.Title)
				assert.Equal(t, "新的内容", art.Content)
				assert.Equal(t, int64(123), art.AuthorID)
				assert.True(t, publishedArt.Ctime > 0)
				assert.True(t, publishedArt.Utime > 0)
			},
			req: web.ArticleReq{
				Id:      2,
				Title:   "新的标题",
				Content: "新的内容",
			},
			wantCode: 200,
			wantRes: web.Result{
				Code: 2,
				Msg:  "发表帖子成功",
				Data: float64(2),
			},
		},
		{
			// 制作库有，但是线上库没有
			name: "更新帖子 并且重新发表",
			before: func(t *testing.T) {
				art := article.Article{
					ID:       3,
					Title:    "原来的标题",
					Content:  "原来的内容",
					AuthorID: 123,
					Status:   domain.ArticleStatusPublished.ToUint8(),
					Utime:    123,
					Ctime:    456,
				}
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				err := s.db.WithContext(ctx).Create(&art).Error
				assert.NoError(t, err)

				pubArt := article.PublishedArticle(art)
				err = s.db.WithContext(ctx).Create(&pubArt).Error
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				var art article.Article
				err := s.db.Where("id = ?", 3).First(&art).Error
				assert.NoError(t, err)
				assert.Equal(t, "新的标题", art.Title)
				assert.Equal(t, "新的内容", art.Content)
				assert.Equal(t, int64(123), art.AuthorID)
				// 创建时间没变
				assert.Equal(t, int64(456), art.Ctime)
				// 更新时间变了
				assert.True(t, art.Utime > 234)
				var publishedArt article.PublishedArticle
				s.db.Where("id = ?", 3).First(&publishedArt)
				assert.Equal(t, "新的标题", art.Title)
				assert.Equal(t, "新的内容", art.Content)
				assert.Equal(t, int64(123), art.AuthorID)
				assert.True(t, publishedArt.Ctime > 0)
				assert.True(t, publishedArt.Utime > 0)
			},
			req: web.ArticleReq{
				Id:      3,
				Title:   "新的标题",
				Content: "新的内容",
			},
			wantCode: 200,
			wantRes: web.Result{
				Code: 2,
				Msg:  "发表帖子成功",
				Data: float64(3),
			},
		},
		{
			name: "更新别人的帖子，并且发表失败",
			before: func(t *testing.T) {
				art := article.Article{
					ID:      4,
					Title:   "我的标题",
					Content: "我的内容",
					Ctime:   456,
					Utime:   234,
					// 注意。这个 AuthorID 我们设置为另外一个人的ID
					AuthorID: 789,
				}
				s.db.Create(&art)
				part := article.PublishedArticle(article.Article{
					ID:       4,
					Title:    "我的标题",
					Content:  "我的内容",
					Ctime:    456,
					Utime:    234,
					AuthorID: 789,
				})
				s.db.Create(&part)
			},
			after: func(t *testing.T) {
				// 更新应该是失败了，数据没有发生变化
				var art article.Article
				s.db.Where("id = ?", 4).First(&art)
				assert.Equal(t, "我的标题", art.Title)
				assert.Equal(t, "我的内容", art.Content)
				assert.Equal(t, int64(456), art.Ctime)
				assert.Equal(t, int64(234), art.Utime)
				assert.Equal(t, int64(789), art.AuthorID)

				var part article.PublishedArticle
				// 数据没有变化
				s.db.Where("id = ?", 4).First(&part)
				assert.Equal(t, "我的标题", part.Title)
				assert.Equal(t, "我的内容", part.Content)
				assert.Equal(t, int64(789), part.AuthorID)
				// 创建时间没变
				assert.Equal(t, int64(456), part.Ctime)
				// 更新时间变了
				assert.Equal(t, int64(234), part.Utime)
			},
			req: web.ArticleReq{
				Id:      4,
				Title:   "新的标题",
				Content: "新的内容",
			},
			wantCode: 200,
			wantRes: web.Result{
				Code: 5,
				Msg:  "系统错误",
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			t := s.T()
			if tc.before != nil {
				tc.before(t)
			}
			data, err := json.Marshal(tc.req)
			assert.NoError(t, err)
			req, err := http.NewRequest(http.MethodPost, "/articles/publish", bytes.NewReader(data))
			assert.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			recorder := httptest.NewRecorder()
			s.server.ServeHTTP(recorder, req)
			assert.Equal(t, tc.wantCode, recorder.Code)
			if recorder.Code != http.StatusOK {
				return
			}

			var res web.Result
			err = json.Unmarshal(recorder.Body.Bytes(), &res)
			assert.NoError(t, err)
			assert.Equal(t, tc.wantRes, res)
			if tc.after != nil {
				tc.after(t)
			}

		})
	}

}

func TestArticle(t *testing.T) {
	suite.Run(t, new(ArticleHandlerTestSuit))
}
