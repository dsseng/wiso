package oidc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/dsseng/wiso/pkg/users"
	"github.com/gin-gonic/gin"
	"github.com/zitadel/oidc/v3/pkg/client/rp"
	"github.com/zitadel/oidc/v3/pkg/oidc"
	"gorm.io/gorm"
)

type OIDCProvider struct {
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
	Issuer       string `yaml:"issuer"`
	BaseURL      *url.URL
	ID           string
	Name         string
	DB           *gorm.DB

	rp rp.RelyingParty
}

func (p OIDCProvider) errorRedirect(err error) string {
	redir := p.BaseURL
	query := redir.Query()
	query.Add("error", err.Error())
	redir.RawQuery = query.Encode()
	redir.Path = "/error"
	return redir.String()
}

func (p OIDCProvider) welcomeRedirect(user users.User, linkOrig string) string {
	redir := p.BaseURL
	redir.Path = "/welcome"
	query := redir.Query()
	query.Add("username", user.Username)
	query.Add("full_name", user.FullName)
	query.Add("link-orig", linkOrig)
	query.Add("picture", user.Picture)
	redir.RawQuery = query.Encode()
	return redir.String()
}

func (p OIDCProvider) processUser(info *oidc.UserInfo, mac string, linkOrig string) string {
	username := info.PreferredUsername + "@" + p.ID

	user, err := users.FindSingle(p.DB, username)
	if err != nil {
		return p.errorRedirect(err)
	}

	if len(user) == 0 {
		user = []users.User{
			{
				Username: username,
				FullName: info.Name,
				Picture:  info.Picture,
			},
		}

		res := p.DB.Create(user)
		if res.Error != nil {
			return p.errorRedirect(res.Error)
		}
	}

	if mac != "" {
		err := users.StartSession(p.DB, user[0], mac, time.Now().Add(time.Hour*168))
		if err != nil {
			return p.errorRedirect(err)
		}
	}

	fmt.Println(linkOrig)
	return p.welcomeRedirect(user[0], linkOrig)
}

func (p OIDCProvider) Setup(r *gin.Engine) error {
	callbackPath := "/" + p.ID + "/callback"
	redirectURI := p.BaseURL
	redirectURI.Path = callbackPath

	options := []rp.Option{
		rp.WithVerifierOpts(rp.WithIssuedAtOffset(5 * time.Second)),
	}
	provider, err := rp.NewRelyingPartyOIDC(
		context.TODO(),
		p.Issuer,
		p.ClientID,
		p.ClientSecret,
		redirectURI.String(),
		[]string{"openid", "profile"},
		options...,
	)
	if err != nil {
		return err
	}

	marshalUserinfo := func(w http.ResponseWriter, r *http.Request, tokens *oidc.Tokens[*oidc.IDTokenClaims], state string, rp rp.RelyingParty, info *oidc.UserInfo) {
		pos := strings.Index(state, "^")
		if pos == -1 {
			fmt.Println("Unknown auth state")
			redir := p.BaseURL
			redir.Path = "/error"
			query := redir.Query()
			query.Add("error", "Unknown auth state")
			redir.RawQuery = query.Encode()
			w.Header().Add("Location", redir.String())
			w.WriteHeader(http.StatusSeeOther)
			return
		}
		redir := p.processUser(info, state[:pos], state[pos+1:])
		w.Header().Add("Location", redir)
		w.WriteHeader(http.StatusSeeOther)

		data, err := json.Marshal(info)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Write(data)
	}

	r.GET("/"+p.ID+"/login", gin.WrapF(func(w http.ResponseWriter, r *http.Request) {
		state := func() string {
			return r.URL.Query().Get("mac") + "^" + r.URL.Query().Get("link-orig")
		}
		(rp.AuthURLHandler(
			state,
			provider,
		))(w, r)
	}))
	r.GET(callbackPath, gin.WrapF(rp.CodeExchangeHandler(rp.UserinfoCallback(marshalUserinfo), provider)))

	return err
}
