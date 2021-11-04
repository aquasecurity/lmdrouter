package lmdrouter

import "time"

type mockConst string

const (
	mockConstOne mockConst = "one"
	mockConstTwo mockConst = "two"
)

type mockListRequest struct {
	ID       string     `lambda:"path.id"`
	Page     int64      `lambda:"query.page"`
	PageSize int64      `lambda:"query.page_size"`
	Terms    []string   `lambda:"query.terms"`
	Numbers  []float64  `lambda:"query.numbers"`
	Const    mockConst  `lambda:"query.const"`
	Bool     bool       `lambda:"query.bool"`
	PBoolOne *bool      `lambda:"query.pbool1"`
	PBoolTwo *bool      `lambda:"query.pbool2"`
	Time     *time.Time `lambda:"query.time"`
	Language string     `lambda:"header.Accept-Language"`
	Encoding []string   `lambda:"header.Accept-Encoding"`
}

type mockGetRequest struct {
	ID            string `lambda:"path.id"`
	ShowSomething bool   `lambda:"query.show_something"`
}

type mockPostRequest struct {
	ID   string    `lambda:"path.id"`
	Name string    `json:"name"`
	Date time.Time `json:"date"`
}

type mockItem struct {
	ID   string
	Name string
	Date time.Time
}
