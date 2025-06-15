package sam

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type Identity struct {
	PubDest string `json:"pubDest"`
	PrivKey string `json:"privKey"`
}

var identityCount int
var reader *bufio.Reader

func Hello(conn net.Conn) {
	hello := "HELLO VERSION MIN=3.1 MAX=3.1\n"
	_, err := conn.Write([]byte(hello))
	if err != nil {
		fmt.Println("Failed to perfom SAM handshake: ", err)
		return
	}
	reader = bufio.NewReader(conn)
	resp, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Failed to read resp: ", err)
		return
	}
	fmt.Println("SAM resp: ", resp)
}

func CreateDestination(conn net.Conn) (Identity, error) {
	cmd := "DEST GENERATE SIGNATURE_TYPE=7\n"
	_, err := conn.Write([]byte(cmd))
	if err != nil {
		fmt.Println("Failed to generate destination: ", err)
		return Identity{}, err
	}
	reader = bufio.NewReader(conn)
	resp, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Failed to read DEST GENERATE resp: ", err)
		return Identity{}, err
	}
	resp = strings.TrimSpace(resp)
	if !strings.HasPrefix(resp, "DEST REPLY") {
		fmt.Println("Unexpected error", resp)
		return Identity{}, err
	}
	parts := strings.Fields(resp)
	var pubKey, privKey string
	for _, part := range parts {
		if strings.HasPrefix(part, "PUB=") {
			pubKey = strings.TrimPrefix(part, "PUB=")
		}
		if strings.HasPrefix(part, "PRIV=") {
			privKey = strings.TrimPrefix(part, "PRIV=")
		}
	}
	fmt.Println("Public Destination: ", pubKey)
	fmt.Println("Private key: ", privKey)
	return Identity{PubDest: pubKey, PrivKey: privKey}, nil
}

func LoadIdentity() (Identity, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return Identity{}, err
	}
	identityPath := filepath.Join(cwd, "storage", "users", "identity.json")

	data, err := os.ReadFile(identityPath)
	if err != nil {
		return Identity{}, err
	}

	var identity Identity
	err = json.Unmarshal(data, &identity)
	if err != nil {
		return Identity{}, err
	}

	fmt.Println("Loaded existing identity from:", identityPath)
	return identity, nil
}

func SaveIdentity(identity Identity) error {
	data, err := json.MarshalIndent(identity, "", " ")
	if err != nil {
		fmt.Println("Error marshalling to JSON: ", err)
		return err
	}
	err = os.WriteFile("../storage/users/identity.json", data, 0644)
	if err != nil {
		fmt.Println("Error writing json to file: ", err)
		return err
	}
	fmt.Println("Identity successfuly saved to identity.json")
	return nil
}

func CreateStreamSession(conn net.Conn, identity Identity) error {
	// Fixed: Added space before SIGNATURE_TYPE
	cmd := fmt.Sprintf("SESSION CREATE STYLE=STREAM ID=default DESTINATION=%s SIGNATURE_TYPE=7\n", identity.PrivKey)

	_, err := conn.Write([]byte(cmd))
	if err != nil {
		fmt.Println("Error creating SAM session: ", err)
		return err
	}

	reader = bufio.NewReader(conn)
	resp, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error in createStreamSession while reading response: ", err)
		return err
	}

	resp = strings.TrimSpace(resp)

	if !strings.Contains(resp, "RESULT=OK") {
		fmt.Printf("Session creation failed with response: %s\n", resp)
		return errors.New("Failed to create session")
	} else {
		// Session created successfully, use the public destination from identity
		fmt.Printf("Session created successfully with public destination: %s\n", identity.PubDest)
		return nil
	}
}

func AcceptIncomingStream(conn net.Conn, id string, wg *sync.WaitGroup) error {
	defer wg.Done()
	cmd := "STREAM ACCEPT ID=" + id
	_, err := conn.Write([]byte(cmd))
	if err != nil {
		fmt.Printf("Could not accept incoming stream: %s\n", err)
		return err
	}
	reader = bufio.NewReader(conn)
	resp, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("Could not read response: %s\n", err)
		return err
	}
	fmt.Println("Incoming stream accepted, response: ", resp)
	return nil
}

func ConnectToExternalStream(conn net.Conn, id string, destination string, wg *sync.WaitGroup) error {
	defer wg.Done()
	cmd := "STREAM CONNECT ID=" + id + "DESTINATION=" + destination
	_, err := conn.Write([]byte(cmd))
	if err != nil {
		fmt.Println("Could not connect to external stream: ", err)
		return err
	}
	fmt.Println("Successfully connected to external stream of id: ", id)
	return nil
}
