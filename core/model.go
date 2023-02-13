package core

type Model struct {
	CreateAt DateTime     `json:"createAt" form:"createAt" time_format:"2006-01-02 15:04:05" time_utc:"1" gorm:"type:datetime;comment:创建时间"`
	DeleteAt DateTime     `json:"deleteAt" form:"deleteAt" time_format:"2006-01-02 15:04:05" time_utc:"1" gorm:"type:datetime;comment:删除时间"`
	Deleted  DeleteStatus `json:"deleted" form:"deleted" gorm:"type:int(11);default:0;comment:是否删除"`
}

var modelArray []interface{} = make([]interface{}, 0)

func RegisterModel(m interface{}) {
	modelArray = append(modelArray, m)
}
func AllRegistedModel() []interface{} {
	return modelArray
}
