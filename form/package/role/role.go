package role

// 角色id
type RoleId struct {
	Id int 	`json:"rId"`
}

// 分配权限id
type AllotRight struct {
	RoleId string 	` json:"rIds"`
}

// scan 扫描id
type Result struct {
	RightId int
}

// 获取角色名称
type GetRoleName struct {
	RoleName string
}

// 权限类型
type Right struct {
	Id int 	`json:"id"`
	AuthName string 	`json:"authName"`
	Path string `json:"path"`
	Children []*Right 	`json:"children"`
}

// 角色类型
type Role struct {
	Id int 	`json:"id" form:"id"`
	RoleName string 	`json:"roleName" form:"roleName"`
	RoleDesc string	`json:"roleDesc" form:"roleDesc"`
}



type RoleMgr struct {
	Id int 	`json:"id"`
	RoleName string 	`json:"roleName"`
	RoleDesc string	`json:"roleDesc"`
	Children []*Right 	`json:"children"`
}

type Message struct {
	Roles []*RoleMgr 	`json:"roles"`
}



func NewRight(id int, authName string, path string) *Right {
	return &Right{
		Id:       id,
		AuthName: authName,
		Path:     path,
		Children: make([]*Right,0,5),
	}
}

// new 三级权限
func NewThirdRight(id int, authName string, path string) *Right {
	return &Right{
		Id:       id,
		AuthName: authName,
		Path:     path,
	}
}

// 创建一个新角色管理者
func NewRoleMgr(id int,roleName string,roleDesc string) *RoleMgr {
	return &RoleMgr{
		Id:       id,
		RoleName: roleName,
		RoleDesc: roleDesc,
		Children: make([]*Right,0,5),
	}
}

func NewMessage() *Message {
	return &Message{
		Roles:make([]*RoleMgr,0,20),
	}
}

// 角色管理者添加 children
func AddChild(parentRole *RoleMgr,childRight *Right) {
	parentRole.Children = append(parentRole.Children,childRight)
}

// 给权限类型添加子类
func AddChildRight(parentRight *Right,childRight *Right) {
	parentRight.Children = append(parentRight.Children,childRight)
}

func AddRole(msg *Message,role *RoleMgr) {
	msg.Roles = append(msg.Roles,role)
}

func AddRoles(msg *Message,roles []Role,roleidOfRight map[int][]*Right) {
	for _, role := range roles {
		slice,ok := roleidOfRight[role.Id]

		ret := NewRoleMgr(role.Id,role.RoleName,role.RoleDesc)

		if ok {
			for _,v := range slice {
				AddChild(ret,v)
			}
		}
		msg.Roles = append(msg.Roles,ret)
	}
}

// 返回一个map
func GetMap(roles []Role) map[int]*Role {
	roleMap := make(map[int]*Role,100)
	for _,v := range roles {
		ret := v
		roleMap[v.Id] = &ret
	}
	return roleMap
}
