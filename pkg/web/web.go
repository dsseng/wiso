package web

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var (
	db *gorm.DB
	//go:embed templates/*
	embedFS embed.FS
)

func setupRouter() *gin.Engine {
	r := gin.Default()
	r.SetTrustedProxies([]string{})
	templ := template.Must(
		template.
			New("").
			ParseFS(embedFS, "templates/*"),
	)
	r.SetHTMLTemplate(templ)

	r.GET("login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", gin.H{
			"title":        "Network login",
			"oidcName":     "Gitea",
			"passwordAuth": false,
		})
	})
	r.GET("style.css", func(c *gin.Context) {
		file, _ := embedFS.ReadFile("templates/style.css")
		c.Data(
			http.StatusOK,
			"text/css",
			file,
		)
	})
	r.GET("bulma.min.css", func(c *gin.Context) {
		file, _ := embedFS.ReadFile("templates/bulma.min.css")
		c.Data(
			http.StatusOK,
			"text/css",
			file,
		)
	})

	return r
}

func Start(port uint16, database *gorm.DB) {
	db = database
	r := setupRouter()
	r.Run(fmt.Sprintf(":%d", port))
}
