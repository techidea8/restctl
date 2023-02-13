package core

import (
	"gorm.io/gorm"
)

type PageArg struct {
	Pagefrom int          `json:"pagefrom" form:"pageform"`
	Pagesize int          `json:"pagesize" form:"pagesize"`
	DescBy   string       `json:"descby" form:"descby"`
	AscBy    string       `json:"ascby" form:"ascby"`
	Datefrom DateTime     `json:"datefrom" form:"datefrom" time_format:"2006-01-02 15:04:05" time_utc:"1"`
	Dateto   DateTime     `json:"dateto"  form:"dateto"    time_format:"2006-01-02 15:04:05" time_utc:"1"`
	Deleted  DeleteStatus `json:"deleted"  form:"deleted"`
}

func (p *PageArg) GetPageSize() int {
	if p.Pagesize == 0 {
		return 100
	} else {
		return p.Pagesize
	}
}
func (p *PageArg) GetPageFrom() int {
	if p.Pagefrom < 1 {
		return 0
	} else {
		return p.Pagefrom - 1
	}
}

func (p *PageArg) Sort() string {
	if len(p.DescBy) > 0 {
		return p.DescBy + " DESC"
	} else if len(p.AscBy) > 0 {
		return p.AscBy + " ASC"
	} else {
		return " id desc "
	}
}

// 分页
func (p *PageArg) Paginate() func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		offset := p.GetPageFrom() * p.GetPageSize()
		return db.Offset(offset).Limit(p.GetPageSize())
	}
}
