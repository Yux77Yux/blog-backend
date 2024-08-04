package model

type UsernameAndPassword struct {
	Username string
	Password string
}

type UserIncidental struct {
	//外键，与User.Id联系不要返回这个数据
	Id int32 `json:"id"`
	//主键，主要用于被搜索，生成后不可更改
	Uid string `json:"uid"`
	//昵称，也可用于被搜索，允许可以更改
	Name string `json:"name"`
	//头像图片地址
	Profile string `json:"profile"`
	//个性签名
	Bio string `json:"bio"`
	//登录状态
	Status bool `json:"status"`
	//主页受欢迎程度，用于优先搜索
	Popularity float32 `json:"popularity"`
}

type UserModifyProfile struct {
	Id      int32  `json:"id"`
	Profile string `json:"profile"`
}

type UserModifyName struct {
	Id   int32  `json:"id"`
	Name string `json:"name"`
}

type UserModifyBio struct {
	Id  int32  `json:"id"`
	Bio string `json:"bio"`
}
