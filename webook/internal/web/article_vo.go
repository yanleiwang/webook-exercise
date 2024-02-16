package web

type ArticleVo struct {
	Id    int64  `json:"id"`
	Title string `json:"title"`
	// 摘要
	Abstract string `json:"abstract"`
	// 内容
	Content string `json:"content"`
	Status  uint8  `json:"status"`
	Author  string `json:"author"`
	Ctime   string `json:"ctime"`
	Utime   string `json:"utime"`
}

type ArticleReq struct {
	Id      int64  `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}
