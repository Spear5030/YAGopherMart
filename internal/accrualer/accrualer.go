package accrualer

import (
	"net/http"
)

type Client struct {
	Client *http.Client
	DSN    string
}

func New(dsn string) *Client {
	return &Client{
		Client: http.DefaultClient,
		DSN:    dsn,
	}
}
