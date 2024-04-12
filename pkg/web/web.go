package web

import (
	"embed"
	"html/template"
	"net/http"
	"net/url"
	"runtime/debug"

	"github.com/dsseng/wiso/pkg/ldap"
	"github.com/dsseng/wiso/pkg/oidc"
	"github.com/dsseng/wiso/pkg/users"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type App struct {
	Base        string // Used by CLI
	BaseURL     *url.URL
	Database    string // Used by CLI
	DB          *gorm.DB
	OIDC        *oidc.OIDCProvider
	LDAP        *ldap.LDAPProvider
	LogoLogin   string `yaml:"logo_login"`
	LogoWelcome string `yaml:"logo_welcome"`
	LogoError   string `yaml:"logo_error"`
	SupportURL  string `yaml:"support_url"`
}

var (
	//go:embed templates
	embedFS embed.FS
)

var gitRevision = func() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				return setting.Value[:8]
			}
		}
	}
	return ""
}()

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
		oidcID := ""
		oidcName := ""
		if a.OIDC != nil {
			oidcID = a.OIDC.ID
			oidcName = a.OIDC.Name
		}

		c.HTML(http.StatusOK, "login.html", gin.H{
			"title":        "Network login",
			"oidcID":       oidcID,
			"oidcName":     oidcName,
			"passwordAuth": a.LDAP != nil,
			"image":        a.LogoLogin,
			"mac":          c.Query("mac"),
			"redirectURL":  c.Query("link-orig"),
			"commit":       gitRevision,
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

	// Args: error to be displayed, mac and link-orig for retru
	r.GET("/error", func(c *gin.Context) {
		c.HTML(http.StatusOK, "error.html", gin.H{
			"title":       "Error",
			"error":       c.Query("error"),
			"logo":        a.LogoError,
			"mac":         c.Query("mac"),
			"redirectURL": c.Query("link-orig"),
			"support":     a.SupportURL,
		})
	})

	if a.OIDC != nil {
		a.OIDC.BaseURL = a.BaseURL
		a.OIDC.DB = a.DB

		err := a.OIDC.Setup(r)
		if err != nil {
			return nil, err
		}
	}

	if a.LDAP != nil {
		a.LDAP.BaseURL = a.BaseURL
		a.LDAP.DB = a.DB

		err := a.LDAP.Setup(r)
		if err != nil {
			return nil, err
		}
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

	// TODO: ensure this is handled well
	defer (func() {
		if a.LDAP != nil {
			a.LDAP.Close()
		}
	})()
	r.Run(":" + a.BaseURL.Port())
	return nil
}
