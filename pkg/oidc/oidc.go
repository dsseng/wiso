package oidc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/dsseng/wiso/pkg/radius"
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
	// TODO: Factor out find or create
	username := info.PreferredUsername + "@" + p.ID
	user := []users.User{{}}
	res := p.DB.Limit(1).Find(&user, "username = ?", username)
	if res.Error != nil {
		fmt.Println("A DB error occured", res.Error)
		redir := p.BaseURL
		query := redir.Query()
		query.Add("error", res.Error.Error())
		redir.RawQuery = query.Encode()
		redir.Path = "/error"
		return redir.String()
	}

	if res.RowsAffected == 0 {
		// TODO: Groups?
		user = []users.User{
			{
				Username: username,
				FullName: info.Name,
				Picture:  info.Picture,
			},
		}
	}

	if mac != "" {
		// TODO: Factor out login
		radcheck := radius.RadCheck{
			Username:  mac,
			Attribute: "Cleartext-Password",
			Op:        ":=",
			Value:     "macauth",
		}
		res = p.DB.Create(&radcheck)
		if res.Error != nil {
			fmt.Println("A DB error occured", res.Error)
			redir := p.BaseURL
			redir.Path = "/error"
			query := redir.Query()
			query.Add("error", res.Error.Error())
			redir.RawQuery = query.Encode()
			return redir.String()
		}

		user[0].DeviceSessions = append(user[0].DeviceSessions, users.DeviceSession{
			DueDate:    time.Now().Add(time.Hour * 168),
			RadcheckID: radcheck.ID,
			MAC:        mac,
		})
	}
	res = p.DB.Save(user)
	if res.Error != nil {
		fmt.Println("A DB error occured", res.Error)
		redir := p.BaseURL
		redir.Path = "/error"
		return redir.String()
	}

	redir := p.BaseURL
	redir.Path = "/welcome"
	query := redir.Query()
	query.Add("username", info.PreferredUsername)
	query.Add("full_name", info.Name)
	query.Add("link-orig", linkOrig)
	query.Add("picture", info.Picture)
	redir.RawQuery = query.Encode()
	return redir.String()
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
