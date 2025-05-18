package posts

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/mineroot/news/internal/paging"
)

type Repository struct {
	collection *mongo.Collection
}

func NewRepository(collection *mongo.Collection) *Repository {
	return &Repository{
		collection: collection,
	}
}

func (r *Repository) FindAllByQueryWithPagination(
	ctx context.Context,
	paginator *paging.Paginator,
	query string,
) ([]*Post, int, error) {
	skip := (paginator.Page() - 1) * paginator.Size()
	var pipeline mongo.Pipeline
	if query != "" {
		pipeline = mongo.Pipeline{
			{{"$match", bson.D{
				{"$text", bson.D{{"$search", query}}}, // filter by query
			}}},
			{{"$sort", bson.D{
				{"score", bson.D{{"$meta", "textScore"}}}, // sort by relevance
			}}},
			{{"$facet", bson.D{ // count total records & apply pagination
				{"metadata", bson.A{
					bson.D{{"$count", "totalCount"}},
				}},
				{"data", bson.A{
					bson.D{{"$skip", skip}},
					bson.D{{"$limit", paginator.Size()}},
				}},
			}}},
		}
	} else {
		pipeline = mongo.Pipeline{
			{{"$sort", bson.D{{"createdAt", -1}}}}, // sort by creation date
			{{"$facet", bson.D{ // count total records & apply pagination
				{"metadata", bson.A{
					bson.D{{"$count", "totalCount"}},
				}},
				{"data", bson.A{
					bson.D{{"$skip", skip}},
					bson.D{{"$limit", paginator.Size()}},
				}},
			}}},
		}
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var aggregatedResults []struct {
		Metadata []struct {
			TotalCount int `bson:"totalCount"`
		} `bson:"metadata"`
		Data []*Post `bson:"data"`
	}

	if err := cursor.All(ctx, &aggregatedResults); err != nil {
		return nil, 0, err
	}

	if len(aggregatedResults) == 0 {
		return []*Post{}, 0, nil
	}

	total := 0
	if len(aggregatedResults[0].Metadata) > 0 {
		total = aggregatedResults[0].Metadata[0].TotalCount
	}

	return aggregatedResults[0].Data, total, nil
}

func (r *Repository) FindById(ctx context.Context, id bson.ObjectID) (*Post, error) {
	var post Post
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&post)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}

	return &post, nil
}

func (r *Repository) Create(ctx context.Context, post *Post) (*Post, error) {
	result, err := r.collection.InsertOne(ctx, post)
	if err != nil {
		return nil, err
	}
	post.ID = result.InsertedID.(bson.ObjectID)

	return post, nil
}

func (r *Repository) UpdateById(ctx context.Context, id bson.ObjectID, title, content string) (*Post, error) {
	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"title":     title,
			"content":   content,
			"updatedAt": time.Now(),
		},
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updatedPost Post
	err := r.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&updatedPost)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}

	return &updatedPost, nil
}

func (r *Repository) DeleteById(ctx context.Context, id bson.ObjectID) (*Post, error) {
	var deletedPost Post
	err := r.collection.FindOneAndDelete(ctx, bson.M{"_id": id}).Decode(&deletedPost)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &deletedPost, nil
}
