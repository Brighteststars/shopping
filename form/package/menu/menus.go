package menu

// 菜单结构体
type Menu struct {
	Id int	`json:"id"`
	AuthName string	`json:"authname"`
	Path string	`json:"path"`
	Children []*Menu	`json:"children"`
}

type Meta struct {
	Msg string	`json:"msg"`
	Status int	`json:"status"`
}

type Message struct {
	Menus []*Menu	`json:"data"`
	Meta	`json:"meta"`
}

// 创建一级菜单
func NewMenu(id int,authName string,path string) *Menu {
	return &Menu{
		Id:id,
		AuthName:authName,
		Path:path,
		Children:make([]*Menu,0,10),
	}
}

// 添加二级菜单
func AddChildMenu(menu *Menu,childMenu *Menu){
	menu.Children = append(menu.Children,childMenu)
}

// 创建新Message对象
func NewMessage() *Message{
	return &Message{
		Menus: make([]*Menu,0,10),
		Meta:  Meta{
			Msg: "获取菜单列表成功",
			Status: 200,
		},
	}
}

func AddNewMenu(msg *Message,menu *Menu) {
	msg.Menus = append(msg.Menus,menu)
}