package main

import (
	fileTrans "FileServer/src/Server/FileTransport"
	userServe "FileServer/src/Server/User"
)

func main() {
	go userServe.MainLoop()
	fileTrans.MainLoop()
}
