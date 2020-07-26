package user

// User --> 数据库表
type User struct {
	ID uint		`json:"id"`
	UserName string	`form:"username" json:"username"`
	Mobile string	`form:"mobile" json:"mobile"`
	Email string	`form:"email" json:"email"`
	RoleName string  `form:"rolename" json:"roleName"` // 超级管理员
	UserState *bool  	`json:"userState"`// 当前用户状态
}


type Meta struct {
	Msg string	`json:"msg"`
	Status int	`json:"status"`
}

type Message struct {
	Pagenum int 	`json:"pagenum"`
	Total int 	`json:"total"`
	Users []User	`json:"data"`
	Meta	`json:"meta"`
}



func NewMessage() *Message{
	return &Message{
		Users: make([]User,0,20),
		Meta:  Meta{
			Msg: "获取用户列表成功",
			Status: 200,
		},
	}
}

// 添加新用户
func AddNewUsers(msg *Message,users []User){
	for _,v := range users{ // v是指针变量
		//ret := v
		msg.Users = append(msg.Users,v)
	}
}

func GetBoolVal(val bool)*bool{
	ret := val
	return &ret
}