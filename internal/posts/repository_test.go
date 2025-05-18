package posts_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/mineroot/news/internal/db"
	"github.com/mineroot/news/internal/paging"
	"github.com/mineroot/news/internal/posts"
)

var mongoClient *mongo.Client
var repo *posts.Repository

func TestMain(m *testing.M) {
	cleanup, err := testMain()
	if err != nil {
		cleanup()
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	code := m.Run()
	cleanup()
	os.Exit(code)
}

func TestRepository_CRUD(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	now := time.Now().In(time.UTC).Truncate(time.Millisecond)
	oid := bson.NewObjectIDFromTimestamp(now)

	// check post with defined id does not exist
	notFound, err := repo.FindById(ctx, oid)
	require.NoError(t, err)
	assert.Nil(t, notFound)

	// create post with defined id
	expected := &posts.Post{
		ID:      oid,
		Title:   "Test Post",
		Content: "Some Content\nLorem Ipsum",
		Created: now,
		Updated: now,
	}
	expected, err = repo.Create(ctx, expected)
	require.NoError(t, err)

	// check post with defined id exist
	actual, err := repo.FindById(ctx, expected.ID)
	require.NoError(t, err)
	assert.Equal(t, expected, actual)

	// update post with defined id
	actual, err = repo.UpdateById(ctx, oid, "Updated Title", "Updated Content\nLorem Ipsum")
	require.NoError(t, err)
	assert.Equal(t, expected.ID, actual.ID)
	assert.Equal(t, "Updated Title", actual.Title)
	assert.Equal(t, "Updated Content\nLorem Ipsum", actual.Content)
	assert.Greater(t, actual.Updated, actual.Created) // assert 'Updated' field has changed

	// delete post with defined id
	actual, err = repo.DeleteById(ctx, oid)
	require.NoError(t, err)
	assert.NotNil(t, actual)

	// check post with defined id does not exist
	notFound, err = repo.FindById(ctx, oid)
	require.NoError(t, err)
	assert.Nil(t, notFound)
}

func TestRepository_FindAllByQueryWithPagination(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	const pageSize = 5
	const expectedTotalPosts = 16

	// ChatGPT generated dummy posts...
	dummyPosts := []*posts.Post{
		{Title: "Echoes of Tomorrow", Content: "Flickering socket flame"},
		{Title: "Hidden Binary Paths", Content: "Null response echo"},
		{Title: "Rusty Feather Code", Content: "Patchy route sketch"},
		{Title: "Quantum Pebble Notes", Content: "Overload pattern haze"},
		{Title: "Shadows on Protocol", Content: "Loop under silence"},
		{Title: "Crimson Logic Thread", Content: "Core panic whisper"},
		{Title: "Idle Syntax Bloom", Content: "Stale commit breeze"},
		{Title: "Dreams in JSON", Content: "Shifting frame lock"},
		{Title: "Silent Kernel Drift", Content: "Memory drift field"},
		{Title: "Neon Stack Mirage", Content: "Shadow ping fall"},
		{Title: "Go Routine Whispers", Content: "Token rain drop"},
		{Title: "Muted Index Flood", Content: "Cold thread slice"},
		{Title: "Broken Cache Signals", Content: "Latency wind trace"},
		{Title: "Phantom Token Leap", Content: "Opaque signal hum"},
		{Title: "Pixel Trace Vault", Content: "Idle frame burst"},
		{Title: "Abstract Port Pulse", Content: "Sync ghost fog"},
	}
	for _, post := range dummyPosts {
		now := time.Now().In(time.UTC).Truncate(time.Millisecond)
		post.Created = now
		post.Updated = now
		_, err := repo.Create(ctx, post)
		require.NoError(t, err)
	}

	// fetch first page
	paginator, err := paging.NewPaginator("1", pageSize)
	require.NoError(t, err)
	foundPosts, actualTotalPosts, err := repo.FindAllByQueryWithPagination(ctx, paginator, "")
	require.NoError(t, err)
	assert.Equal(t, expectedTotalPosts, actualTotalPosts) // check total posts count
	assert.Len(t, foundPosts, pageSize)                   // assert page is "full"
	paginator.SetRealItemsCount(actualTotalPosts)
	assert.Equal(t, 1, paginator.Page())
	assert.False(t, paginator.NeedsRedirect())

	// fetch 99th page, which does not exist
	paginator, err = paging.NewPaginator("99", pageSize)
	require.NoError(t, err)
	foundPosts, actualTotalPosts, err = repo.FindAllByQueryWithPagination(ctx, paginator, "")
	require.NoError(t, err)
	assert.Equal(t, expectedTotalPosts, actualTotalPosts) // should be the same
	assert.Len(t, foundPosts, 0)                          // should be 0 as 99th page is too large
	paginator.SetRealItemsCount(actualTotalPosts)
	assert.Equal(t, 4, paginator.Page())      // "the real last page" should be 4 as ceil(16 / 5) = 4
	assert.True(t, paginator.NeedsRedirect()) // so we should redirect to page 4

	// search
	paginator, err = paging.NewPaginator("1", pageSize)
	require.NoError(t, err)
	foundPosts, actualTotalPosts, err = repo.FindAllByQueryWithPagination(ctx, paginator, "echo")
	require.NoError(t, err)
	require.Len(t, foundPosts, 2) // as only first two posts match search query "echo"
	assert.Equal(t, dummyPosts[0], foundPosts[0])
	assert.Equal(t, dummyPosts[1], foundPosts[1])

	// cleanup
	_, err = mongoClient.Database(db.Name).Collection(db.PostsCollection).DeleteMany(ctx, bson.M{})
	require.NoError(t, err)
}

func testMain() (func(), error) {
	cleanup := func() {}
	pool, err := dockertest.NewPool("")
	if err != nil {
		return cleanup, fmt.Errorf("could not connect to docker: %w", err)
	}

	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "mongo",
		Tag:        "8.0.9",
		Env:        []string{"TZ=Europe/Kyiv"},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		return cleanup, fmt.Errorf("could not start resource: %w", err)
	}
	cleanup = func() { _ = pool.Purge(resource) }

	mongoClient, err = db.CreateMongoClient(context.Background(), fmt.Sprintf("mongodb://localhost:%s", resource.GetPort("27017/tcp")))
	if err != nil {
		return cleanup, fmt.Errorf("could not connect to docker mongodb: %w", err)
	}

	repo = posts.NewRepository(db.GetPostsCollection(mongoClient))
	return cleanup, nil
}
