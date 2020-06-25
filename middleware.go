package pagination

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type SortParam []string

type RangeParam []int
type FilterParam []Filter

type contextKey struct {
	name string
}

func (k *contextKey) String() string {
	return "pagenation " + k.name
}

var (
	SortCtxKey   = &contextKey{"sort"}
	OffsetCtxKey = &contextKey{"offset"}
	LimitCtxKey  = &contextKey{"limit"}
	FilterCtxKey = &contextKey{"filter"}
)

func Middleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		q := r.URL.Query()

		// sort
		s := q.Get("sort")
		var sort SortParam
		if s != "" {
			d := json.NewDecoder(strings.NewReader(s))
			err := d.Decode(&sort)
			if err != nil {
				http.Error(w, "Bad sort format", http.StatusBadRequest)
				return
			}
			if len(sort) != 2 {
				http.Error(w, "Bad sort format", http.StatusBadRequest)
				return
			}
			sortStr := fmt.Sprintf("%v %v", sort[0], sort[1])
			ctx = context.WithValue(ctx, SortCtxKey, sortStr)
		}

		// range
		rn := q.Get("range")
		var rnge RangeParam

		if rn != "" {
			d := json.NewDecoder(strings.NewReader(rn))
			err := d.Decode(&rnge)
			if err != nil {
				http.Error(w, "Bad range format", http.StatusBadRequest)
				return
			}

			offset := rnge[0]
			limit := rnge[1] - rnge[0]
			ctx = context.WithValue(ctx, OffsetCtxKey, offset)
			ctx = context.WithValue(ctx, LimitCtxKey, limit)
		}

		// filter
		f := q.Get("filter")
		var filter FilterParam
		if f != "" {
			d := json.NewDecoder(strings.NewReader(f))
			err := d.Decode(&filter)
			if err != nil {
				http.Error(w, "Bad filter format", http.StatusBadRequest)
				return
			}
			ctx = context.WithValue(ctx, FilterCtxKey, filter)
		}

		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}
