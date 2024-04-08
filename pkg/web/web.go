package web

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"os"

	"github.com/dsseng/wiso/pkg/oidc"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var (
	//go:embed templates
	embedFS embed.FS
	db      *gorm.DB
	baseURL *url.URL
)

func setupRouter() (*gin.Engine, error) {
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

	pr := oidc.OIDCProvider{
		ClientID:     os.Getenv("CLIENT_ID"),
		ClientSecret: os.Getenv("CLIENT_SECRET"),
		Issuer:       os.Getenv("ISSUER"),
		Name:         "oidc",
		BaseURL:      baseURL,
	}

	err := pr.Setup(r)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func Start(baseUrl string, database *gorm.DB) error {
	var err error
	baseURL, err = url.Parse(baseUrl)
	if err != nil {
		return err
	}

	db = database

	r, err := setupRouter()
	if err != nil {
		return err
	}

	r.Run(":" + baseURL.Port())

	return nil
}
