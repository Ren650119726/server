package model

import "database/sql"

/*
复合主键(将多个字段设置为主键以启用复合主键)
*/
type example struct {
	ID           string `gorm:"primary_key"`
	LanguageCode string `gorm:"primary_key"`
}

/*
索引
*/
type example2 struct {
	ID     int
	UserID int    `gorm:"index"`                          // 外键 (属于), tag `index`是为该列创建索引
	Email  string `gorm:"type:varchar(100);unique_index"` // `type`设置sql类型, `unique_index` 为该列设置唯一索引
	Name   string `gorm:"index:idx_name_code"`            // 创建索引并命名，如果找到其他相同名称的索引则创建组合索引
	Code   string `gorm:"index:idx_name_code"`            // `unique_index` also works
}

type example3 struct {
	AnimalId int64          `gorm:"primary_key"`             // 设置AnimalId为主键
	Age      int64          `gorm:"column:age_of_the_beast"` // 设置列名为`age_of_the_beast`
	Address1 string         `gorm:"not null;unique"`         // 设置字段为非空并唯一
	Address2 string         `gorm:"type:varchar(100);unique"`
	Post     sql.NullString `gorm:"not null"`
	IgnoreMe int            `gorm:"-"`              // 忽略这个字段
	Num      int            `gorm:"AUTO_INCREMENT"` // 自增
}
