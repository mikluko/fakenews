package fakenews

import (
	"context"
	"net/http"
)

type Source interface {
	Fetch(context.Context) error
	Items() []string
}

type Client interface {
	Do(*http.Request) (*http.Response, error)
}
