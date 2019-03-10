package infra

import (
	"net/http"
	"net/url"
)

// Login logins to the bookinfo app and returns the cookies
func Login(rawURL, user, password string) ([]*http.Cookie, error) {
	form := url.Values{}
	form["username"] = []string{user}
	form["passwd"] = []string{password}

	_, cs, err := PostBody(rawURL, form)
	return cs, err
}
