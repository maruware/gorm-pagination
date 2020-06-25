package pagination

import (
	"context"
	"fmt"

	"github.com/jinzhu/gorm"
)

func Paginate(db *gorm.DB, sort string, offset int, limit int, filters []Filter, total *uint) (*gorm.DB, error) {
	q := db
	if sort != "" {
		q = q.Order(sort)
	}
	for _, filter := range filters {
		if filter.Op == FilterOpEqual {
			if len(filter.Values) != 1 {
				return nil, fmt.Errorf("FilterOpEqual requires 1 value. values=%v", filter.Values)
			}
			q = q.Where(fmt.Sprintf("%s = ?", filter.Column), filter.Values[0])
		} else if filter.Op == FilterOpContains {
			if len(filter.Values) != 1 {
				return nil, fmt.Errorf("FilterOpEqual requires 1 values. values=%v", filter.Values)
			}
			q = q.Where(fmt.Sprintf("%s LIKE ?", filter.Column), fmt.Sprintf("%%%v%%", filter.Values[0]))
		} else if filter.Op == FilterOpBetween {
			if len(filter.Values) != 2 {
				return nil, fmt.Errorf("FilterOpEqual requires 2 values. values=%v", filter.Values)
			}
			if filter.Values[0] != nil {
				q = q.Where(fmt.Sprintf("%s >= ?", filter.Column), filter.Values[0])
			}
			if filter.Values[1] != nil {
				q = q.Where(fmt.Sprintf("%s <= ?", filter.Column), filter.Values[1])
			}
		} else if filter.Op == FilterOpIn {
			q = q.Where(fmt.Sprintf("%s IN (?)", filter.Column), filter.Values)
		} else if filter.Op == FilterOpNull {
			q = q.Where(fmt.Sprintf("%s IS NULL", filter.Column))
		} else if filter.Op == FilterOpNotNull {
			q = q.Where(fmt.Sprintf("%s IS NOT NULL", filter.Column))
		} else {
			return nil, fmt.Errorf("Bad filter Op. %v", filter.Op)
		}
	}

	q.Count(total)

	if offset > 0 {
		q = q.Offset(offset)
	}
	if limit > 0 {
		q = q.Limit(limit)
	}
	return q, nil
}

func PagenateWithContext(ctx context.Context, db *gorm.DB, total *uint) (*gorm.DB, error) {
	sortV := ctx.Value(SortCtxKey)
	sort, ok := sortV.(string)
	if !ok {
		sort = ""
	}
	offsetV := ctx.Value(OffsetCtxKey)
	offset, ok := offsetV.(int)
	if !ok {
		offset = 0
	}
	limitV := ctx.Value(LimitCtxKey)
	limit, ok := limitV.(int)
	if !ok {
		limit = 0
	}

	f := ctx.Value(FilterCtxKey)
	filters, ok := f.(FilterParam)
	if !ok {
		filters = FilterParam{}
	}

	return Paginate(db, sort, offset, limit, filters, total)
}
