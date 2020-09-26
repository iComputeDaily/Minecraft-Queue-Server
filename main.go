package main

import "fmt"
import "os"
import "github.com/Tnze/go-mc/net"
import "github.com/Tnze/go-mc/net/packet"
import "crypto/rand"
import "crypto/rsa"
import "crypto/x509"
import "errors"

var privKey *rsa.PrivateKey

// Player represents information about a connected player.
type Player struct {
	name string
	uuid packet.UUID
	connection net.Conn
}

// startListener is called once to start listening for connections.
func startListener() *net.Listener {
	listener, err := net.ListenMC(":25565")
	if err != nil {
		fmt.Println("Could Not Start Server. Aborting! Error: ", err)
		os.Exit(1)
	}
	return listener
}

// handleConnection is called on al incomming connections.
func handleConnection(connection net.Conn) {
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
			handleLogin(connection)
	}
	
}

// handlePing is called by handleConnection on any connections with handshake intention 1(status).
func handlePing(connection net.Conn) {
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

// handleLogin is called by handleConnection on any connections with handshake intention 2(login).
func handleLogin(connection net.Conn) {
	var token [4]byte
	
	for packetNum := 0; packetNum < 2; packetNum++ {
		data, err := connection.ReadPacket()
		if err != nil {
			fmt.Println("Failed To Read Login Packet! Error:", err)
			return
		}
		
		switch data.ID {
			default:
				fmt.Println("Invalid Login Packet Id! ID: ", data.ID)
			case 0x00:
				player, err := handleLoginStart(data)
				if err != nil {
					fmt.Println("Failed To Parse Login Packet! Error: ", err)
				}
				
				fmt.Println("Player Name:", player.name, "Requested join")
				
				_, err = rand.Read(token[:])
				if err != nil {
					fmt.Println("Failed To Read Random Data For Login Token! Error: ", err)
					return
				}
				sendEncRequest(token, connection)
				
			case 0x01:
				sharedSecret, err := handleEncResponse(data, token)
				if err != nil {
					fmt.Println("Failed To Prosses Encryption Response! Error: ", err)
					return
				}
				
				// BUG(iComputeDaily): Might need to send raw public key data not shure
				err = authUser(sharedSecret)
				if err != nil {
					fmt.Println("Failed To Authenticate User! Error: ", err)
					return
				}
		}
	}
}

// handleLoginStart is called by handlelogin to process the login start packet
func handleLoginStart(data packet.Packet) (Player, error) {
	var player Player
	err := data.Scan((*packet.String)(&player.name))
	if err != nil {
		return player, err
	}
	
	return player, nil
}

// sendEncRequestRequest is called by handlelogin to send the encryption request packet
func sendEncRequest(token [4]byte, connection net.Conn) error {
	derPubKey, err := x509.MarshalPKIXPublicKey(privKey.Public())
	if err != nil {
		fmt.Println("Failed To Encode RSA Key To DER Format! Error: ", err)
		return errors.New("1")
	}
	
	err = connection.WritePacket(packet.Marshal(0x01, packet.String(""), packet.ByteArray(derPubKey), packet.ByteArray(token[:])))
	if err != nil {
		fmt.Println("Failed To Write Encryption Rquest Packet! Error: ", err)
		return errors.New("1")
	}
	
	return nil
}

// handleEncResponse is called by handlelogin to prosses the encryption response packet
func handleEncResponse(data packet.Packet, token [4]byte) ([16]byte, error) {
	var (
		encSharedSecret, encVerifyToken packet.ByteArray
	)
	
	err := data.Scan(&encSharedSecret, &encVerifyToken)
	if err != nil {
		fmt.Println("Failed To Parse Encryption Response Packet! Error: ", err)
		return [16]byte{0}, errors.New("1")
	}
	fmt.Println(encSharedSecret, encVerifyToken)
	
	var decVerifyToken [4]byte
	err = rsa.DecryptPKCS1v15SessionKey(rand.Reader, privKey, encVerifyToken, decVerifyToken[:])
	if err != nil {
		fmt.Println("Failed To Decrypt Verify Token! Error: ", err)
		return [16]byte{0}, errors.New("1")
	}
	if token != decVerifyToken {
		fmt.Println("Verify Tokens Do not Match!")
		return [16]byte{0}, errors.New("1")
	}
	
	var decSharedSecret [16]byte
	err = rsa.DecryptPKCS1v15SessionKey(rand.Reader, privKey, encSharedSecret, decSharedSecret[:])
	if err != nil {
		fmt.Println("Failed To Decrypt Shared Secret! Error: ", err)
		return [16]byte{0}, errors.New("1")
	}
	
	return decSharedSecret, nil
}

// authUser is called by handleLogin to authentication the user with mojang
func authUser(sharedSecret [16]byte) error {
	derPubKey, err := x509.MarshalPKIXPublicKey(privKey.Public())
	if err != nil {
		fmt.Println("Failed To Encode RSA Key To DER Format! Error: ", err)
		return errors.New("1")
	}
	
	fmt.Println("PubKey:", derPubKey)
	
	return nil
}

func main() {
	fmt.Println("Starting Server...")
	listener := startListener()
	defer listener.Close()
	
	var err error
	privKey, err = rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		fmt.Println("Failed To Generate RSA Key! Error: ", err)
		os.Exit(1)
	}
	
	privKey.Precompute()
	
	for {
		connection, err := listener.Accept()
		if err != nil {
			fmt.Println("Failed To Recive Connection! Error: ", err)
			continue
		}
		go handleConnection(connection)
	}
}
