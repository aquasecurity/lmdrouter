package lmdrouter

import "time"

type mockListRequest struct {
	ID       string `lambda:"path.id"`
	Page     int64  `lambda:"query.page"`
	PageSize int64  `lambda:"query.page_size"`
	Language string `lambda:"header.Accept-Language"`
}

type mockGetRequest struct {
	ID            string `lambda:"path.id"`
	ShowSomething bool   `lambda:"query.show_something"`
}

type mockPostRequest struct {
	Name string    `json:"name"`
	Date time.Time `json:"date"`
}

type mockItem struct {
	ID   string
	Name string
	Date time.Time
}
