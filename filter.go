package pagination

type FilterOp string

const (
	FilterOpEqual    = "equal"
	FilterOpContains = "contains"
	FilterOpBetween  = "between"
	FilterOpIn       = "in"
	FilterOpNull     = "null"
	FilterOpNotNull  = "not_null"
)

type Filter struct {
	Column string        `json:"column"`
	Op     FilterOp      `json:"op"`
	Values []interface{} `json:"values"`
}
