package lmdrouter

import "time"

type mockListRequest struct {
	ID       string `lambda:"path.id"`
	Page     int64  `lambda:"query.page"`
	PageSize int64  `lambda:"query.page_size"`
}

type mockGetRequest struct {
	ID            string `lambda:"path.id"`
	ShowSomething bool   `lambda:"query.show_something"`
}

type mockPostRequest struct {
	Name string
	Date time.Time
}

type mockItem struct {
	ID   string
	Name string
	Date time.Time
}
