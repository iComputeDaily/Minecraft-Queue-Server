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
	
	dataLength, legnth, err := ReadVarInt(data[:5])
	if err != nil {
		fmt.Println(err)
	}
	
	packetId, legnth2, err := ReadVarInt(data[legnth:legnth + 5])
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(legnth, legnth2)
	legnth += legnth2
	fmt.Println(legnth)
	
	connectionAddr, legnth32, err := ReadString(data[legnth + 1:])
	if err != nil {
		fmt.Println(err)
	}
	legnth32 += uint32(legnth)
	
	fmt.Printf("length: %d reported legnth: %d packet id: %d connection address: %s data: %b\n", n, dataLength, packetId, connectionAddr, data)
	return nil
}

func ReadVarInt(varInt []byte) (uint32, uint8, error) {
	if len(varInt) > 5 {
		return 0, 0, errors.New("Network: varInt Too Big!")
	}
	var result uint32
	for index, data := range varInt {
		value := uint32(data) & uint32(0b01111111)
		result = result | (value << (7 * index))
		if (data & 0b010000000) == 0 {
			return result, uint8(index) + 1, nil
		}
	}
	return 0, 0, errors.New("Network: idk what but something got f'd up while reading varInt")
}

func ReadVarlong(varLong []byte) (uint64, uint8, error) {
	if cap(varLong) >= 10 {
		return 0, 0, errors.New("Network: varLong Too Big!")
	}
	var result uint64
	for index, data := range varLong {
		value := uint64(data) & uint64(0b01111111)
		result = result | (value << (7 * index))
		if (data & 0b010000000) == 0 {
			return result, uint8(index) + 1, nil
		}
	}
	return 0, 0, errors.New("Network: idk what but something got f'd up while reading varLong")
}

func ReadString(bytes []byte) (string, uint32, error) {
	stringLegnth, legnthLegnth, err := ReadVarInt(bytes[:5])
	if err != nil {
		return "nil", 0, errors.New(err.Error())
	}
	fmt.Println("StringLegnth: ", stringLegnth, "legnthLegnth: ", legnthLegnth)
	stringEnd := stringLegnth + uint32(legnthLegnth)
	result := string(bytes[legnthLegnth:stringEnd])
	return result, stringLegnth + uint32(legnthLegnth), nil
}
