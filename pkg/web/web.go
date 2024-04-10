package web

import (
	"embed"
	"html/template"
	"net/http"
	"net/url"

	"github.com/dsseng/wiso/pkg/oidc"
	"github.com/dsseng/wiso/pkg/users"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type App struct {
	BaseURL      *url.URL
	DB           *gorm.DB
	OIDC         *oidc.OIDCProvider
	PasswordAuth bool
	LogoLogin    string
	LogoWelcome  string
	LogoError    string
	SupportURL   string
}

var (
	//go:embed templates
	embedFS embed.FS
)

func (a App) setupRouter() (*gin.Engine, error) {
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
		c.FileFromFS("templates/static/"+c.Param("path"), staticFS)
	})

	// Args: mac of the device which is being authorized
	r.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", gin.H{
			"title":        "Network login",
			"oidcID":       a.OIDC.ID,
			"oidcName":     a.OIDC.Name,
			"passwordAuth": a.PasswordAuth,
			"image":        a.LogoLogin,
			"mac":          c.Query("mac"),
			"redirectURL":  c.Query("link-orig"),
		})
	})

	// Args: picture URL, link-orig, full_name and username to be displayed
	r.GET("/welcome", func(c *gin.Context) {
		c.HTML(http.StatusOK, "welcome.html", gin.H{
			"title":     "Success",
			"picture":   c.Query("picture"),
			"full_name": c.Query("full_name"),
			"username":  c.Query("username"),
			"logo":      a.LogoWelcome,
		})
	})

	// Args: error to be displayed
	r.GET("/error", func(c *gin.Context) {
		c.HTML(http.StatusOK, "error.html", gin.H{
			"title":   "Error",
			"error":   c.Query("error"),
			"logo":    a.LogoError,
			"support": a.SupportURL,
		})
	})

	a.OIDC.BaseURL = a.BaseURL
	a.OIDC.DB = a.DB

	err := a.OIDC.Setup(r)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func (a App) Start() error {
	a.DB.AutoMigrate(&users.User{})
	a.DB.AutoMigrate(&users.DeviceSession{})

	r, err := a.setupRouter()
	if err != nil {
		return err
	}

	r.Run(":" + a.BaseURL.Port())

	return nil
}
