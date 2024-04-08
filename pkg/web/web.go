package web

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"os"

	// "strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zitadel/oidc/v3/pkg/client/rp"

	// httphelper "github.com/zitadel/oidc/v3/pkg/http"
	"github.com/zitadel/oidc/v3/pkg/oidc"
	"gorm.io/gorm"
)

var (
	//go:embed templates
	embedFS embed.FS
	db      *gorm.DB
	baseUrl = "http://localhost:8989"
)

func processUser(info *oidc.UserInfo, mac string) string {
	// put state as a mac into db
	fmt.Println("logging in", info.PreferredUsername, mac)
	ur, err := url.Parse(baseUrl)
	if err != nil {
		panic(err)
	}
	ur.Path = "/welcome"
	query := ur.Query()
	query.Add("username", info.PreferredUsername)
	query.Add("picture", info.Picture)
	ur.RawQuery = query.Encode()
	return ur.String()
}

func setupRouter() *gin.Engine {
	clientID := os.Getenv("CLIENT_ID")
	clientSecret := os.Getenv("CLIENT_SECRET")
	issuer := os.Getenv("ISSUER")
	redirectURI := fmt.Sprintf("%v%v", baseUrl, "/oidc/callback")

	options := []rp.Option{
		rp.WithVerifierOpts(rp.WithIssuedAtOffset(5 * time.Second)),
	}
	provider, err := rp.NewRelyingPartyOIDC(
		context.TODO(),
		issuer,
		clientID,
		clientSecret,
		redirectURI,
		[]string{"openid", "profile"},
		options...,
	)
	if err != nil {
		fmt.Printf("error creating provider %s", err.Error())
	}

	r := gin.Default()
	r.SetTrustedProxies([]string{})
	templ := template.Must(
		template.
			New("").
			ParseFS(embedFS, "templates/*.html"),
	)
	r.SetHTMLTemplate(templ)
	staticFS := http.FS(embedFS)

	r.GET("/static/:path", func(c *gin.Context) {
		fmt.Println(c.Param("path"))
		c.FileFromFS("templates/static/"+c.Param("path"), staticFS)
	})

	r.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", gin.H{
			"title":        "Network login",
			"oidcName":     "Gitea",
			"passwordAuth": false,
			"image":        "/static/logo.png",
			"mac":          c.Query("mac"),
		})
	})

	r.GET("/welcome", func(c *gin.Context) {
		c.HTML(http.StatusOK, "welcome.html", gin.H{
			"title":    "Success",
			"picture":  c.Query("picture"),
			"username": c.Query("username"),
			"logo":     "/static/logo-welcome.png",
		})
	})

	marshalUserinfo := func(w http.ResponseWriter, r *http.Request, tokens *oidc.Tokens[*oidc.IDTokenClaims], state string, rp rp.RelyingParty, info *oidc.UserInfo) {
		redir := processUser(info, state)
		w.Header().Add("Location", redir)
		w.WriteHeader(http.StatusSeeOther)

		data, err := json.Marshal(info)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Write(data)
	}

	r.GET("/oidc/login", gin.WrapF(func(w http.ResponseWriter, r *http.Request) {
		state := func() string {
			return r.URL.Query().Get("mac")
		}
		(rp.AuthURLHandler(
			state,
			provider,
		))(w, r)
	}))
	r.GET("/oidc/callback", gin.WrapF(rp.CodeExchangeHandler(rp.UserinfoCallback(marshalUserinfo), provider)))

	return r
}

func Start(port uint16, database *gorm.DB) {
	db = database
	r := setupRouter()
	r.Run(fmt.Sprintf(":%d", port))
}
