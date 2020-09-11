package main

import "fmt"
import "net"
import "os"

func startListener() net.Listener {
	listener, err := net.Listen("tcp", ":25565")
	if err != nil {
		fmt.Println("Could Not Start Server. Aborting!")
		fmt.Println(err)
		os.Exit(1)
	}
	return listener
}

func handleConnection(connection net.Conn){
	data := make([]byte, 17)
	n, err := connection.Read(data)
	if err != nil {
		fmt.Println("Failed To Read Data!")
		fmt.Println(err)
	}
	
	fmt.Printf("length %d: %b\n", n, data)
}

func main() {
	fmt.Println("Starting Server...")
	listener := startListener()
	defer listener.Close()
	
	for {
		connection, err := listener.Accept()
		if err != nil {
			fmt.Println("Failed To Recive Connection!")
			fmt.Println(err)
		}
		go handleConnection(connection)
	}
}
