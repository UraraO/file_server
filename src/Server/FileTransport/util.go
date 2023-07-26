package fileTrans

import (
	"fmt"
	"io"
	"os"
	"strings"
)

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

func SplitTest() {
	msg := "Urara|file1|file2"
	splitMessage := strings.Split(msg, "|")
	fmt.Println(splitMessage)
	for i := 1; i < len(splitMessage); i++ {
		println(splitMessage[i])
	}
}

func RemoveTconn(Tconn TconnType, fileserver *FileServer, username string) {
	Tconn.Tconn.Close()
	delete(fileserver.Users[username].TransConns, Tconn.Filename)
}
