# gorm-pagination

Pagination with gorm and net/http

[![Actions Status](https://github.com/maruware/gorm-pagination/workflows/Test/badge.svg)](https://github.com/maruware/gorm-pagination/actions)

## Usage

Example (Error handling is omitted)

```go
package main

import (
    "net/http"
    "github.com/jinzhu/gorm"
	pagination "github.com/maruware/gorm-pagination"
)

type Sample struct {
	db *gorm.DB
}
func (s *Sample) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    q := s.db.Model(Post{})
    var total uint
    q, _ := pagination.PagenateWithContext(r.Context(), q, &total)

    var posts []Post
    q.Find(&posts)
    
    res := Res{
		Total: total,
		Posts: posts,
	}
    b, _ := json.Marshal(res)
    w.WriteHeader(http.StatusOK)
    w.Write(b)
}

func main() {
    s := &Sample{db}
    h := pagination.Middleware(s)

    http.ListenAndServe(":3000", h)
}


```

And request with below queries

| Key | Value | Type | Desc |
|------|-----|-------|-------|
| sort | "["title", "DESC"]" | Encoding []string to json | Query with ORDER |
| range | "[10, 20]" | Encoding []int to json | Query with OFFSET and LIMIT |
| filter | "[{"column": "title", "op": "contains", "values": ["abc"]}]" | Encoding Filter to json | Query with WHERE | 


For details [pagination_test.go](pagination_test.go)

