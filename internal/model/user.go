package model

type UsernameAndPassword struct {
	Username string
	Password string
}

type UserActive struct {
	//外键，与User.Id联系不要返回这个数据
	Id int32 `json:"id"`
	//昵称，也可用于被搜索，允许可以更改
	Name string `json:"name"`
	//头像图片地址
	Profile string `json:"profile"`
	//个性签名
	Bio string `json:"bio"`
	//登录状态
	Status bool `json:"status"`
}

type UserIncidental struct {
	UserActive
	//主键，主要用于被搜索，生成后不可更改
	Uid string `json:"uid"`
	//主页受欢迎程度，用于优先搜索
	Popularity int32 `json:"popularity"`
}
