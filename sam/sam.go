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
	"time"
)

const samAddr = "127.0.0.1:7656"

type Identity struct {
	PubDest string `json:"pubDest"`
	PrivKey string `json:"privKey"`
}

func hello(conn net.Conn) error {
	_, err := conn.Write([]byte("HELLO VERSION MIN=3.1 MAX=3.1\n"))
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

func connectToSAM() (net.Conn, *bufio.Reader, error) {
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

func CreateDestination() (Identity, error) {
	conn, reader, err := connectToSAM()
	if err != nil {
		return Identity{}, err
	}
	defer conn.Close()

	_, err = conn.Write([]byte("DEST GENERATE SIGNATURE_TYPE=7\n"))
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

func CreateStreamSession(sessionID string, privKey string) error {
	conn, reader, err := connectToSAM()
	if err != nil {
		return err
	}
	defer conn.Close()

	cmd := fmt.Sprintf("SESSION CREATE STYLE=STREAM ID=%s DESTINATION=%s SIGNATURE_TYPE=7\n", sessionID, privKey)
	_, err = conn.Write([]byte(cmd))
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

func AcceptStream(sessionID string) error {
	conn, reader, err := connectToSAM()
	if err != nil {
		return err
	}
	defer conn.Close()

	cmd := fmt.Sprintf("STREAM ACCEPT ID=%s\n", sessionID)
	_, err = conn.Write([]byte(cmd))
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

func ConnectToStream(sessionID, dest string) error {
	conn, reader, err := connectToSAM()
	if err != nil {
		return err
	}
	defer conn.Close()

	cmd := fmt.Sprintf("STREAM CONNECT ID=%s DESTINATION=%s\n", sessionID, dest)
	_, err = conn.Write([]byte(cmd))
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
