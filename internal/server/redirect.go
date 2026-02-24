package server

import "net/url"

func buildSearchURL(query string) string {
	return "https://google.com/search?q=" + url.QueryEscape(query)
}
