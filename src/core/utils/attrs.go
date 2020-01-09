package utils

/* 实体对象属性集合 */
type Attrs struct {
	values []interface{}
}

/* 创建一个属性集 */
func NewAttrs(limit int32) *Attrs {
	attrs := &Attrs{}
	attrs.values = make([]interface{}, limit, limit)
	return attrs
}

// 根据下标获取对应的值
func (self *Attrs) Get(idx int32) interface{} {
	return self.values[idx]
}

// 根据下标获取对应的值
func (self *Attrs) Set(idx int32, val interface{}) {
	self.values[idx] = val
}
