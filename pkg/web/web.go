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
	//go:embed templates
	embedFS embed.FS
)

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

	r.GET("login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", gin.H{
			"title":        "Network login",
			"oidcName":     "Gitea",
			"passwordAuth": false,
			"image":        "/static/logo.png",
		})
	})
	r.GET("/static/:path", func(c *gin.Context) {
		fmt.Println(c.Param("path"))
		c.FileFromFS("templates/static/"+c.Param("path"), staticFS)
	})

	return r
}

func Start(port uint16, database *gorm.DB) {
	db = database
	r := setupRouter()
	r.Run(fmt.Sprintf(":%d", port))
}
