package ldap

import (
	"crypto/tls"
	"fmt"
	"time"

	"net/http"
	"net/url"

	"github.com/dsseng/wiso/pkg/common"
	"github.com/dsseng/wiso/pkg/users"
	"github.com/gin-gonic/gin"
	"github.com/go-ldap/ldap/v3"
	"gorm.io/gorm"
)

type LDAPProvider struct {
	URL        string `yaml:"url"`
	BindDN     string `yaml:"bind_dn"`
	BindPass   string `yaml:"bind_pass"`
	BaseDN     string `yaml:"base_dn"`
	UserFilter string `yaml:"user_filter"`
	NameAttr   string `yaml:"name_attr"`

	BaseURL *url.URL
	DB      *gorm.DB

	l *ldap.Conn
}

func (p LDAPProvider) processLogin(uid string, password string, mac string, linkOrig string) string {
	err := p.l.Bind(p.BindDN, p.BindPass)
	if err != nil {
		return common.ErrorRedirect(p.BaseURL, err, mac, linkOrig)
	}

	req := ldap.NewSearchRequest(
		p.BaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf(p.UserFilter, uid),
		[]string{p.NameAttr},
		nil,
	)

	res, err := p.l.Search(req)
	if err != nil {
		return common.ErrorRedirect(p.BaseURL, err, mac, linkOrig)
	}

	if len(res.Entries) != 1 {
		return common.ErrorRedirect(
			p.BaseURL,
			fmt.Errorf("User does not exist or too many entries returned"),
			mac,
			linkOrig,
		)
	}

	fullName := res.Entries[0].GetAttributeValue(p.NameAttr)

	userdn := res.Entries[0].DN
	err = p.l.Bind(userdn, password)
	if err != nil {
		return common.ErrorRedirect(p.BaseURL, err, mac, linkOrig)
	}

	username := uid + "@ldap"
	user, err := users.FindSingle(p.DB, username)
	if err != nil {
		return common.ErrorRedirect(p.BaseURL, err, mac, linkOrig)
	}

	if len(user) == 0 {
		user = []users.User{
			{
				Username: username,
				FullName: fullName,
			},
		}

		res := p.DB.Create(user)
		if res.Error != nil {
			return common.ErrorRedirect(p.BaseURL, res.Error, mac, linkOrig)
		}
	} else if user[0].FullName != fullName {
		user[0].FullName = fullName
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

func (p LDAPProvider) login(c *gin.Context) {
	uid, v := c.GetPostForm("uid")
	if !v {
		c.Header("Location", common.ErrorRedirect(
			p.BaseURL,
			fmt.Errorf("uid not provided"),
			"",
			"",
		))
		c.Status(http.StatusSeeOther)
		return
	}
	pass, v := c.GetPostForm("pass")
	if !v {
		c.Header("Location", common.ErrorRedirect(
			p.BaseURL,
			fmt.Errorf("pass not provided"),
			"",
			"",
		))
		c.Status(http.StatusSeeOther)
		return
	}
	linkOrig, v := c.GetPostForm("linkOrig")
	if !v {
		linkOrig = ""
	}
	mac, v := c.GetPostForm("mac")
	if !v {
		mac = ""
	}

	c.Header("Location", p.processLogin(uid, pass, mac, linkOrig))
	c.Status(http.StatusSeeOther)
}

func (p LDAPProvider) Setup(r *gin.Engine) error {
	var err error
	p.l, err = ldap.DialURL(p.URL)
	if err != nil {
		return err
	}

	// TODO: support LDAPS and actual security checks
	err = p.l.StartTLS(&tls.Config{InsecureSkipVerify: true})
	if err != nil {
		return err
	}

	err = p.l.Bind(p.BindDN, p.BindPass)
	if err != nil {
		return err
	}

	r.POST("/ldap/login", p.login)

	return nil
}

func (p LDAPProvider) Close() {
	p.l.Close()
}
