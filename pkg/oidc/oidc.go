package oidc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/dsseng/wiso/pkg/common"
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

func (p OIDCProvider) processUser(info *oidc.UserInfo, mac string, linkOrig string) string {
	username := info.PreferredUsername + "@" + p.ID

	user, err := users.FindSingle(p.DB, username)
	if err != nil {
		return common.ErrorRedirect(p.BaseURL, err, mac, linkOrig)
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
			return common.ErrorRedirect(p.BaseURL, res.Error, mac, linkOrig)
		}
	} else if user[0].FullName != info.Name || user[0].Picture != info.Picture {
		user[0].FullName = info.Name
		user[0].Picture = info.Picture
		p.DB.Save(user[0])
	}

	if mac != "" {
		err := users.StartSession(p.DB, user[0], mac, time.Now().Add(time.Hour*168))
		if err != nil {
			return common.ErrorRedirect(p.BaseURL, err, mac, linkOrig)
		}
	}

	return common.WelcomeRedirect(p.BaseURL, user[0], linkOrig)
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
			redir := common.ErrorRedirect(
				p.BaseURL,
				fmt.Errorf("Unknown auth state"),
				"",
				"",
			)
			w.Header().Add("Location", redir)
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
