package lmdrouter

import "time"

type mockConst string

const (
	mockConstTwo mockConst = "two"
)

type Number string

const (
	numberOne   Number = "one"
	numberTwo   Number = "two"
	numberThree Number = "three"
)

type mockListReq struct {
	ID         string              `lambda:"path.id"`
	Page       int64               `lambda:"query.page"`
	PageSize   int64               `lambda:"query.page_size"`
	Terms      []string            `lambda:"query.terms"`
	Numbers    []float64           `lambda:"query.numbers"`
	Const      mockConst           `lambda:"query.const"`
	Bool       bool                `lambda:"query.bool"`
	PBoolOne   *bool               `lambda:"query.pbool1"`
	PBoolTwo   *bool               `lambda:"query.pbool2"`
	Time       *time.Time          `lambda:"query.time"`
	Alias      stringAliasExample  `lambda:"query.alias"`
	AliasPtr   *stringAliasExample `lambda:"query.alias_ptr"`
	CommaSplit []Number            `lambda:"query.commaSplit"`
	Language   string              `lambda:"header.Accept-Language"`
	Encoding   []string            `lambda:"header.Accept-Encoding"`
}

type mockGetReq struct {
	ID            string `lambda:"path.id"`
	ShowSomething bool   `lambda:"query.show_something"`
}

type mockPostReq struct {
	ID   string    `lambda:"path.id"`
	Name string    `json:"name"`
	Date time.Time `json:"date"`
}

type mockItem struct {
	ID   string
	Name string
	Date time.Time
}
