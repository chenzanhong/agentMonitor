package login

import (
	"cmd/server/middlewire"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"gorm.io/gorm"

	m_init "cmd/server/model/init"
	u "cmd/server/model/user"
)

// 初始化数据库并创建用户表
func InitDB() (*sql.DB, error) {
	// 连接 PostgreSQL 数据库，替换连接信息
	connStr := "host=192.168.31.251 port=5432 user=postgres password=cCyjKKMyweCer8f3 dbname=monitor sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("连接数据库时出错: %v", err)
	}

	// 检查数据库连接
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("检查连接时出错: %v", err)
	}
	return db, err
}

// 注册
func Register(c *gin.Context) {
	// 定义用于接收 JSON 数据的结构体
	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	// 解析 JSON 数据
	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "请求数据格式错误"})
		return
	}

	// 数据验证
	if len(input.Name) == 0 || len(input.Email) == 0 || len(input.Password) < 6 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "参数不完整或密码长度不足"})
		return
	}

	// 检查邮箱是否存在
	var user u.User
	err := m_init.DB.Where("email = ?", input.Email).First(&user).Error
	if err == nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "用户已存在"})
		return
	}else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		// 如果 err 不为 nil 且不是因为记录未找到导致的，则是其他数据库错误
		c.JSON(http.StatusInternalServerError, gin.H{"message": "数据库查询失败"})
		return
	}

	// 创建用户
	newUser := u.User{
        Name:       input.Name,
        Email:      input.Email,
        Password:   input.Password,
        IsVerified: true,
    }

	err = m_init.DB.Create(&newUser).Error;
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "用户创建失败"})
		return
	}

	// 返回结果，包括加密后的密码
	c.JSON(http.StatusOK, gin.H{
		"message": "注册成功",
	})
}

// 登录
func Login(c *gin.Context) {
	// 定义用于接收 JSON 数据的结构体
	var input struct {
		Name     string `json:"name"`
		Password string `json:"password"`
	}

	// 解析 JSON 数据
	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "请求数据格式错误"})
		return
	}

	// 查找用户
	var user u.User
	err := m_init.DB.Where("name = ?", input.Name).First(&user).Error;
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "用户不存在"})
		return
	}

	// 验证密码
	if user.Password != input.Password {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "密码错误"})
		return
	}

	// 生成 JWT
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &middlewire.Claims{
		Username: input.Name,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(middlewire.JwtKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "生成 token 错误"})
		return
	}

	// 登录成功
	c.JSON(http.StatusOK, gin.H{
		"message": "登录成功",
		"token":   tokenString,
	})
}
