package commands

import "fmt"

type Command string

const (
	cmdHello               Command = "HELLO VERSION MIN=3.1 MAX=3.2\n"
	cmdGenerateDestination Command = "DEST GENERATE SIGNATURE_TYPE=7\n"
	cmdCreateSession       Command = "SESSION CREATE STYLE=STREAM ID=%s DESTINATION=%s SIGNATURE_TYPE=7\n"
	cmdAcceptStream        Command = "STREAM ACCEPT ID=%s\n"
	cmdConnectToStream     Command = "STREAM CONNECT ID=%s DESTINATION=%s\n"
)

func Hello() string {
	return string(cmdHello)
}

func GenerateDestination() string {
	return string(cmdGenerateDestination)
}

func CreateSession(id, dest string) string {
	return fmt.Sprintf(string(cmdCreateSession), id, dest)
}

func AcceptStream(id string) string {
	return fmt.Sprintf(string(cmdAcceptStream), id)
}

func ConnectToStream(id, dest string) string {
	return fmt.Sprintf(string(cmdConnectToStream), id, dest)
}
