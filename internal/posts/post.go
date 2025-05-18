package posts

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Post struct {
	ID      bson.ObjectID `bson:"_id,omitempty"`
	Title   string        `bson:"title"`
	Content string        `bson:"content"`
	Created time.Time     `bson:"createdAt"`
	Updated time.Time     `bson:"updatedAt"`
}
