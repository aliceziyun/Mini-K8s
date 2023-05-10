package tcp

import (
	"fmt"
	"net"
)

func Client(addr string, str string) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Println("[ERROR] (Client) net.Dial err: ", err)
		return
	}
	defer conn.Close()

	conn.Write([]byte(str))

}

func Server(addr string) {
	listen, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Println("[ERROR] (Server) net.Listen err: ", err)
		return
	}
	defer listen.Close()

	// waiting for the client to establish a connection
	fmt.Println("Server is waiting...")

	conn, err := listen.Accept()
	if err != nil {
		fmt.Println("[ERROR] (Server) listen.Accept() err: ", err)
		return
	}
	defer conn.Close()

	fmt.Println("\tEstablish connection.")

	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Println("[ERROR] (Server) conn.Read() err: ", err)
		return
	}

	fmt.Println("\tServer receive: ", string(buf[:n]))

}
