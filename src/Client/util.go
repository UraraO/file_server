package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/howeyc/gopass"
	"golang.org/x/crypto/bcrypt"
	"io"
	"net"
	"net/http"
	"os"
)

type UserSend struct {
	Username string
	Password string
}

func SendRegisterReq(username string, psw string) bool {
	user := UserSend{
		Username: username,
		Password: psw,
	}
	body, _ := json.Marshal(user)
	resp, err := http.Post("http://127.0.0.1:8080/api/register", "application/json", bytes.NewBuffer(body))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		respBody, err := io.ReadAll(resp.Body)
		var result RspRegister
		if err = json.Unmarshal(respBody, &result); err != nil {
			fmt.Println("SendRegisterReq json.Unmarshal ERROR,", err)
			return false
		}
		fmt.Println(result.Message)
		return true
	} else {
		fmt.Println("register resp error:", resp.Status)
		return false
	}
}

func SendLoginingReq(username string, psw string, pswclear string, thisClient *Client) bool {
	user := UserSend{
		Username: username,
		Password: psw,
	}
	body, _ := json.Marshal(user)
	resp, err := http.Post("http://127.0.0.1:8080/api/logining", "application/json", bytes.NewBuffer(body))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		respBody, err := io.ReadAll(resp.Body)
		var result RspLogining
		if err = json.Unmarshal(respBody, &result); err != nil {
			fmt.Println("SendLoginReq json.Unmarshal ERROR,", err)
			return false
		}
		// fmt.Println(result.Message)
		if result.Message == "username not exist" {
			fmt.Println("username is not exist, please try again")
			return false
		}
		if ComparePwd(result.Password, pswclear) {
			SendLoginReq(user.Username, user.Password, thisClient)
			fmt.Println("password right, success login~")
		} else {
			fmt.Println("password wrong, please try again")
		}
		return true
	} else {
		fmt.Println("logining resp error:", resp.Status)
		return false
	}
}

func SendLoginReq(username string, psw string, thisClient *Client) bool {
	user := UserSend{
		Username: username,
		Password: psw,
	}
	body, _ := json.Marshal(user)
	resp, err := http.Post("http://127.0.0.1:8080/api/login", "application/json", bytes.NewBuffer(body))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		respBody, err := io.ReadAll(resp.Body)
		var result RspLogin
		if err = json.Unmarshal(respBody, &result); err != nil {
			fmt.Println("SendLoginReq json.Unmarshal ERROR,", err)
			return false
		}
		if result.Message == "login!" {
			conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", result.IP, result.Port))
			if err != nil {
				fmt.Println("login but connect to server ERROR,", err)
				return false
			}
			thisClient.conn = conn
			thisClient.Username = user.Username
			thisClient.Password = user.Password
			thisClient.IsLogin = true
			thisClient.JwtToken = result.Token
			thisClient.CliStatus = Menu
			// fmt.Println(conn.RemoteAddr().String())
		}
		return true
	} else {
		fmt.Println("login resp error:", resp.Status)
		return false
	}
}

// PrintPassword 命令行输入密码显示成星号，返回密码明文
func PrintPassword() string {
	bpwd, err := gopass.GetPasswdMasked()
	if err != nil {
		fmt.Println(err)
		return ""
	}
	pwd := string(bpwd)
	return pwd
}

// GetPwd 给密码加密
func GetPwd(pwd string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	return string(hash), err
}

// ComparePwd 比对密码
func ComparePwd(pwd1 string, pwd2 string) bool {
	// Returns true on success, pwd1 is for the database.
	err := bcrypt.CompareHashAndPassword([]byte(pwd1), []byte(pwd2))
	if err != nil {
		return false
	} else {
		return true
	}
}

func CheckFileExist(fileName string) bool {
	_, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		return false
	}
	return true
}

func ReadFile(filepath string) []byte {
	f, err := os.OpenFile(filepath, os.O_RDWR, os.ModeTemporary)
	if err != nil {
		fmt.Println("read file fail", err)
		return []byte{}
	}
	defer f.Close()

	fd, err := io.ReadAll(f)
	if err != nil {
		fmt.Println("read to fd fail", err)
		return []byte{}
	}
	// fmt.Println(string(fd))
	return fd
}

func WriteF(src []byte, fileName string) {
	err := os.WriteFile(fileName, src, 0666)
	if err != nil {
		fmt.Println("write fail")
	}
	fmt.Println("write success")
}
