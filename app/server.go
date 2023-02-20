package main

import (
	"fmt"
	"net"
	"os"
	"regexp"
	"strconv"
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
			conn.Write([]byte("+" + responseBody + "\r\n"))
		}

		if reqCommand == "echo" {
			responseBody := ""
			for i := 0; i < len(reqArgs); i++ {
				responseBody += reqArgs[i]
			}
			conn.Write([]byte("+" + responseBody + "\r\n"))
		}

		if reqCommand == "set" {
			key, val := reqArgs[0], reqArgs[1]
			db[key] = val
			conn.Write([]byte("+OK\r\n"))
			if len(reqArgs) > 2 {
				option := reqArgs[2]
				if option == "px" {
					ms, _ := strconv.Atoi(reqArgs[3])
					expiry := time.Duration(ms)
					time.AfterFunc(expiry*time.Millisecond, func() {
						db[key] = "-1"
					})
				}
			}
		}
		if reqCommand == "get" {
			key := reqArgs[0]
			responseBody := db[key]
			if responseBody != "-1" {
				responseBody = "+" + responseBody + "\r\n"
			} else {
				responseBody = "*-1\r\n"
			}
			conn.Write([]byte(responseBody))
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
