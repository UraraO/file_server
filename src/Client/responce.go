package main

type RspRegister struct {
	Message string `json:"message"`
}

type RspLogining struct {
	Message  string `json:"message"`
	Password string `json:"password"`
}

type RspLogin struct {
	Message string `json:"message"`
	Token   string `json:"token"`
	IP      string `json:"ip"`
	Port    string `json:"port"`
}
