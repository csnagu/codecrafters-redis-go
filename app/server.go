package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"regexp"
	"time"
)

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}
	defer l.Close()

	// as in-memory db
	var db = map[string]string{}

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		go pong(conn, db)
	}
}

func pong(conn net.Conn, db map[string]string) {
	defer conn.Close()
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	request := make([]byte, 128)

	for {
		readLen, err := conn.Read(request)
		if err != nil {
			fmt.Println("read error:", err)
			break
		}
		if readLen == 0 {
			break // connection already closed by client
		}

		reqCommand, reqArgs := parseRequest(request)

		if reqCommand == "ping" {
			responseBody := "PONG"
			_, err := io.WriteString(conn, "+"+responseBody+"\r\n")
			if err != nil {
				fmt.Println("io write error:", err)
				break
			}
		}

		if reqCommand == "echo" {
			responseBody := ""
			for i := 0; i < len(reqArgs); i++ {
				responseBody += reqArgs[i]
			}
			_, err = io.WriteString(conn, "+"+responseBody+"\r\n")
			if err != nil {
				fmt.Println("io write error:", err)
				break
			}
		}

		if reqCommand == "set" {
			key, val := reqArgs[0], reqArgs[1]
			db[key] = val
			_, err = io.WriteString(conn, "+OK\r\n")
			if err != nil {
				fmt.Println("io write error:", err)
				break
			}
		}
		if reqCommand == "get" {
			fmt.Println("get")
			key := reqArgs[0]
			responseBody := db[key]
			if db[key] == "" {
				responseBody = "nil"
			}
			_, err = io.WriteString(conn, "+"+responseBody+"\r\n")
			if err != nil {
				fmt.Println("io write error:", err)
				break
			}
		}
	}
}

func parseRequest(request []byte) (reqCommand string, reqArgs []string) {
	req := string(request)
	// ex.) *2\r\n$4\r\necho\r\n$5\r\nworld\r\n

	rep := regexp.MustCompile(`\r\n`)
	resultArr := rep.Split(req, -1)
	for i := 0; i < len(resultArr); i++ {
		symbol := resultArr[i][:1]

		if symbol == "*" || symbol == "$" {
			continue
		}
		if reqCommand == "" {
			reqCommand = resultArr[i]
		} else {
			reqArgs = append(reqArgs, resultArr[i])
		}
	}
	// remove EOF
	reqArgs = reqArgs[:len(reqArgs)-1]

	return reqCommand, reqArgs
}
