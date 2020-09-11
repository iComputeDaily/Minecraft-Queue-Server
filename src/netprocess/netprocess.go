package netprocess

import "errors"
import "fmt"
import "net"

func StartListener() (net.Listener, error) {
	listener, err := net.Listen("tcp", ":25565")
	if err != nil {
		return nil, errors.New("Network: Could Not Bind To Port!")
	}
	return listener, nil
}

func HandleConnection(connection net.Conn) error {
	data := make([]byte, 17)
	n, err := connection.Read(data)
	if err != nil {
		return errors.New("Network: Could Not Read TCP Packet!")
	}
	
	fmt.Printf("length %d: %b\n", n, data)
	return nil
}
