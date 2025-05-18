package handlers

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/mineroot/news/internal/paging"
	"github.com/mineroot/news/internal/posts"
)

type PostsPaginator interface {
	FindAllByQueryWithPagination(
		ctx context.Context,
		paginator *paging.Paginator,
		query string,
	) ([]*posts.Post, int, error)
}

type PostFinder interface {
	FindById(ctx context.Context, id bson.ObjectID) (*posts.Post, error)
}

type PostCreator interface {
	Create(ctx context.Context, post *posts.Post) (*posts.Post, error)
}

type PostUpdater interface {
	UpdateById(ctx context.Context, id bson.ObjectID, title, content string) (*posts.Post, error)
}

type PostDeleter interface {
	DeleteById(ctx context.Context, id bson.ObjectID) (*posts.Post, error)
}
