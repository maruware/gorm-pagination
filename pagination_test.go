package pagination_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	pagination "github.com/maruware/gorm-pagination"
)

type Post struct {
	ID      uint   `gorm:"primary_key" json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

func buildDatabase() (*gorm.DB, error) {
	db, err := gorm.Open("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	db.AutoMigrate(&Post{})

	return db, nil
}

type getPosts struct {
	db *gorm.DB
}

type getPostsRes struct {
	Total uint `json:"total"`
	Posts []Post
}

func (h *getPosts) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	q := h.db.Model(Post{})
	var total uint
	q, err := pagination.PagenateWithContext(r.Context(), q, &total)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var posts []Post
	if err := q.Find(&posts).Error; err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	res := getPostsRes{
		Total: total,
		Posts: posts,
	}

	b, err := json.Marshal(res)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

func buildHttp(db *gorm.DB) http.Handler {
	h := &getPosts{db}

	return pagination.Middleware(h)
}

func createInitialData(db *gorm.DB, num int) error {
	for i := 0; i < num; i++ {
		post := Post{Title: fmt.Sprintf("title%d", i), Content: "content"}
		if err := db.Create(&post).Error; err != nil {
			return err
		}
	}
	return nil
}

func TestIntegrate(t *testing.T) {
	db, err := buildDatabase()
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}

	handler := buildHttp(db)

	num := 300
	if err := createInitialData(db, num); err != nil {
		t.Fatalf("failed to create: %v", err)
	}

	q := url.Values{}

	// Pagination params
	sort, _ := json.Marshal(pagination.SortParam{"title", "desc"})
	q.Set("sort", string(sort))

	range_, _ := json.Marshal(pagination.RangeParam{10, 20})
	q.Set("range", string(range_))

	filters, _ := json.Marshal(pagination.FilterParam{pagination.Filter{
		Op:     pagination.FilterOpContains,
		Column: "title",
		Values: []interface{}{"1"},
	}})
	q.Set("filter", string(filters))

	u, _ := url.Parse("/")
	u.RawQuery = q.Encode()

	// request
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", u.String(), nil)

	handler.ServeHTTP(w, r)

	if err != nil {
		t.Fatalf("failed to get posts; %v", err)
	}
	if c := w.Result().StatusCode; c != http.StatusOK {
		t.Fatalf("failed to get posts; status = %d", c)
	}

	var data getPostsRes
	dec := json.NewDecoder(w.Result().Body)
	if err := dec.Decode(&data); err != nil {
		t.Fatal("bad response")
	}

	// 1, 10-19, 21, 31, ..., 100-199, 201...
	expectTotal := 138
	if data.Total != uint(expectTotal) {
		t.Errorf("expect total is %d, but %d", expectTotal, data.Total)
	}
	if len(data.Posts) != 10 {
		t.Errorf("expect to respond %d posts, but %d", 10, len(data.Posts))
	}
}
