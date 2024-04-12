package common

import (
	"net/url"

	"github.com/dsseng/wiso/pkg/users"
)

func ErrorRedirect(baseURL *url.URL, err error, mac string, linkOrig string) string {
	redir := *baseURL
	query := redir.Query()
	query.Add("error", err.Error())
	query.Add("mac", mac)
	query.Add("link-orig", linkOrig)
	redir.RawQuery = query.Encode()
	redir.Path = "/error"
	return redir.String()
}

func WelcomeRedirect(baseURL *url.URL, user users.User, linkOrig string) string {
	redir := *baseURL
	redir.Path = "/welcome"
	query := redir.Query()
	query.Add("username", user.Username)
	query.Add("full_name", user.FullName)
	query.Add("link-orig", linkOrig)
	query.Add("picture", user.Picture)
	redir.RawQuery = query.Encode()
	return redir.String()
}
