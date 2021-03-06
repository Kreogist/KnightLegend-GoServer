package main

import (
	"net"
	"strconv"
)

var maxRead = 600
var clientPool [1000]net.Conn

var EVERYONE = 1
var ITSELF = 2

func main() {
	hostAndPort := "localhost:9825"
	listener := initServer(hostAndPort)

	for {
		var sn = findFreePool()
		conn, err := listener.Accept()
		clientPool[sn] = conn
		checkError(err, "Accept: ")
		go connectionHandler(clientPool[sn], sn)
	}
}

func initServer(hostAndPort string) *net.TCPListener {
	serverAddr, err := net.ResolveTCPAddr("tcp", hostAndPort)
	checkError(err, "Resolving address:port failed: '"+hostAndPort+"'")

	listener, err := net.ListenTCP("tcp", serverAddr)
	checkError(err, "ListenTCP: ")
	println("Listening to: ", listener.Addr().String())
	return listener
}

func connectionHandler(conn net.Conn, sn int) {
	//aliveTimer := time.NewTimer(time.Second * 60)
	connFrom := conn.RemoteAddr().String()
	defer conn.Close()

	println("Connection from: ", connFrom)

	/*go func() {
		<-aliveTimer.C
		conn.Close()
		clientPool[sn]=nil
	}()*/

	for {
		var buffer []byte = make([]byte, maxRead+1)
		length, err := conn.Read(buffer[0:maxRead])
		buffer[maxRead] = 0

		switch err {
		case nil:
			{
				//aliveTimer.Reset(0)
				handleMessage(string(buffer[0:length]), conn)
			}
		default:
			{
				conn.Close()
				clientPool[sn] = nil
				println("Disconnection from:", connFrom)
				return
			}
		}
	}
}

func handleMessage(msg string, socket net.Conn) {
	/*
		COUNT Format:
			{xxxxx}
		POSITION Format:
			{xxx,yyy}
		Message Format:
			{"TYPE","ID-Length","ID","DATA"}
			    4bit    2bit	20bit EOF
		DATA Format:
			{"TYPE"}
				{"UPLE","ATTK","CHAT","RMAP","BUYS","BUYD"}

				UPLE Format:
					{"CASTLE","WALL","...}
				ATTK Format:
					{"Target Position","Unit Type Count",{"UNIT TYPE","COUNT"},...}
				CHAT Format:
					{"Message"}
				RMAP Format:
					{"RMAP"}
				BUYS Format:
					{"TYPE","COUNT"}
				BUYD Format:
					{"COUNT"}
	*/
	var msgType string
	var IDLenght = 0
	var msgAuthor string
	var msgData string

	length := len(msg)
	msgType = msg[0:3]
	IDLenght, err := strconv.Atoi(msg[4:5])
	msgAuthor = msg[6 : 6+IDLenght]
	msgData = msg[7+IDLenght : length]

	checkError(err, "Convte String to int failed")

	print(msg, "\n")
	switch msgType {
	case "CHAT":
		{
			sendTO(EVERYONE, socket, msgAuthor+":"+msgData)
			print("Chat Message\n")
		}
	}
}

func sendTO(target int, socket net.Conn, msg string) {
	switch target {
	case EVERYONE:
		{
			var i = 0
			for i = 0; i < 1000; i++ {
				if clientPool[i] != nil {
					clientPool[i].Write([]byte(msg))
				}
			}
		}
	case ITSELF:
		{
			socket.Write([]byte(msg))
		}
	}
}

func checkError(error error, info string) {
	if error != nil {
		panic("ERROR: " + info + " " + error.Error())
	}
}

func findFreePool() int {
	var i = 0
	for i = 0; i < 1000; i++ {
		if clientPool[i] == nil {
			return i
		}
	}
	return -1
}
