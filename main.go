// All roads lead to ro*cough* main
package main

import "fmt"
import "os"
import "crypto/rand"
import "crypto/rsa"
import mcnet "github.com/Tnze/go-mc/net"
import "github.com/Tnze/go-mc/net/packet"

var PrivKey *rsa.PrivateKey

// startListener is called once to start listening for connections.
func startListener() *mcnet.Listener {
	listener, err := mcnet.ListenMC(":25565")
	if err != nil {
		fmt.Println("Could Not Start Server. Aborting! Error: ", err)
		os.Exit(1)
	}
	return listener
}

// handleConnection is called on al incomming connections.
func handleConnection(connection mcnet.Conn) {
	defer connection.Close()
	
	var (
		Protocol, Intention packet.VarInt
		ServerAddress       packet.String        // ignored
		ServerPort          packet.UnsignedShort // ignored
	)
	
	data, err := connection.ReadPacket()
	if err != nil {
		fmt.Println("Failed To Read Handshake Packet! Error: ", err)
		return
	}
	
	err = data.Scan(&Protocol, &ServerAddress, &ServerPort, &Intention)
	if err != nil {
		fmt.Println("Failed To Parse Handshake Packet! Error: ", err)
		return
	}
	
	fmt.Printf("Packet Recived: %b\n\tProtocal: %d\n\tServer Address: %s\n\tServer Port: %d\n\tIntention: %d\n", data, Protocol, ServerAddress, ServerPort, Intention)
	
	switch Intention {
		default:
			fmt.Println("Unknown Intention!")
			return
		case 1:
			handlePing(connection)
		case 2:
			var player Player
			player.connection = connection
			player.handleLogin()
	}
	
}

// handlePing is called by handleConnection on any connections with handshake intention 1(status).
func handlePing(connection mcnet.Conn) {
	for packetNum := 0; packetNum < 2; packetNum++ {
		data, err := connection.ReadPacket()
		if err != nil {
			fmt.Println("Failed To Read Ping Packet! Error: ", err)
			return
		}
		
		switch data.ID {
			default:
				fmt.Println("Invalid Ping Packet Id! ID: ", data.ID)
				return
			case 0x00:
				err = connection.WritePacket(packet.Marshal(0x00, packet.String(`{"version":{"name":"1.16.2","protocol":751},"players":{"max":5,"online":700,"sample":[{"name":"WatterBottle","id":"c8fe4d7e-9d7f-49f8-ba19-3ec7a22f62a6"}]},"description":{"text":"This is an amazing server"},"favicon":"data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAEAAAABAAQMAAACQp+OdAAABb2lDQ1BpY2MAACiRdZE7SwNBFIW/JIqvSEAtRERSxEeRgCiIpcYiTZAQIxi1SdY8hDyW3QQJtoKNRcBCtPFV+A+0FWwVBEERRCz8Bb4aCesdE0iQOMvs/Tgz5zJzBuzBjJY1W8YhmysY4YDfvRRddre90oGTHkYZimmmPhsKBfl3fN1jU/XOp3r9v6/p6FpLmBrY2oWnNN0oCM8IBzcKuuId4T4tHVsTPhL2GnJA4Wulx6v8ojhV5Q/FRiQ8B3bV051q4HgDa2kjKzwm7MlmilrtPOomzkRucUHqgMxBTMIE8OMmTpF1MhTwSc1JZs1947++efLi0eSvU8IQR4q0eL2iFqVrQmpS9IR8GUoq9795msnJiWp3px9any3rfRjadqFStqzvY8uqnIDjCS5zdX9ecpr+FL1c1zyH4NqC86u6Ft+Di23of9RjRuxXcsi0J5PwdgbdUei9hc6Vala1dU4fILIpT3QD+wcwIvtdqz/y22gEXjTGVwAAAAZQTFRF////AAAAVcLTfgAAAAlwSFlzAAAAJwAAACcBKgmRTwAAAV9JREFUKM9tkTFPwkAYhg9CaeJEKowmgDYNCwmb2qUSjpStmF6PwSYuWOPUNL3cWjZGonFwYDSp8Q8QXGDjB+jIf/E9jBP9pifv3fvd995HyH8JqcciBpgXZcd0AJZVjq17pbTKtYNi4o4Zk6KKpC65OvL7Fep7gIdEj4KDYpSGvuoTJbqwvEK7kJWV/ZURQmk1b3egTCMtt68uoQy1/KQHJRGDlf2ZFdpTD++jNRt4ba2vIG229T4H1Jt/ipTNREsKp08psguBUSmyMwqlj+yMqzjIzhghPEF2KZwCO2Oz5T7cbgjzb8aNMckI283ezsIM4JbGDQql486WP3vcOS6Zrs4/trs5ki5aPTQhcgrIXn1Sp4tWhwDSdGG/b+bzAjvncSTCrhrDYWyU4zd4zMXoCUpQC5gCLuLp7bJb/PNV6V6v77B3jT0aLx72rsnJ6fM39q4FE0OBrDLXWIfH7l+9b2jYO4Q0tgAAAABJRU5ErkJggg=="}`)))
			case 0x01:
				err = connection.WritePacket(data)
		}
		if err != nil {
			fmt.Println("Failed To Write Ping Packet! Error: ", err)
			return
		}
	}
}

func main() {
	fmt.Println("Starting Server...")
	listener := startListener()
	defer listener.Close()
	
	var err error
	PrivKey, err = rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		fmt.Println("Failed To Generate RSA Key! Error: ", err)
		os.Exit(1)
	}
	
	PrivKey.Precompute()
	
	for {
		connection, err := listener.Accept()
		if err != nil {
			fmt.Println("Failed To Recive Connection! Error: ", err)
			continue
		}
		go handleConnection(connection)
	}
}
