package sam

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"i2chat/sam/commands"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const samAddr = "127.0.0.1:7656"

type Identity struct {
	PubDest string `json:"pubDest"`
	PrivKey string `json:"privKey"`
}

func hello(conn net.Conn) error {
	_, err := conn.Write([]byte(commands.Hello()))
	if err != nil {
		return err
	}
	reader := bufio.NewReader(conn)
	resp, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	if !strings.Contains(resp, "HELLO REPLY RESULT=OK") {
		return fmt.Errorf("HELLO failed: %s", resp)
	}
	return nil
}

func ConnectToSAM() (net.Conn, *bufio.Reader, error) {
	conn, err := net.DialTimeout("tcp", samAddr, 3*time.Second)
	if err != nil {
		return nil, nil, err
	}
	if err := hello(conn); err != nil {
		conn.Close()
		return nil, nil, err
	}
	reader := bufio.NewReader(conn)
	return conn, reader, nil
}

func CreateDestination(conn net.Conn, reader *bufio.Reader) (Identity, error) {
	_, err := conn.Write([]byte(commands.GenerateDestination()))
	if err != nil {
		return Identity{}, err
	}
	resp, err := reader.ReadString('\n')
	if err != nil {
		return Identity{}, err
	}
	fields := strings.Fields(resp)
	var pub, priv string
	for _, f := range fields {
		if strings.HasPrefix(f, "PUB=") {
			pub = strings.TrimPrefix(f, "PUB=")
		}
		if strings.HasPrefix(f, "PRIV=") {
			priv = strings.TrimPrefix(f, "PRIV=")
		}
	}
	if pub == "" || priv == "" {
		return Identity{}, errors.New("failed to extract keys")
	}
	return Identity{PubDest: pub, PrivKey: priv}, nil
}

func LoadIdentity() (Identity, error) {
	data, err := os.ReadFile(filepath.Join("storage", "users", "identity.json"))
	if err != nil {
		return Identity{}, err
	}
	var id Identity
	err = json.Unmarshal(data, &id)
	return id, err
}

func SaveIdentity(id Identity) error {
	data, err := json.MarshalIndent(id, "", "  ")
	if err != nil {
		return err
	}
	path := filepath.Join("storage", "users")
	os.MkdirAll(path, 0755)
	return os.WriteFile(filepath.Join(path, "identity.json"), data, 0644)
}

func CreateStreamSession(conn net.Conn, reader *bufio.Reader, sessionID string, privKey string) error {
	_, err := conn.Write([]byte(commands.CreateSession(sessionID, privKey)))
	if err != nil {
		return err
	}
	resp, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	if !strings.Contains(resp, "RESULT=OK") {
		return fmt.Errorf("session failed: %s", resp)
	}
	return nil
}

func AcceptStream(conn net.Conn, reader *bufio.Reader, sessionID string) error {
	_, err := conn.Write([]byte(commands.AcceptStream(sessionID)))
	if err != nil {
		return err
	}
	resp, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	fmt.Println("ACCEPTED: ", resp)
	return nil
}

func ConnectToStream(conn net.Conn, reader *bufio.Reader, sessionID, dest string) error {
	_, err := conn.Write([]byte(commands.ConnectToStream(sessionID, dest)))
	if err != nil {
		return err
	}
	resp, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	if !strings.Contains(resp, "RESULT=OK") {
		return fmt.Errorf("CONNECT failed: %s", resp)
	}
	fmt.Println("CONNECTED: ", resp)
	return nil
}
