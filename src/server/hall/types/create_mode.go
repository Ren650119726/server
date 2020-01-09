package types

type ECreateMode uint8

// 登录类型
const (
	MODE_FIXED_MATCH_TYPE   ECreateMode = 1
	MODE_PARS_STRING_PARAM  ECreateMode = 2
	MODE_ACCORD_PARAM_INDEX ECreateMode = 3
)

var strCreateModeType = map[ECreateMode]string{
	MODE_FIXED_MATCH_TYPE:   "固定匹配档次模式",
	MODE_PARS_STRING_PARAM:  "解析字符串参数中的值为匹配档次",
	MODE_ACCORD_PARAM_INDEX: "参数下标为匹配档次",
}

func (e ECreateMode) String() string {
	return strCreateModeType[e]
}

func (e ECreateMode) Value() uint8 {
	return uint8(e)
}
