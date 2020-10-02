package main

import "fmt"
import "errors"
import "crypto/rand"
import "crypto/rsa"
import "crypto/x509"
import "crypto/cipher"
import "github.com/Tnze/go-mc/net"
import "github.com/Tnze/go-mc/net/packet"

// Player represents information about a connected player.
type Player struct {
	name string
	uuid packet.UUID
	connection net.Conn
	token [4]byte
	key [16]byte
	encoStream, decoStream cipher.Stream
}

// handleLogin is called by handleConnection on any connections with handshake intention 2(login).
func (player *Player) handleLogin() {
	for packetNum := 0; packetNum < 2; packetNum++ {
		fmt.Println("Handlelogin loop was entered")
		
		data, err := player.connection.ReadPacket()
		if err != nil {
			fmt.Println("Failed To Read Login Packet! Error:", err)
			return
		}
		
		switch data.ID {
			default:
				fmt.Println("Invalid Login Packet Id! ID: ", data.ID)
				
			case 0x00:
				err := player.handleLoginStart(data)
				if err != nil {
					fmt.Println("Failed To Parse Login Packet! Error: ", err)
				}
				
				fmt.Println("Player Name:", player.name, "Requested join")
				
				_, err = rand.Read(player.token[:])
				if err != nil {
					fmt.Println("Failed To Read Random Data For Login Token! Error: ", err)
					return
				}
				err = player.sendEncRequest()
				if err != nil {
					fmt.Println("Failed To Send Encryption Request! Error: ", err)
					return
				}
				
			case 0x01:
				sharedSecret, err := player.handleEncResponse(data)
				if err != nil {
					fmt.Println("Failed To Prosses Encryption Response! Error: ", err)
					return
				}
				
				// BUG(iComputeDaily): Might need to send raw public key data not shure
				err = player.authUser(sharedSecret)
				if err != nil {
					fmt.Println("Failed To Authenticate User! Error: ", err)
					return
				}
		}
	}
}

// handleLoginStart is called by handlelogin to process the login start packet
func (player *Player) handleLoginStart(data packet.Packet) (error) {
	err := data.Scan((*packet.String)(&player.name))
	if err != nil {
		return err
	}
	
	return nil
}

// sendEncRequestRequest is called by handlelogin to send the encryption request packet
func (player *Player) sendEncRequest() error {
	derPubKey, err := x509.MarshalPKIXPublicKey(PrivKey.Public())
	if err != nil {
		return errors.New(fmt.Sprintln("Failed To Encode RSA Key To DER Format! Error: ", err))
	}
	
	err = player.connection.WritePacket(packet.Marshal(0x01, packet.String(""), packet.ByteArray(derPubKey), packet.ByteArray(player.token[:])))
	if err != nil {
		return errors.New(fmt.Sprintln("Failed To Write Encryption Rquest Packet! Error: ", err))
	}
	
	return nil
}

// handleEncResponse is called by handlelogin to prosses the encryption response packet
func (player *Player) handleEncResponse(data packet.Packet) ([16]byte, error) {
	var (
		encSharedSecret, encVerifyToken packet.ByteArray
	)
	
	err := data.Scan(&encSharedSecret, &encVerifyToken)
	if err != nil {
		return [16]byte{0}, errors.New(fmt.Sprintln("Failed To Parse Encryption Response Packet! Error: ", err))
	}
	fmt.Println("encSharedSecret: ", encSharedSecret, "encVerifyToken: ", encVerifyToken)
	
	var decVerifyToken [4]byte
	err = rsa.DecryptPKCS1v15SessionKey(rand.Reader, PrivKey, encVerifyToken, decVerifyToken[:])
	if err != nil {
		fmt.Println("Failed To Decrypt Verify Token! Error: ", err)
		return [16]byte{0}, errors.New("1")
	}
	if player.token != decVerifyToken {
		return [16]byte{0}, errors.New(fmt.Sprintln("Verify Tokens Do not Match!"))
	}
	
	var decSharedSecret [16]byte
	err = rsa.DecryptPKCS1v15SessionKey(rand.Reader, PrivKey, encSharedSecret, decSharedSecret[:])
	if err != nil {
		fmt.Println()
		return [16]byte{0}, errors.New(fmt.Sprintln("Failed To Decrypt Shared Secret! Error: ", err))
	}
	
	return decSharedSecret, nil
}

// authUser is called by handleLogin to authentication the user with mojang
func (player *Player) authUser(sharedSecret [16]byte) error {
	derPubKey, err := x509.MarshalPKIXPublicKey(PrivKey.Public())
	if err != nil {
		return errors.New(fmt.Sprintln("Failed To Encode RSA Key To DER Format! Error: ", err))
	}
	
	fmt.Println("PubKey:", derPubKey)
	
	return nil
}
