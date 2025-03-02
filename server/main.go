package main

import (
	"cmd/server/config"
	db "cmd/server/model/init"
	"cmd/server/handle/agent/install"
	"cmd/server/handle/server/monitor" // 引入 monitor 包
	"cmd/server/handle/user/login"
	"cmd/server/middlewire"
	"log"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	go monitor.CheckServerStatus()
	//读取DBConfig.yaml文件
	config, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}
	//设置数据库连接的环境变量
	os.Setenv("DB_USER", config.DB.User)
	os.Setenv("DB_PASSWORD", config.DB.Password)
	os.Setenv("DB_HOST", config.DB.Host)
	os.Setenv("DB_PORT", config.DB.Port)
	os.Setenv("DB_NAME", config.DB.Name)
	
	router := gin.Default()// 连接数据库
	if err := db.ConnectDatabase(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	// 初始化数据库
	if err := db.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	
	router.POST("/agent/register", login.Register)
	router.POST("/agent/login", login.Login)
	// 需要 JWT 认证的路由
	auth := router.Group("/agent", middlewire.JWTAuthMiddleware())
	{
		auth.POST("/install", install.InstallAgent)
		auth.POST("/system_info", monitor.GetMessage)
		auth.GET("/list", monitor.ListAgent)
		router.GET("/:hostname", monitor.GetAgentInfo)
	}
	router.Run("0.0.0.0:8080")
}
