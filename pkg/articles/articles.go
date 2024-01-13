package articles

type Article struct {
}

type ArticleRepository interface {
	FindById(int64) (Article, error)
	FindByBarcode(string) (Article, error)
	StoreArticle(Article) error
	UpdateArticle(Article) error
	DeleteById(int64) error
}

type Service struct {
	articleRepo ArticleRepository
}
