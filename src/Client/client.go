/*
//	客户端结构定义
//	客户端主循环
//	具体业务逻辑位于util和menu中
*/

package main

import "net"

type Client struct {
	Username      string
	Password      string
	CliStatus     int
	IsLogin       bool
	IsDownloading bool
	IsUploading   bool
	JwtToken      string
	conn          net.Conn
	TransConns    map[string]net.Conn // 文件名映射连接
	firstin       bool
}

// CliStatus,标识客户端操作界面和状态
const (
	Init          = 0
	Registering   = 1
	Login         = 2
	Menu          = 100
	Upload        = 200
	UploadOne     = 201
	UploadMulti   = 202
	Download      = 300
	DownloadOne   = 301
	DownloadMulti = 302
	Quit          = -1
)

func NewClient() Client {
	cli := Client{
		Username:      "",
		Password:      "",
		JwtToken:      "",
		CliStatus:     Init,
		IsLogin:       false,
		IsUploading:   false,
		IsDownloading: false,
		firstin:       true,
		conn:          nil,
	}
	return cli
}

func (thisClient *Client) Run() {
	for {
		thisClient.Menu(thisClient.CliStatus)
		switch thisClient.CliStatus {
		case Init:
			thisClient.HandleInit()
		case Registering:
			thisClient.HandleRegister()
		case Login:
			thisClient.HandleLogin()
		case Menu:
			thisClient.HandleMenu()
		case Upload:
			thisClient.HandleUploadMenu()
		case UploadOne:
			thisClient.HandleUploadOne()
		case UploadMulti:
			thisClient.HandleUploadMulti()
		case Download:
			thisClient.HandleDownloadMenu()
		case DownloadOne:
			thisClient.HandleDownloadOne()
		case DownloadMulti:
			thisClient.HandleDownloadMulti()
		case Quit:
			return
		default:
			thisClient.CliStatus = Init
		}
	}
}

func main() {
	Cli := NewClient()
	Cli.Run()
}
