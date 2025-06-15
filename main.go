package main

import (
	"bufio"
	"fmt"
	"i2chat/internal/toolset"
	"i2chat/sam"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var wg sync.WaitGroup

var (
	identity sam.Identity
	samAddr  string = "127.0.0.1:7656"
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

	if _, err = os.Stat("./storage/users/identity.json"); err != nil {
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

	operations := []string{
		"1. CREATE SESSION",
		"2. CONVERT DESTIANTION TO PUBLIC ADDRESS",
		"3. ACCEPT INCOMING STREAMS",
		"4. CONNECT TO A STREAM",
	}

	fmt.Printf("Choose from the following operations:\n")
	for _, operation := range operations {
		fmt.Println(operation)
	}

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	answer := strings.TrimSpace(scanner.Text())
	ans_int, err := strconv.Atoi(answer)
	if err != nil {
		fmt.Println("Enter a valid number")
		return
	}

	switch ans_int {
	case 1:
		err = sam.CreateStreamSession(conn, identity)
		if err != nil {
			fmt.Println("Unable to crease SAM session: ", err)
			return
		}
		fmt.Println("Session successfully created!")
		break
	case 2:
		var pubDest string = strings.TrimRight(identity.PubDest, "=")
		fmt.Println("pubDest: ", pubDest)
		publicAddress, err := toolset.ConvertPubDestToB32(pubDest)
		if err != nil {
			fmt.Println("Error converting destination into public address: ", err)
			return
		}
		fmt.Printf("Your public address is: %s", publicAddress)
		break

	case 3:
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
	case 4:
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
