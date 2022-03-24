package dao

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/popstk/inke/internal/config"
	"strings"
)

type Dao struct {
	Db map[string]*gorm.DB
}

var (
	IgnoreKeys = map[string]struct{}{
		"max_idle":         {},
		"max_active":       {},
		"max_lifetime_sec": {},
	}
)

func IgnoreSQL(u string) string {
	seg := strings.Split(u, "?")
	if len(seg) <= 1 {
		return u
	}
	parts := strings.Split(seg[1], "&")
	var out []string
	for _, part := range parts {
		fields := strings.Split(part, "=")
		if len(fields) != 2 {
			out = append(out, part)
		}
		if _, ok := IgnoreKeys[fields[0]]; !ok {
			out = append(out, part)
		}
	}

	return seg[0] + "?" + strings.Join(out, "&")
}

func NewDao(c *config.Config) (*Dao, error) {
	dao := &Dao{
		Db: make(map[string]*gorm.DB),
	}

	for _, db := range c.Database {
		u := IgnoreSQL(db.Master)
		c, err := gorm.Open("mysql", u)
		if err != nil {
			return nil, fmt.Errorf("can not connect db name %s: %v", db.Name, err)
		}
		dao.Db[db.Name] = c
	}

	return dao, nil
}

func (d *Dao) GetDB(name string) *gorm.DB {
	return d.Db[name]
}
