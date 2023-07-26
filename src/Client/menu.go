package main

import (
	fileTrans "FileServer/src/Server/FileTransport"
	"fmt"
	"github.com/howeyc/gopass"
	"net"
	"strconv"
)

// Menu 根据CliStatus输出操作提示，菜单信息
func (thisClient *Client) Menu(status int) {
	switch status {
	case Init:
		if thisClient.firstin {
			fmt.Println("\nHello! Welcome to this simple file transport tool, operation method is here, print number to enter any of selection:")
			thisClient.firstin = false
		}
		fmt.Println("*** Start ***")
		fmt.Println("1.Register an account")
		fmt.Println("2.Login")
		fmt.Println("3.Quit")
	case Registering:
		fmt.Println("*** Registering ***")
	case Login:
		fmt.Println("*** Login ***")
	case Menu:
		fmt.Println("*** Menu ***")
		fmt.Println("1.Upload File")
		fmt.Println("2.Download File")
		fmt.Println("3.Quit")
	case Upload:
		fmt.Println("*** Upload ***")
		fmt.Println("1.one file")
		fmt.Println("2.some files")
		fmt.Println("3.back")
	case UploadOne:
		fmt.Println("*** Upload One ***")
		fmt.Println("print the absolute path of the file, \"ENTER\" to upload, \"QUIT\" to quit")
	case UploadMulti:
		fmt.Println("*** Upload Multi ***")
		fmt.Println("print the absolute path of the file, \"ENTER\" to continue, double \"ENTER\" to upload all, \"QUIT\" to quit")
	case Download:
		fmt.Println("*** Download ***")
		fmt.Println("1.one file")
		fmt.Println("2.some files")
		fmt.Println("3.back")
	case DownloadOne:
		fmt.Println("*** Download One ***")
		fmt.Println("print the absolute path of the file, \"ENTER\" to download, \"QUIT\" to quit")
	case DownloadMulti:
		fmt.Println("*** Download Multi ***")
		fmt.Println("print the absolute path of the file, \"ENTER\" to continue, double \"ENTER\" to download all, \"QUIT\" to quit")
	case Quit:
		fmt.Println("Quit")
	default:
		fmt.Println("***Menu default case, some BUG happen")
	}

}

func (thisClient *Client) HandleInit() {
	var cmd string
	for thisClient.CliStatus == Init {
		_, err := fmt.Scanln(&cmd)
		if err != nil {
			fmt.Println("Client.HandleInit Scanln error,", err)
			return
		}
		if cmd == "1" { // Register an account
			thisClient.CliStatus = Registering
		} else if cmd == "2" { // Login
			thisClient.CliStatus = Login
		} else if cmd == "3" { // Quit
			thisClient.CliStatus = Quit
		} else {
			fmt.Println("invalid command, please type again:")
		}
	}
}

func (thisClient *Client) HandleRegister() {
	var username string
	var passwordClear string
	fmt.Printf("your Username:")
	_, err := fmt.Scanln(&username)
	if err != nil {
		fmt.Println("Scan username ERROR,", err)
		return
	}
	fmt.Printf("your Password:")
	// _, err = fmt.Scanln(&password_clear)
	bytePwd, err := gopass.GetPasswdMasked()
	if err != nil {
		fmt.Println("Scan password ERROR,", err)
		return
	}
	passwordClear = string(bytePwd)
	pwd, err := GetPwd(passwordClear)
	if err != nil {
		fmt.Println("Get Bcrypt password ERROR,", err)
		return
	}
	// thisClient.Password = pwd
	thisClient.CliStatus = Init
	if SendRegisterReq(username, pwd) {
		// fmt.Println("successfully register!")
	} else {
		fmt.Println("register false, please try again")
	}
}

func (thisClient *Client) HandleLogin() {
	var username string
	var passwordClear string
	fmt.Printf("your Username:")
	_, err := fmt.Scanln(&username)
	if err != nil {
		fmt.Println("Scan username ERROR,", err)
		return
	}
	fmt.Printf("your Password:")
	bytePwd, err := gopass.GetPasswdMasked()
	if err != nil {
		fmt.Println("Scan password ERROR,", err)
		return
	}
	passwordClear = string(bytePwd)
	pwd, err := GetPwd(passwordClear)
	if err != nil {
		fmt.Println("Get Bcrypt password ERROR,", err)
		return
	}
	if SendLoginingReq(username, pwd, passwordClear, thisClient) {
		// fmt.Println("successfully register!")
	} else {
		// fmt.Println("register false, please try again")
	}
}

func (thisClient *Client) HandleMenu() {
	var cmd string
	for thisClient.CliStatus == Menu {
		_, err := fmt.Scanln(&cmd)
		if err != nil {
			fmt.Println("Client.HandleMenu Scanln error,", err)
			return
		}
		if cmd == "1" { // upload file
			thisClient.CliStatus = Upload
		} else if cmd == "2" { // download file
			thisClient.CliStatus = Download
		} else if cmd == "3" { // Quit
			thisClient.CliStatus = Quit
		} else {
			fmt.Println("invalid command, please type again:")
		}
	}
}

func (thisClient *Client) HandleUploadMenu() {
	var cmd string
	for thisClient.CliStatus == Upload {
		_, err := fmt.Scanln(&cmd)
		if err != nil {
			fmt.Println("Client.HandleUploadMenu Scanln error,", err)
			return
		}
		if cmd == "1" { // upload 1
			thisClient.CliStatus = UploadOne
		} else if cmd == "2" { // upload multi
			thisClient.CliStatus = UploadMulti
		} else if cmd == "3" { // Quit
			thisClient.CliStatus = Menu
		} else {
			fmt.Println("invalid command, please type again:")
		}
	}
}

func (thisClient *Client) HandleUploadOne() {
	fmt.Println("***please input the file you want to upload:")
	var filename string
	fmt.Scanf("%s", &filename)
	if filename == "QUIT" {
		thisClient.CliStatus = Upload
		return
	}
	if !CheckFileExist(filename) {
		fmt.Println("the file is not exist, please try again")
		return
	}
	fd := ReadFile(filename)
	if len(fd) == 0 {
		fmt.Println("the file you choose is an empty file")
		return
	}
	preMsg := thisClient.Username + "|" + "U" + "|"
	preMsg += filename
	thisClient.conn.Write([]byte(preMsg + "\n"))
	buffer := make([]byte, 10)
	sz, _ := thisClient.conn.Read(buffer)
	if string(buffer[:sz-1]) == "PreACK" {
		Tconn, _ := net.Dial("tcp", fmt.Sprintf("%s:%s", fileTrans.FileServerIP, fileTrans.FileServerPort))
		Tconn.Write([]byte(thisClient.Username + "|" + filename + "\n"))
		buffer := make([]byte, 20)
		sz, _ := Tconn.Read(buffer)
		if string(buffer[:sz-1]) == "TconnFilenameACK" {
			Tconn.Write([]byte(strconv.Itoa(len(fd)) + "\n"))
			buffer := make([]byte, 20)
			sz, _ = Tconn.Read(buffer)
			if string(buffer[:sz-1]) == "TconnFilesizeACK" {
				Tconn.Write(fd)
			}
			Tconn.Close()
		}
	}
}

func (thisClient *Client) HandleUploadMulti() {

}

func (thisClient *Client) HandleDownloadMenu() {
	var cmd string
	for thisClient.CliStatus == Download {
		_, err := fmt.Scanln(&cmd)
		if err != nil {
			fmt.Println("Client.HandleDownloadMenu Scanln error,", err)
			return
		}
		if cmd == "1" { // download 1
			thisClient.CliStatus = DownloadOne
		} else if cmd == "2" { // download multi
			thisClient.CliStatus = DownloadMulti
		} else if cmd == "3" { // Quit
			thisClient.CliStatus = Menu
		} else {
			fmt.Println("invalid command, please type again:")
		}
	}
}

func (thisClient *Client) HandleDownloadOne() {
	fmt.Println("***please input the file you want to download:")
	var filename string
	fmt.Scanf("%s", &filename)
	if filename == "QUIT" {
		thisClient.CliStatus = Download
		return
	}
	preMsg := thisClient.Username + "|" + "D" + "|"
	preMsg += filename
	thisClient.conn.Write([]byte(preMsg + "\n"))
	fmt.Println("HandleDownloadOne 254 conn Write over")
	buffer := make([]byte, 10)
	sz, _ := thisClient.conn.Read(buffer)
	fmt.Println("HandleDownloadOne 257 conn Read over")
	if string(buffer[:sz-1]) == "PreACK" {
		fmt.Println("HandleDownloadOne 259 PreACK over")
		Tconn, _ := net.Dial("tcp", fmt.Sprintf("%s:%s", fileTrans.FileServerIP, fileTrans.FileServerPort))
		Tconn.Write([]byte(thisClient.Username + "|" + filename + "\n"))
		fmt.Println("HandleDownloadOne 262 Tconn and Write over")
		buffer := make([]byte, 20)
		sz, _ := Tconn.Read(buffer)
		if string(buffer[:sz-1]) == "TconnFilenameACK" {
			fmt.Println("HandleDownloadOne 266 TconnFilenameACK")
			lenbuffer := make([]byte, 64)
			sz, err := Tconn.Read(lenbuffer)
			if err != nil {
				fmt.Println("HandleDownloadOne Tconn Read lenbuffer ERROR,", err)
				Tconn.Close()
				return
			} else if sz == 0 {
				fmt.Println("HandleDownloadOne Tconn Read lenbuffer sz = 0")
				return
			}

			filesize, _ := strconv.Atoi(string(lenbuffer[:sz-1]))
			fmt.Println("HandleDownloadOne 279 filesize:", filesize)
			Tconn.Write([]byte("TconnFilesizeACK\n"))
			fmt.Println("HandleDownloadOne 281 Tconn Write sizeACK")
			file := make([]byte, filesize)
			sz, _ = Tconn.Read(file)
			WriteF(file, filename+"3")
			Tconn.Close()
		}
	}
}

func (thisClient *Client) HandleDownloadMulti() {

}
