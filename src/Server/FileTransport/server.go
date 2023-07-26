package fileTrans

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
)

type FileServer struct {
	IP    string
	Port  string
	Users map[string]UserModel // 用户名-Model
	Mut   sync.Mutex
}

func NewFileServer(ip string, port string) *FileServer {
	fserver := &FileServer{
		IP:    ip,
		Port:  port,
		Users: make(map[string]UserModel),
	}
	return fserver
}

type PreMsg struct {
	Msg  string
	Conn net.Conn
}

var FileServerIP = "127.0.0.1"
var FileServerPort = "8082"

// ChLoginUserModel 用于登录时，与客户端建立首次连接Conn，转交给UserServer
var ChLoginUserModel = make(chan UserModel, 100)

// ChConn2FileServer 用于UserServer包装好用户信息UserModel后，读取用户信息
var ChConn2FileServer = make(chan UserModel, 100)

// ChFileMsg2FileServer 用于UserServer接到PreMsg时，将信息转交给FileServer
var ChFileMsg2FileServer = make(chan PreMsg, 100)

func HandleUploadFile(fileserver *FileServer, username string, filenames []string, conn net.Conn) {
	s := "PreACK\n"
	conn.Write([]byte(s))
	fmt.Println("HandleUploadFile, 47, conn.Write over")
	var wg sync.WaitGroup
	wg.Add(len(filenames))
	for i := 0; i < len(filenames); i++ {
		TTconn := <-fileserver.Users[username].ChTconn
		fmt.Println("HandleUploadFile, 50, TTconn got from ChTconn")
		TTconn.Tconn.Write([]byte("TconnFilenameACK\n"))
		fmt.Println("HandleUploadFile, 52, Tconn.Write over")
		go func(Tconn TconnType) {
			lenbuffer := make([]byte, 64)
			sz, err := Tconn.Tconn.Read(lenbuffer)
			if err != nil {
				fmt.Println("HandleUploadFile Tconn Read lenbuffer ERROR,", err)
				Tconn.Tconn.Close()
				wg.Done()
				return
			} else if sz == 0 {
				fmt.Println("HandleUploadFile Tconn Read lenbuffer sz = 0")
				wg.Done()
				return
			}
			fmt.Println("HandleUploadFile, 64, Tconn.Read lenbuffer is:", string(lenbuffer))
			filesize, _ := strconv.Atoi(string(lenbuffer[:sz-1]))
			fmt.Println("HandleUploadFile, 66, filesize is:", filesize)
			Tconn.Tconn.Write([]byte("TconnFilesizeACK\n"))
			fmt.Println("HandleUploadFile, 568, Tconn.Write FilesizeACK over")
			file := make([]byte, filesize)
			sz, _ = Tconn.Tconn.Read(file)
			fmt.Println("HandleUploadFile, 56, Tconn.Read file over")
			WriteF(file, Tconn.Filename+"2")
			wg.Done()
		}(TTconn)
	}
	wg.Wait()
	for i := 0; i < len(filenames); i++ {
		fileserver.Users[username].TransConns[filenames[i]].Close()
	}
}

func HandleDownloadFile(fileserver *FileServer, username string, filenames []string, conn net.Conn) {
	s := "PreACK\n"
	conn.Write([]byte(s))
	fmt.Println("HandleDownloadFile, 86 do")
	var wg sync.WaitGroup
	wg.Add(len(filenames))
	fmt.Println("HandleDownloadFile, 89 len(filenames):", len(filenames))
	for i := 0; i < len(filenames); i++ {
		TTconn := <-fileserver.Users[username].ChTconn
		fmt.Println("HandleDownloadFile, 92 TTconn got from ChTconn")
		if !CheckFileExist(TTconn.Filename) {
			fmt.Println("the file is not exist, please try again")
			RemoveTconn(TTconn, fileserver, username)
			wg.Done()
			continue
		}
		fd := ReadFile(TTconn.Filename)
		fmt.Println("HandleDownloadFile, 100 fd have read")
		if len(fd) == 0 {
			fmt.Println("the file you choose is an empty file")
			RemoveTconn(TTconn, fileserver, username)
			wg.Done()
			continue
		}
		TTconn.Tconn.Write([]byte("TconnFilenameACK\n"))
		fmt.Println("HandleDownloadFile, 108 Tconn Write filenameACK")
		go func(Tconn TconnType) {
			Tconn.Tconn.Write([]byte(strconv.Itoa(len(fd)) + "\n"))
			buffer := make([]byte, 20)
			sz, _ := Tconn.Tconn.Read(buffer)
			if string(buffer[:sz-1]) == "TconnFilesizeACK" {
				Tconn.Tconn.Write(fd)
			}
			wg.Done()
		}(TTconn)
	}
	wg.Wait()
}

func HandleAcceptTconn(fileserver *FileServer, listener net.Listener) { // 创建Tconn并将Tconn转交FileServer
	for {
		Tconn, err := listener.Accept()
		if err != nil {
			fmt.Println("Server.Start, listener.Accept error: ", err)
			continue
		}
		fmt.Println("FileServer.MainLoop, 129, Tconn conn over")
		TransTconnToFileserver(Tconn, fileserver)
		/*go func(Tconn net.Conn) { // Tconn接收文件信息，转交
			buffer := make([]byte, 128)
			sz, err := Tconn.Read(buffer)
			if err != nil {
				fmt.Println("Tconn transfer to FileServer Read ERROR,", err)
				return
			} else if sz == 0 {
				fmt.Println("Tconn transfer to FileServer Read sz = 0")
				return
			}
			fmt.Println("FileServer.MainLoop, 104, Tconn.Read buffer:", string(buffer))
			msg := string(buffer[:sz-1])
			splitedmsg := strings.Split(msg, "|")
			fmt.Println("FileServer.MainLoop, 107, splitedmsg:", splitedmsg)
			fileserver.Users[splitedmsg[0]].TransConns[splitedmsg[1]] = Tconn
			tc := TconnType{
				Tconn:    Tconn,
				Filename: splitedmsg[1],
			}
			fileserver.Users[splitedmsg[0]].ChTconn <- tc
			fmt.Println("FileServer.MainLoop, 114, tc have put into ChTconn")
			// 根据客户端的文件信息消息，将Tconn交由FileServer维护
		}(Tconn)*/
	}
}

func TransTconnToFileserver(Tconn net.Conn, fileserver *FileServer) { // Tconn接收文件信息，转交
	buffer := make([]byte, 128)
	sz, err := Tconn.Read(buffer)
	if err != nil {
		fmt.Println("Tconn transfer to FileServer Read ERROR,", err)
		return
	} else if sz == 0 {
		fmt.Println("Tconn transfer to FileServer Read sz = 0")
		return
	}
	fmt.Println("TransTconnToFileserver, 167, Tconn.Read buffer:", string(buffer))
	msg := string(buffer[:sz-1])
	splitedmsg := strings.Split(msg, "|")
	fmt.Println("TransTconnToFileserver, 170, splitedmsg:", splitedmsg)
	fileserver.Users[splitedmsg[0]].TransConns[splitedmsg[1]] = Tconn
	tc := TconnType{
		Tconn:    Tconn,
		Filename: splitedmsg[1],
	}
	fileserver.Users[splitedmsg[0]].ChTconn <- tc
	fmt.Println("TransToFileserver, 177, tc have put into ChTconn")
	// 根据客户端的文件信息消息，将Tconn交由FileServer维护
}

func MainLoop() {
	listener, _ := net.Listen("tcp", fmt.Sprintf("%s:%s", FileServerIP, FileServerPort))
	defer listener.Close()
	fileserver := NewFileServer(FileServerIP, FileServerPort)

	/*go func() { // 创建Tconn并将Tconn转交FileServer
		for {
			Tconn, err := listener.Accept()
			if err != nil {
				fmt.Println("Server.Start, listener.Accept error: ", err)
				continue
			}
			fmt.Println("FileServer.MainLoop, 93, Tconn conn over")
			go func(Tconn net.Conn) { // Tconn接收文件信息，转交
				buffer := make([]byte, 128)
				sz, err := Tconn.Read(buffer)
				if err != nil {
					fmt.Println("Tconn transfer to FileServer Read ERROR,", err)
					return
				} else if sz == 0 {
					fmt.Println("Tconn transfer to FileServer Read sz = 0")
					return
				}
				fmt.Println("FileServer.MainLoop, 104, Tconn.Read buffer:", string(buffer))
				msg := string(buffer[:sz-1])
				splitedmsg := strings.Split(msg, "|")
				fmt.Println("FileServer.MainLoop, 107, splitedmsg:", splitedmsg)
				fileserver.Users[splitedmsg[0]].TransConns[splitedmsg[1]] = Tconn
				tc := TconnType{
					Tconn:    Tconn,
					Filename: splitedmsg[1],
				}
				fileserver.Users[splitedmsg[0]].ChTconn <- tc
				fmt.Println("FileServer.MainLoop, 114, tc have put into ChTconn")
				// 根据客户端的文件信息消息，将Tconn交由FileServer维护
			}(Tconn)
		}
	}()*/
	go HandleAcceptTconn(fileserver, listener)

	// 处理ConnChannel的转交信息
	go func() {
		for {
			um := <-ChConn2FileServer
			fileserver.Users[um.Name] = um
			fmt.Println("FileServer.MainLoop, 226, got um from ChConn2FileServer:", um.Name)
		}
	}()

	for { // 从ChFileMsg2FileServer中取出PreMsg
		premsg := <-ChFileMsg2FileServer
		fmt.Println("FileServer.MainLoop, 232, got premsg from ChFileMsg2FileServer", premsg.Msg)
		splitMessage := strings.Split(premsg.Msg, "|")
		username := splitMessage[0]
		UorD := splitMessage[1]
		if UorD == "U" {
			fmt.Println("FileServer.MainLoop, 237, go HandleUploadFile")
			go HandleUploadFile(fileserver, username, splitMessage[2:], premsg.Conn)
		}
		if UorD == "D" {
			fmt.Println("FileServer.MainLoop, 241, go HandleDownloadFile")
			go HandleDownloadFile(fileserver, username, splitMessage[2:], premsg.Conn)
		}
	}
}
