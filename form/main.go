package main

import (
	"fmt"
	"form/middleware"
	menus "form/package/menu"
	"form/package/right"           // 权限类型
	"form/package/role"            // 角色类型
	"form/package/roleRight"       // 角色和权限关联表
	"form/package/user"            // 用户类型
	viperData "form/package/viper" // viper数据结构
	"net/http"
	"strconv"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/spf13/viper"
)

// Binding from JSON
type Login struct {
	User     string `form:"user" json:"username" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}

type QueryInfo struct {
	Query    string `form:"query" `
	Pagenum  int    `form:"pagenum" `
	Pagesize int    `form:"pagesize" `
}

var db *gorm.DB
var err error

func main() {
	initViper()

	var C viperData.Config

	if err := viper.Unmarshal(&C); err != nil {
		fmt.Printf("unable to decode into struct, %v\n", err)
		return
	}

	mysqlConfig := fmt.Sprintf("%s:%s@(%s:%v)/%s?charset=utf8&parseTime=True&loc=Local",
		C.DbName, C.DbName, C.Host, C.MysqlConfig.Port, C.DbName)

	//db, err = gorm.Open("mysql",
	//	"dmfs9000:dmfs9000@(127.0.0.1:3306)/dmfs9000?charset=utf8&parseTime=True&loc=Local")
	db, err = gorm.Open("mysql",
		mysqlConfig)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// 禁用默认表名的复数形式，如果置为 true，则 `User` 的默认表名是 `user`
	db.SingularTable(true)

	// 创建表 自动迁移 (结构体与数据库表对应)
	db.AutoMigrate(&user.User{}, &right.Right{}, &role.Role{}, &roleRight.RoleRight{})

	r := gin.Default()

	r.Use(middleware.Cors()) //使用全局中间件

	//r.Delims("{[{", "}]}")
	//
	//r.Static("/static", "static")
	//
	//r.LoadHTMLGlob("templates/*")
	//r.LoadHTMLFiles("./Login.html", "./vue.html") //加载模板文件

	// 绑定JSON的示例 ({"user": "q1mi", "password": "123456"})
	r.POST("/login", loginHandler)

	r.GET("/home", middleware.JWTAuthMiddleware(), homeHandler)

	// 获取菜单列表
	r.GET("/menus", middleware.JWTAuthMiddleware(), menusHandler)

	// 获取用户列表
	r.GET("/users", middleware.JWTAuthMiddleware(), usersHandler)

	// 分配用户角色
	r.PUT("/user/:id/role", middleware.JWTAuthMiddleware(), allotUserRoleHandler)

	// 用户表信息
	userGroup := r.Group("/users")
	{
		// 根据id 获取用户信息
		userGroup.GET("/:uId", middleware.JWTAuthMiddleware(), getUserInfoHandler)

		// 更新用户信息
		userGroup.POST("/:uId", middleware.JWTAuthMiddleware(), updateUserInfoHandler)

		// 删除用户信息
		userGroup.DELETE("/:uId", middleware.JWTAuthMiddleware(), removeUserInfoHandler)
	}

	// 接收用户信息 创建新用户
	r.POST("/users", middleware.JWTAuthMiddleware(), receiveUsersHandler)

	// 修改用户状态 改变userState
	r.PUT("/users/:uId/state/:type", middleware.JWTAuthMiddleware(), modifyUserState)

	// 权限接口
	r.GET("/rights/:type", middleware.JWTAuthMiddleware(), rightsHandler)

	// 角色接口
	r.GET("/roles", middleware.JWTAuthMiddleware(), rolesHandler)

	// 角色 接收数据
	r.POST("/roles", middleware.JWTAuthMiddleware(), receiveRolesHandler)

	roleGroup := r.Group("/roles")
	{
		// 根据id 获取角色信息
		roleGroup.GET("/:id", middleware.JWTAuthMiddleware(), getRoleInfoHandler)

		// 更新角色信息
		roleGroup.POST("/:id", middleware.JWTAuthMiddleware(), updateRoleInfoHandler)
		//
		// 删除用户信息
		roleGroup.DELETE("/:id", middleware.JWTAuthMiddleware(), removeRoleInfoHandler)
	}

	// 删除角色权限
	r.DELETE("/role/:roleId/right/:rightId", middleware.JWTAuthMiddleware(), removeRoleRightHandler)

	// 分配权限
	r.POST("/role/:roleId/right", middleware.JWTAuthMiddleware(), receiveRolesId)

	r.Run(":9090")
}

// 初始化viper，解析config.yaml配置文件
func initViper() {

	// 设置默认值
	viper.SetDefault("fileDir", "./")

	// 读取配置文件
	//viper.SetConfigFile("config.yaml") // 指定配置文件
	viper.SetConfigName("config")         // 配置文件名称(无扩展名)
	viper.SetConfigType("yaml")           // 如果配置文件的名称中没有扩展名，则需要配置此项
	viper.AddConfigPath("/etc/appname/")  // 查找配置文件所在的路径
	viper.AddConfigPath("$HOME/.appname") // 多次调用以添加多个搜索路径
	viper.AddConfigPath(".")              // 还可以在工作目录中查找配置
	err := viper.ReadInConfig()           // 查找并读取配置文件
	if err != nil {                       // 处理读取配置文件的错误
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	// 监控配置文件变化
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		// 配置文件发生变更之后会调用的回调函数
		fmt.Println("Config file changed:", e.Name)
	})
}

func homeHandler(c *gin.Context) {
	username := c.MustGet("username").(string)
	fmt.Printf("homeHandler-username:%s\n", username)
	c.JSON(http.StatusOK, gin.H{
		"code":    2000,
		"msg":     "success",
		"package": gin.H{"username": username},
	})
}

// 分配用户角色
func allotUserRoleHandler(c *gin.Context) {
	username := c.MustGet("username").(string)
	fmt.Printf("homeHandler-username:%s\n", username)

	userId := c.Param("id")
	fmt.Printf("userId:%s\n", userId)

	var roleId role.RoleId
	if err := c.ShouldBind(&roleId); err == nil {

		var result role.GetRoleName
		err := db.Table("role").Select("role_name").
			Where("id = ?", roleId.Id).Scan(&result).Error

		// 访问数据库出错(包含没有访问到内容)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"msg":    err.Error(),
				"status": 500,
				"data":   "",
			})
		} else {
			err = db.Model(&user.User{}).Where("id = ?", userId).
				Update("role_name", result.RoleName).Error
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"msg":    err.Error(),
					"status": 500,
					"data":   "",
				})
			} else {
				//输出json结果给调用方
				c.JSON(http.StatusOK, gin.H{
					"msg":      "分配角色",
					"status":   200,
					"roleId":   roleId.Id,
					"userId":   userId,
					"roleName": result.RoleName,
				})
			}
		}

	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

}

// 分配权限
func receiveRolesId(c *gin.Context) {
	username := c.MustGet("username").(string)
	fmt.Printf("homeHandler-username:%s\n", username)

	roleId := c.Param("roleId")
	fmt.Printf("roleId:%v\n", roleId)

	var allotRightInfo role.AllotRight

	if err := c.ShouldBind(&allotRightInfo); err == nil {
		fmt.Printf("allotRightInfo:%#v\n", allotRightInfo)
		rId := allotRightInfo.RoleId // id 字符串 逗号分隔

		// 解析字符串id 存入sliceInt
		slice := strings.Split(rId, ",")
		sliceInt := make([]int, 0, 100)
		for _, str := range slice {
			v, _ := strconv.Atoi(str)
			sliceInt = append(sliceInt, v)
		}

		// 访问数据库获取id
		var results []role.Result
		db.Debug().Table("role_right").Select("right_id").Where("role_id = ?", roleId).Scan(&results)
		fmt.Printf("results:%#v\n", results)

		var exist bool
		// 不存在的ids
		inexist := make([]int, 0, 30)
		// 交集存在的ids
		existSlice := make([]int, 0, 50)
		for _, v1 := range results {
			exist = false
			for _, v2 := range sliceInt {
				if v1.RightId == v2 {
					exist = true
					break
				}
			}
			if !exist {
				inexist = append(inexist, v1.RightId)
			} else {
				existSlice = append(existSlice, v1.RightId)
			}
		}

		newAddIdSlice := make([]int, 0, 30)
		// web前端传过来的roleID总数 大于 role_right表中的权限

		if len(sliceInt) > len(existSlice) {
			for _, v1 := range sliceInt {
				exist = false
				for _, v2 := range existSlice {
					if v2 == v1 {
						exist = true
						break
					}
				}
				if !exist {
					newAddIdSlice = append(newAddIdSlice, v1)
				}
			}
		}

		if len(newAddIdSlice) != 0 {

			var rights []right.Right
			// 查询所有的记录
			db.Debug().Find(&rights)

			m := make(map[int]*right.Right, 50)
			sql := "INSERT INTO `role_right` (`role_id`,`right_id`,`level`,`pid`) VALUES "
			for _, v := range rights {
				ret := v
				m[ret.ID] = &ret
			}

			for index, v := range newAddIdSlice {

				if len(newAddIdSlice)-1 == index {
					//最后一条数据 以分号结尾
					sql += fmt.Sprintf("(%s,%d,%s,%d);", roleId, m[v].ID, m[v].Level, *(m[v].Pid))
				} else {
					sql += fmt.Sprintf("(%s,%d,%s,%d),", roleId, m[v].ID, m[v].Level, *(m[v].Pid))
				}
			}
			// 批量新增操作
			db.Exec(sql)
		}

		// 数据库删除操作
		db.Debug().Delete(roleRight.RoleRight{}, "role_id = ? AND right_id IN (?)", roleId, inexist)

		//输出json结果给调用方
		c.JSON(http.StatusOK, gin.H{
			"msg":     "权限id",
			"status":  200,
			"data":    allotRightInfo.RoleId,
			"roleId":  roleId,
			"results": results,
			"inexist": inexist, // 删除的ids
			"newAdd":  newAddIdSlice,
		})

	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
}

// 移除角色权限
func removeRoleRightHandler(c *gin.Context) {
	username := c.MustGet("username").(string)
	fmt.Printf("homeHandler-username:%s\n", username)

	roleId := c.Param("roleId")
	rightId := c.Param("rightId")

	// 将字符串转化为数值 int 类型
	roleIdInt, err := strconv.Atoi(roleId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	var roleRightRet roleRight.RoleRight
	// 查询所有的记录
	db.Debug().Where("role_id = ? AND right_id = ?", roleId, rightId).First(&roleRightRet)

	if roleRightRet.Level == 3 {
		// 3级权限 直接删除
		db.Debug().Where("role_id = ? AND right_id = ?", roleId, rightId).Delete(roleRight.RoleRight{})

	} else if roleRightRet.Level == 2 {
		// 2级权限
		// 先删除3级权限
		db.Debug().Where("role_id = ? AND pid = ?", roleId, rightId).Delete(roleRight.RoleRight{})
		// 再删除2级权限
		db.Debug().Where("role_id = ? AND right_id = ?", roleId, rightId).Delete(roleRight.RoleRight{})
	} else {
		// 1级权限
		// 先删除3级权限
		//db.Debug().Where("role_id = ? AND pid IN ?",roleId,
		//	db.Table("role_right").
		//	Select("right_id").Where("role_id = ? AND pid = ?", roleId,rightId ).SubQuery()).
		//	Delete(roleRight.RoleRight{})

		db.Debug().Exec("DELETE FROM role_right WHERE (role_id = ? AND pid IN ( SELECT tmp.right_id FROM "+
			"( SELECT right_id FROM role_right WHERE role_id = ? AND pid = ? ) tmp ) )", roleId, roleId, rightId)

		// 删除2级权限
		db.Debug().Where("role_id = ? AND pid = ?", roleId, rightId).Delete(roleRight.RoleRight{})

		// 删除1级权限
		db.Debug().Where("role_id = ? AND right_id = ?", roleId, rightId).Delete(roleRight.RoleRight{})

	}

	mapRights := getMapRights()
	rightSlice, _ := mapRights[roleIdInt]

	c.JSON(http.StatusOK, gin.H{
		"status": 200,
		"msg":    "删除角色权限数据成功",
		"data":   rightSlice,
	})

}

func removeRoleInfoHandler(c *gin.Context) {
	username := c.MustGet("username").(string)
	fmt.Printf("homeHandler-username:%s\n", username)

	roleId := c.Param("id")

	// 将字符串转化为数值
	id, err := strconv.ParseInt(roleId, 0, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	db.Debug().Delete(role.Role{}, "id = ?", id)

	c.JSON(http.StatusOK, gin.H{
		"status": 200,
		"msg":    "删除角色数据成功",
	})

}

// 更新角色信息
func updateRoleInfoHandler(c *gin.Context) {
	username := c.MustGet("username").(string)
	fmt.Printf("homeHandler-username:%s\n", username)

	roleId := c.Param("id")

	// 将字符串转化为数值
	id, err := strconv.ParseInt(roleId, 0, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	var roleInfo role.Role

	if err := c.ShouldBind(&roleInfo); err == nil {
		fmt.Printf("roleInfo:%#v\n", roleInfo)

		// 使用 struct 更新多个属性，只会更新其中有变化且为非零值的字段
		db.Debug().Model(&(role.Role{})).Where("id = ?", id).Updates(roleInfo)

		//输出json结果给调用方
		c.JSON(http.StatusOK, gin.H{
			"msg":    "更新role数据成功",
			"status": 200,
			"data":   roleInfo,
		})

	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

}

func getRoleInfoHandler(c *gin.Context) {
	username := c.MustGet("username").(string)
	fmt.Printf("homeHandler-username:%s\n", username)

	roleId := c.Param("id")

	// 将字符串转化为数值
	id, err := strconv.ParseInt(roleId, 0, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	var role role.Role
	db.Where("id = ?", id).First(&role)

	//输出json结果给调用方
	c.JSON(http.StatusOK, gin.H{
		"msg":  "获取user数据成功",
		"code": 200,
		"data": role,
	})
}

// 角色接收信息
func receiveRolesHandler(c *gin.Context) {
	username := c.MustGet("username").(string)
	fmt.Printf("homeHandler-username:%s\n", username)

	var roleInfo role.Role

	if err := c.ShouldBind(&roleInfo); err == nil {
		fmt.Printf("userInfo:%#v\n", roleInfo)

		//userInfo.UserState = user.GetBoolVal(true)

		db.Create(&roleInfo)
		c.JSON(http.StatusOK, gin.H{
			"status":   200,
			"roleInfo": roleInfo,
		})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

}

// 生成三级权限
func getMapRights() map[int][]*role.Right {
	// 权限表
	var rights []right.Right
	db.Find(&rights)
	rightsMap := right.GetMap(rights)
	// 角色和权限关联表
	var roleRights []roleRight.RoleRight
	db.Find(&roleRights)

	// key 角色id value 权限
	rightsMapStore := make(map[int][]*role.Right, 20)

	for _, roleRightRef := range roleRights {

		// 生成一级权限
		if roleRightRef.Level == 1 {
			// 一级权限id
			firstRightId := roleRightRef.RightId
			rightRet, _ := rightsMap[firstRightId]
			// new 一级权限
			firstRight := role.NewRight(rightRet.ID, rightRet.AuthName, *(rightRet.Path))
			for _, v := range roleRights {
				// 生成二级权限
				secondRight := v
				secondRightId := secondRight.RightId
				if *(secondRight.Pid) == firstRightId {
					rightRet, _ := rightsMap[secondRightId]
					// new 二级权限
					secondRight := role.NewRight(rightRet.ID, rightRet.AuthName, *(rightRet.Path))

					for _, v := range roleRights {
						// 生成三级权限
						thirdRight := v
						thirdRightId := thirdRight.RightId
						if *(thirdRight.Pid) == secondRightId {
							rightRet, _ := rightsMap[thirdRightId]
							// new 三级级权限
							thirdRight := role.NewThirdRight(rightRet.ID, rightRet.AuthName, *(rightRet.Path))
							// 二级权限添加三级权限
							role.AddChildRight(secondRight, thirdRight)
						}
					}
					// 一级权限添加二级权限
					role.AddChildRight(firstRight, secondRight)
				}
			}
			sliceRight, ok := rightsMapStore[roleRightRef.RoleId]
			if !ok {
				sliceRight = make([]*role.Right, 0, 10)
			}
			sliceRight = append(sliceRight, firstRight)
			rightsMapStore[roleRightRef.RoleId] = sliceRight
		}
	}
	return rightsMapStore
}

// 角色处理 获取角色
func rolesHandler(c *gin.Context) {
	username := c.MustGet("username").(string)
	fmt.Printf("homeHandler-username:%s\n", username)

	var roles []role.Role
	// 查询所有的记录 角色表
	db.Find(&roles)

	// 获取三级权限
	rightsMapStore := getMapRights()

	msg := role.NewMessage()
	role.AddRoles(msg, roles, rightsMapStore)

	c.JSON(http.StatusOK, gin.H{
		"status": 200,
		"msg":    "获取角色列表成功!",
		"data":   msg,
	})
}

// 权限处理
func rightsHandler(c *gin.Context) {
	username := c.MustGet("username").(string)
	fmt.Printf("homeHandler-username:%s\n", username)

	ret := c.Param("type")

	if ret == "list" {

		var rights []right.Right
		// 查询所有的记录
		db.Debug().Find(&rights)

		msg := right.NewMessage()
		right.AddNewAuthRights(msg, rights)

		c.JSON(http.StatusOK, msg)
	} else if ret == "tree" {
		var rights []right.Right
		// 查询所有的记录
		db.Debug().Find(&rights)

		rightSlice := right.GetRreeSlice(rights)

		c.JSON(http.StatusOK, gin.H{
			"status": 200,
			"msg":    "获取权限树形结构信息成功!",
			"data":   rightSlice,
		})

	}

}

func removeUserInfoHandler(c *gin.Context) {
	username := c.MustGet("username").(string)
	fmt.Printf("homeHandler-username:%s\n", username)

	userId := c.Param("uId")

	// 将字符串转化为数值
	id, err := strconv.ParseInt(userId, 0, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	db.Debug().Delete(user.User{}, "id = ?", id)

	c.JSON(http.StatusOK, gin.H{
		"status": 200,
		"msg":    "删除用户数据成功",
	})

}

// 更新用户信息
func updateUserInfoHandler(c *gin.Context) {
	username := c.MustGet("username").(string)
	fmt.Printf("homeHandler-username:%s\n", username)

	userId := c.Param("uId")

	// 将字符串转化为数值
	id, err := strconv.ParseInt(userId, 0, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	var userInfo user.User

	if err := c.ShouldBind(&userInfo); err == nil {
		fmt.Printf("userInfo:%#v\n", userInfo)

		// 使用 struct 更新多个属性，只会更新其中有变化且为非零值的字段
		db.Debug().Model(&(user.User{})).Where("id = ?", id).Updates(userInfo)

		//输出json结果给调用方
		c.JSON(http.StatusOK, gin.H{
			"msg":    "更新user数据成功",
			"status": 200,
			"data":   userInfo,
		})

	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

}

// 获取用户信息
func getUserInfoHandler(c *gin.Context) {
	username := c.MustGet("username").(string)
	fmt.Printf("homeHandler-username:%s\n", username)

	userId := c.Param("uId")

	// 将字符串转化为数值
	id, err := strconv.ParseInt(userId, 0, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	var user user.User
	db.Where("id = ?", id).First(&user)

	//输出json结果给调用方
	c.JSON(http.StatusOK, gin.H{
		"msg":  "获取user数据成功",
		"code": 200,
		"data": user,
	})
}

// 接收用户信息
func receiveUsersHandler(c *gin.Context) {
	username := c.MustGet("username").(string)
	fmt.Printf("homeHandler-username:%s\n", username)

	var userInfo user.User

	if err := c.ShouldBind(&userInfo); err == nil {
		fmt.Printf("userInfo:%#v\n", userInfo)

		userInfo.UserState = user.GetBoolVal(true)

		db.Create(&userInfo)
		c.JSON(http.StatusOK, gin.H{
			"status":   200,
			"userInfo": userInfo,
		})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

}

// 修改用户状态
func modifyUserState(c *gin.Context) {
	username := c.MustGet("username").(string)
	fmt.Printf("homeHandler-username:%s\n", username)
	userId := c.Param("uId")
	userState := c.Param("type")

	fmt.Printf("userId:%s userState:%s\n", userId, userState)

	// 将字符串转化为数值
	id, err := strconv.ParseInt(userId, 0, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
	state, err := strconv.ParseBool(userState)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	// 更新单个属性，如果它有变化
	db.Debug().Model(&(user.User{})).Where("id=?", id).Update("user_state", state)
	//输出json结果给调用方
	c.JSON(http.StatusOK, gin.H{
		"msg":  "更新user数据成功",
		"code": 200,
	})
}

// 获取用户数据 分页
func usersHandler(c *gin.Context) {
	username := c.MustGet("username").(string)
	fmt.Printf("homeHandler-username:%s\n", username)

	var queryInfo QueryInfo

	if err := c.ShouldBind(&queryInfo); err == nil {
		fmt.Printf("queryinfo:%#v\n", queryInfo)
		// 分页查询
		var users []user.User
		pagesize := queryInfo.Pagesize // 一页2条数据
		pagenum := queryInfo.Pagenum   // 第三页
		like := "%" + queryInfo.Query + "%"
		db.Debug().Where("user_name LIKE ?", like).Limit(pagesize).Offset(pagesize * (pagenum - 1)).Order("id asc").Find(&users)

		// 总数查询
		var count int
		db.Debug().Where("user_name LIKE ?", like).Model(&(user.User{})).Count(&count)

		msg := user.NewMessage()
		msg.Total = count

		user.AddNewUsers(msg, users)
		msg.Pagenum = pagenum

		c.JSON(http.StatusOK, msg)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

}

// 绑定JSON的示例 ({"user": "q1mi", "password": "123456"})
func loginHandler(c *gin.Context) {
	var login Login

	if err := c.ShouldBind(&login); err == nil {
		fmt.Printf("login info:%#v\n", login)
		//c.JSON(http.StatusOK, gin.H{
		//	"username":     login.User,
		//	"password": login.Password,
		//})
		// 校验用户名和密码是否正确
		if login.User == "luffy" && login.Password == "123" {
			// 生成Token
			tokenString, _ := middleware.GenToken(login.User)
			c.JSON(http.StatusOK, gin.H{
				"code":     2000,
				"msg":      "success",
				"token":    tokenString,
				"username": login.User,
				"password": login.Password,
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"code": 2002,
			"msg":  "鉴权失败",
		})
		return
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
}

// 获取菜单列表
func menusHandler(c *gin.Context) {
	username := c.MustGet("username").(string)
	fmt.Printf("homeHandler-username:%s\n", username)

	msg := menus.NewMessage()

	// 用户
	menu := menus.NewMenu(101, "用户管理", "")
	childMenu := menus.NewMenu(104, "用户列表", "users")
	menus.AddChildMenu(menu, childMenu)
	menus.AddNewMenu(msg, menu)

	// 权限
	menuRight := menus.NewMenu(110, "权限管理", "")
	roleMenu := menus.NewMenu(111, "角色列表", "roles")
	rightsMenu := menus.NewMenu(112, "权限列表", "rights")
	menus.AddChildMenu(menuRight, roleMenu)
	menus.AddChildMenu(menuRight, rightsMenu)
	menus.AddNewMenu(msg, menuRight)

	// 商品
	menuGoods := menus.NewMenu(105, "商品管理", "")
	goodsListMenu := menus.NewMenu(106, "商品列表", "goods")
	classifyMenu := menus.NewMenu(107, "分类列表", "classify")
	goodsClassifyMenu := menus.NewMenu(108, "商品分类", "goodsClassify")
	menus.AddChildMenu(menuGoods, goodsListMenu)
	menus.AddChildMenu(menuGoods, classifyMenu)
	menus.AddChildMenu(menuGoods, goodsClassifyMenu)
	menus.AddNewMenu(msg, menuGoods)

	// 订单管理
	menuOrders := menus.NewMenu(120, "订单管理", "orders")
	orderListMenu := menus.NewMenu(121, "订单列表", "orderList")
	menus.AddChildMenu(menuOrders, orderListMenu)
	menus.AddNewMenu(msg, menuOrders)

	// 数据统计
	menuReports := menus.NewMenu(131, "数据统计", "reports")
	reportListMenu := menus.NewMenu(134, "数据列表", "reportList")
	menus.AddChildMenu(menuReports, reportListMenu)
	menus.AddNewMenu(msg, menuReports)

	c.JSON(http.StatusOK, msg)
}
