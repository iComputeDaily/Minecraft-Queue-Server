package main

import "fmt"
import "os"
import "netprocess"

func main() {
	fmt.Println("Starting Server...")
	listener, err := netprocess.StartListener()
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	defer listener.Close()
	
	for {
		connection, err := listener.Accept()
		if err != nil {
			fmt.Println("Failed To Recive Connection!")
			fmt.Println(err)
		}
		go netprocess.HandleConnection(connection)
	}
}
