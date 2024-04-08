package web

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"os"

	oidc_login "github.com/dsseng/wiso/pkg/oidc"
	"github.com/gin-gonic/gin"
	"github.com/zitadel/oidc/v3/pkg/oidc"
	"gorm.io/gorm"
)

var (
	//go:embed templates
	embedFS embed.FS
	db      *gorm.DB
	baseURL = url.URL{
		Host:   "localhost:8989",
		Scheme: "http",
	}
)

func processUser(info *oidc.UserInfo, mac string) string {
	// put state as a mac into db
	fmt.Println("logging in", info.PreferredUsername, mac)
	redir := baseURL
	redir.Path = "/welcome"
	query := redir.Query()
	query.Add("username", info.PreferredUsername)
	query.Add("picture", info.Picture)
	redir.RawQuery = query.Encode()
	return redir.String()
}

func setupRouter() *gin.Engine {
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

	// Args: mac of the device which is being authorized
	r.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", gin.H{
			"title":        "Network login",
			"oidcName":     "Gitea",
			"passwordAuth": false,
			"image":        "/static/logo.png",
			"mac":          c.Query("mac"),
		})
	})

	// Args: picture URL and username to be displayed
	r.GET("/welcome", func(c *gin.Context) {
		c.HTML(http.StatusOK, "welcome.html", gin.H{
			"title":    "Success",
			"picture":  c.Query("picture"),
			"username": c.Query("username"),
			"logo":     "/static/logo-welcome.png",
		})
	})

	pr := oidc_login.OIDCProvider{
		ProcessUser:  processUser,
		ClientID:     os.Getenv("CLIENT_ID"),
		ClientSecret: os.Getenv("CLIENT_SECRET"),
		Issuer:       os.Getenv("ISSUER"),
		Name:         "oidc",
		BaseURL:      baseURL,
	}

	err := pr.Setup(r)
	if err != nil {
		panic(err)
	}

	return r
}

func Start(port uint16, database *gorm.DB) {
	db = database
	r := setupRouter()
	r.Run(fmt.Sprintf(":%d", port))
}
