package utils

import (
	"net/http"
)

// https://api.github.com/repos/sstark-mason/iceslab/releases/latest
// https://api.github.com/repos/sstark-mason/iceslab/releases/tags/bookmarks-latest

const (
	owner                     = "sstark-mason"
	repo                      = "iceslab"
	branch                    = "main"
	latestBookmarksReleaseURL = "https://github.com/sstark-mason/iceslab/releases/download/bookmarks-latest/bookmarks.zip"
	latestSourceURL           = "https://github.com/sstark-mason/iceslab/zipball/main"
)

type Client struct {
	http   *http.Client
	apiURL string
	token  string
}

func NewClient(token string) *Client {
	return &Client{
		http:   &http.Client{},
		apiURL: "https://api.github.com",
		token:  token,
	}
}
