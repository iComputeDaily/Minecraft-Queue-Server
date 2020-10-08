package main

import "fmt"
import "errors"
import "strings"
import "encoding/json"
import "net/http"
import "io/ioutil"
import "crypto/rand"
import "crypto/rsa"
import "crypto/x509"
import "crypto/aes"
import "crypto/sha1"
import "github.com/google/uuid"
import mcnet "github.com/Tnze/go-mc/net"
import "github.com/Tnze/go-mc/net/packet"
import "github.com/Tnze/go-mc/net/CFB8"

// Player represents information about a connected player.
type Player struct {
	name string
	uuid uuid.UUID
	connection mcnet.Conn
	token [4]byte
}

// AuthResponse represents a response from the mojang auth server
type AuthResponse struct {
	Id string
	Name string
	Properties []struct {
		Name string
		Value string
		Signature string
	}
}

// handleLogin is called by handleConnection on any connections with handshake intention 2(login).
func (player *Player) handleLogin() {
	for packetNum := 0; packetNum < 3; packetNum++ {
		data, err := player.connection.ReadPacket()
		if err != nil {
			fmt.Println("Failed To Read Login Packet! Number: ", packetNum, "Error:", err)
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
				
				fmt.Println("sharedSecret: ", sharedSecret)
				
				err = player.authUser(sharedSecret)
				if err != nil {
					fmt.Println("Failed To Authenticate User! Error: ", err)
					return
				}
				
//				err = player.connection.WritePacket(packet.Marshal(0x03, packet.VarInt(-1)))
//				if err != nil {
//					fmt.Println("Failed To Send Set Compression Packet! Error: ", err)
//				}
				
				err = player.sendLoginSucsess()
				if err != nil {
					fmt.Println("Failed To Send Login Sucsess Packet! Error: ", err)
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
	hash, err := authDigest("", sharedSecret)
	if err != nil {
		return err
	}
	
	fmt.Println("Hash:", hash)
	
	resp, err := http.Get(fmt.Sprintf("https://sessionserver.mojang.com/session/minecraft/hasJoined?username=%s&serverId=%s", player.name, hash))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return errors.New(fmt.Sprintln("Mojang Auth Server Responded With An Error! Error:", resp.Status))
	}
	
	body, err := ioutil.ReadAll(resp.Body)
	
	var response AuthResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return err
	}
	
	fmt.Printf("response: %+v\n\nbody: %s\n", response, body)
	
	player.name = response.Name
	player.uuid, err = uuid.ParseBytes([]byte(response.Id))
	if err != nil {
		return err
	}
	
	cipher, err := aes.NewCipher(sharedSecret[:])
	if err != nil {
		return err
	}
	decoStream := CFB8.NewCFB8Decrypt(cipher, sharedSecret[:])
	encoStream := CFB8.NewCFB8Encrypt(cipher, sharedSecret[:])
	
	player.connection.SetCipher(encoStream, decoStream)
	
	return nil
}

func authDigest(serverID string, sharedSecret [16]byte) (string, error) {
	derPubKey, err := x509.MarshalPKIXPublicKey(PrivKey.Public())
	if err != nil {
		return "", errors.New(fmt.Sprintln("Failed To Encode RSA Key To DER Format! Error: ", err))
	}
	
	h := sha1.New()
	h.Write([]byte(serverID))
	h.Write(sharedSecret[:])
	h.Write(derPubKey)
	hash := h.Sum(nil)
	
	// Check for negative hashes
	negative := (hash[0] & 0x80) == 0x80
	if negative {
		hash = twosComplement(hash)
	}
	
	// Trim away zeroes
	res := strings.TrimLeft(fmt.Sprintf("%x", hash), "0")
	if negative {
		res = "-" + res
	}
	
	return res, nil
}

// little endian
func twosComplement(p []byte) []byte {
	carry := true
	for i := len(p) - 1; i >= 0; i-- {
		p[i] = byte(^p[i])
		if carry {
			carry = p[i] == 0xff
			p[i]++
		}
	}
	return p
}

func (player *Player) sendLoginSucsess() error {
	fmt.Println("name: ", player.name, "uuid: ", player.uuid.String())
	
	err := player.connection.WritePacket(packet.Marshal(0x02, packet.String(player.uuid.String()), packet.String(player.name)))
	if err != nil {
		return err
	}
	
	return nil
}
