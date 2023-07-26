package userServe

import (
	"FileServer/src/Server/FileTransport"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"net"
)

var db *gorm.DB

type User struct {
	gorm.Model
	Username string `gorm:"column:username;type:varchar(32);index:user_name;not null;unique"`
	Password string `gorm:"column:password;type:varchar(255);not null"`
}

func (u User) TableName() string {
	return "users"
}

const (
	MySQLUsername = "root"
	MySQLPassword = "123456"
	MySQLHost     = "localhost"
	MySQLPort     = 3306
	MySQLDatabase = "K_file_server"
)

func ConnectDB() error {
	// 连接数据库
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local", MySQLUsername, MySQLPassword, MySQLHost, MySQLPort, MySQLDatabase)
	DB, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	db = DB
	if err != nil {
		panic("failed to connect database, error=" + err.Error())
	}
	return err
}

func handleRegister(c *gin.Context) {
	var user User
	if err := c.ShouldBind(&user); err != nil {
		fmt.Println("gin.Context.ShouldBind ERROR,", err)
	}
	if err := db.Create(&user).Error; err != nil {
		fmt.Println("插入失败", err)
		c.JSON(200, gin.H{
			"message": "***username have exist!",
		})
		return
	}
	c.JSON(200, gin.H{
		"message": "***successfully register",
	})
}

func matchUser(lhs *User, rhs *User) bool {
	if lhs.Username == rhs.Username && lhs.Password == rhs.Password {
		return true
	}
	return false
}

func handleLogining(c *gin.Context) {
	var user User
	if err := c.ShouldBind(&user); err != nil {
		fmt.Println("gin.Context.ShouldBind ERROR,", err)
	}
	res := db.First(&user, "username = ?", user.Username)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		fmt.Println("user have not registered")
		c.JSON(200, gin.H{
			"message":  "username not exist",
			"password": "",
		})
		return
	}
	if res.Error == nil {
		c.JSON(200, gin.H{
			"message":  "return password",
			"password": user.Password,
		})
		return
	}
}

func handleLogin(c *gin.Context) {
	var user User
	if err := c.ShouldBind(&user); err != nil {
		fmt.Println("gin.Context.ShouldBind ERROR,", err)
	}
	fmt.Println("user have registered, and match")
	u := Userinfo{
		Username: user.Username,
		Password: user.Password,
	}
	token, _ := MakeToken(&u)
	um := fileTrans.NewUserModel(u.Username)
	um.Token = token
	fileTrans.ChLoginUserModel <- um
	c.JSON(200, gin.H{
		"message": "login!",
		"token":   token,
		"ip":      UserServerIP,
		"port":    UserServerPort,
	})
}

// HandlePreReq 接收用户上传下载请求，并将请求转发给FileServer
func HandlePreReq(conn net.Conn) {
	for {
		buffer := make([]byte, 1024)
		sz, err := conn.Read(buffer)
		if err != nil {
			fmt.Println("user_serve.HandlePreReq conn.Read error:", err)
			return
		} else if sz == 0 {
			fmt.Println("user_serve.HandlePreReq conn.Read sz == 0 error:", err)
			return
		}
		fmt.Println("User MainLoop, 125, conn.Read buffer:", string(buffer))
		msg := string(buffer[:sz-1]) // PreMsg原文
		/*for i := 1; i < len(splitMessage); i++ {
			println(splitMessage[i])
		}*/
		premsg := fileTrans.PreMsg{
			Msg:  msg,
			Conn: conn,
		}
		fileTrans.ChFileMsg2FileServer <- premsg
		fmt.Println("User MainLoop, 135, premsg put into ChFileMsg2FileServer")
	}
}

var UserServerIP = "127.0.0.1"
var UserServerPort = "8081"

// TODO FileServer从通道中接收一个UserModel时，需要初始化其Server为自身

func MainLoop() {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%s", UserServerIP, UserServerPort))
	defer listener.Close()
	// 与客户端建立Conn，初始化UserModel并转发给FileServer
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				fmt.Println("Server.Start, listener.Accept error: ", err)
				continue
			}
			fmt.Println("User MainLoop, 155, making um")
			um := <-fileTrans.ChLoginUserModel
			um.Conn = conn
			um.IP = conn.RemoteAddr().String()
			fmt.Println("User MainLoop, 159, client ip:", um.IP)
			fileTrans.ChConn2FileServer <- um
			go HandlePreReq(conn)
		}
	}()

	err = ConnectDB()
	r := gin.Default()
	r.POST("/api/register", handleRegister)
	r.POST("api/logining", handleLogining)
	r.POST("api/login", handleLogin)
	err = r.Run("127.0.0.1:8080")
	if err != nil {
		fmt.Println("user_serve router.Run ERROR", err)
		return
	}
}
