package infra

import (
	"net/http"
	"net/url"
)

// Login logins to the bookinfo app and returns the cookies
func Login(rawURL, user, password string) (string, []*http.Cookie, error) {
	form := url.Values{}
	form["username"] = []string{user}
	form["passwd"] = []string{password}

	// follow=false, else the cookies returned are from the redirect, not from the login
	return PostBody(rawURL, form, false)
}
