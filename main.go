package main

import (
	"bufio"
	"fmt"
	"i2chat/sam"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

var wg sync.WaitGroup

var (
	identity sam.Identity
	samAddr  string = "127.0.0.1:7656"
	session  sam.SAMSession
)

func main() {
	conn, err := net.DialTimeout("tcp", samAddr, 3*time.Second)
	if err != nil {
		fmt.Println("Failed to connect to SAM bridge: ", err)
		return
	}
	fmt.Println("Connected to SAM bridge at: ", samAddr)
	defer conn.Close()

	sam.Hello(conn)

	cwd, err := os.Getwd()
	if err != nil {
		return
	}
	identityPath := filepath.Join(cwd, "storage", "users", "identity.json")
	if _, err = os.Stat(identityPath); err != nil {
		fmt.Println("No saved identity found, creating new one")
		identity, err = sam.CreateDestination(conn)
		if err != nil {
			fmt.Println("Error occured creating identity: ", err)
			return
		}
		err = sam.SaveIdentity(identity)
		if err != nil {
			fmt.Println("Error occurred saving identity: ", err)
			return
		}
	} else {
		identity, err = sam.LoadIdentity()
		if err != nil {
			fmt.Println("An error occurred while trying to load existing identity: ", err)
			return
		}
		fmt.Println("Using existing identity:")
		fmt.Println("Public Destination: ", identity.PubDest)
	}

	pubDestShort := identity.PubDest
	if len(pubDestShort) > 50 {
		pubDestShort = pubDestShort[:50] + "..."
	}

	scanner := bufio.NewScanner(os.Stdin)

	fmt.Printf("Enter the session id to connect to: ")
	scanner.Scan()
	sessionId := strings.TrimSpace(scanner.Text())
	session := sam.SAMSession{
		ID:       sessionId,
		Identity: identity,
		Conn:     conn,
	}
	err = sam.CreateStreamSession(&session)
	if err != nil {
		fmt.Println("Unable to crease SAM session: ", err)
		return
	}
	fmt.Println("Session successfully created!")

	operations := []string{
		"1. ACCEPT INCOMING STREAMS",
		"2. CONNECT TO A STREAM",
	}

	fmt.Printf("Choose from the following operations:\n")
	for _, operation := range operations {
		fmt.Println(operation)
	}

	scanner = bufio.NewScanner(os.Stdin)
	scanner.Scan()
	answer := strings.TrimSpace(scanner.Text())
	ans_int, err := strconv.Atoi(answer)
	if err != nil {
		fmt.Println("Enter a valid number")
		return
	}

	switch ans_int {
	case 1:
		errChan := make(chan error)
		wg.Add(1)
		fmt.Printf("Enter the stream id to accept: ")
		scanner.Scan()
		id := strings.TrimSpace(scanner.Text())
		go func() {
			err := sam.AcceptIncomingStream(conn, id, &wg)
			errChan <- err
		}()
		err := <-errChan
		if err != nil {
			fmt.Println("Error accepting stream: ", err)
			return
		}
		fmt.Println("Successfully accepted incoming stream!")
		break
	case 2:
		errChan := make(chan error)
		wg.Add(1)
		fmt.Printf("Enter the stream id to connect to: ")
		scanner.Scan()
		id := strings.TrimSpace(scanner.Text())
		fmt.Printf("Enter the destination to connect to: ")
		scanner.Scan()
		dest := strings.TrimSpace(scanner.Text())
		go func() {
			err := sam.ConnectToExternalStream(conn, id, dest, &wg)
			errChan <- err
		}()
		err := <-errChan
		if err != nil {
			fmt.Println("Error connecting to external stream: ", err)
			return
		}
		fmt.Println("Successfully connected to external stream!")
		break
	default:
		fmt.Printf("%d is an invalid choice\n", ans_int)
		break
	}
}
