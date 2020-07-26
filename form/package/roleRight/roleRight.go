package roleRight


// 角色和权限关联表
type RoleRight struct {
	ID int
	RoleId int  // 角色id
	RightId int  // 权限id
	Level int // 权限等级
	Pid *int // 当前权限的父级id
}