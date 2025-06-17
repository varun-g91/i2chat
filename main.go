package main

import (
	"bufio"
	"fmt"
	"i2chat/sam"
	"os"
	"path/filepath"
	"strconv"

	"github.com/google/uuid"
)

func main() {
	var id sam.Identity
	var err error

	conn, reader, err := sam.ConnectToSAM()
	defer conn.Close()

	identityPath := filepath.Join("storage", "users", "identity.json")
	if _, err = os.Stat(identityPath); os.IsNotExist(err) {
		id, err = sam.CreateDestination(conn, reader)
		if err != nil {
			fmt.Println("Failed to create identity:", err)
			return
		}
		if err := sam.SaveIdentity(id); err != nil {
			fmt.Println("Failed to save identity:", err)
			return
		}
	} else {
		id, err = sam.LoadIdentity()
		if err != nil {
			fmt.Println("Failed to load identity:", err)
			return
		}
	}

	sessionID := uuid.New()
	if err := sam.CreateStreamSession(conn, reader, sessionID.String(), id.PrivKey); err != nil {
		fmt.Println("Session create failed:", err)
		return
	}
	fmt.Println("Session created.")

	fmt.Println("Choose operation:")
	fmt.Println("1. Accept Stream")
	fmt.Println("2. Connect to Stream")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	choice, _ := strconv.Atoi(scanner.Text())

	switch choice {
	case 1:
		if err := sam.AcceptStream(conn, reader, sessionID.String()); err != nil {
			fmt.Println("Accept failed:", err)
		}
	case 2:
		fmt.Print("Enter destination to connect to: ")
		scanner.Scan()
		dest := scanner.Text()
		if err := sam.ConnectToStream(conn, reader, sessionID.String(), dest); err != nil {
			fmt.Println("Connect failed:", err)
		}
	default:
		fmt.Println("Invalid choice.")
	}
}
