package right

// 权限数据类型

// 用户权限
type Right struct {
	ID int  `form:"id" json:"id"`// 用户id
	AuthName string 	` form:"authname" json:"authname"`
	Level string 		`form:"level" json:"level"`
	Pid *int	`form:"pid" json:"pid"`
	Path *string 	`form:"path" json:"path"`
}

// 权限树形类型
type RightRree struct {
	ID int 	`json:"id"`
	AuthName string 	`json:"authName"`
	Path *string `json:"path"`
	Pid *int 	`json:"pid"`
	Children []*RightRree `json:"children"`
}

type Meta struct {
	Msg string	`json:"msg"`
	Status int	`json:"status"`
}

type Message struct {
	Data []Right `json:"data"`
	Meta	`json:"meta"`
}

// 生成权限类型 树结构
func NewRightRree(right Right) *RightRree {
	return &RightRree{
		ID:       right.ID,
		AuthName: right.AuthName,
		Path:     right.Path,
		Pid:      right.Pid,
		Children: make([]*RightRree,0,10),
	}
}

// 获取权限 树形结构切片
func GetRreeSlice(rights []Right) []*RightRree {
	rightSlice := make([]*RightRree,0,10)
	for _,v := range rights {
		// 1级权限
		if v.Level == "1" {
			firstRight := NewRightRree(v)
			firstRightId := v.ID

			for _,v2 := range rights {
				// 二级权限pid = 一级权限id
				if *(v2.Pid) == firstRightId {
					secondRight := NewRightRree(v2)

					secondRightId := v2.ID
					for _,v3 := range rights {
						if *(v3.Pid) == secondRightId {
							thirdRight := NewRightRree(v3)

							secondRight.Children = append(secondRight.Children,thirdRight)
						}
					}

					firstRight.Children = append(firstRight.Children,secondRight)
				}
			}

			rightSlice = append(rightSlice,firstRight)

		}
	}
	return rightSlice
}

func NewMessage() *Message{
	return &Message{
		Data: make([]Right,0,20),
		Meta:  Meta{
			Msg: "获取权限列表成功",
			Status: 200,
		},
	}
}

// 添加新用户
func AddNewAuthRights(msg *Message,rights []Right){
	for _,v := range rights{ // v是指针变量

		msg.Data = append(msg.Data,v)
	}
}

// 返回一个map
func GetMap(rights []Right) map[int]*Right {
	rightMap := make(map[int]*Right,100)
	for _,v := range rights {
		ret := v
		rightMap[v.ID] = &ret
	}
	return rightMap
}