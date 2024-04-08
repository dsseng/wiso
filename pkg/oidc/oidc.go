package oidc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zitadel/oidc/v3/pkg/client/rp"
	"github.com/zitadel/oidc/v3/pkg/oidc"
)

type OIDCProvider struct {
	ClientID     string
	ClientSecret string
	Issuer       string
	BaseURL      *url.URL
	Name         string

	rp rp.RelyingParty
}

func (p OIDCProvider) processUser(info *oidc.UserInfo, mac string) string {
	// put state as a mac into db
	fmt.Println("logging in", info.PreferredUsername, mac)
	redir := p.BaseURL
	redir.Path = "/welcome"
	query := redir.Query()
	query.Add("username", info.PreferredUsername)
	query.Add("picture", info.Picture)
	redir.RawQuery = query.Encode()
	return redir.String()
}

func (p OIDCProvider) Setup(r *gin.Engine) error {
	callbackPath := "/" + p.Name + "/callback"
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
		redir := p.processUser(info, state)
		w.Header().Add("Location", redir)
		w.WriteHeader(http.StatusSeeOther)

		data, err := json.Marshal(info)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Write(data)
	}

	r.GET("/"+p.Name+"/login", gin.WrapF(func(w http.ResponseWriter, r *http.Request) {
		state := func() string {
			return r.URL.Query().Get("mac")
		}
		(rp.AuthURLHandler(
			state,
			provider,
		))(w, r)
	}))
	r.GET(callbackPath, gin.WrapF(rp.CodeExchangeHandler(rp.UserinfoCallback(marshalUserinfo), provider)))

	return err
}