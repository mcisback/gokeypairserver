package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

const (
	SERVER_HOST = "localhost"
	SERVER_PORT = "9988"
	SERVER_TYPE = "tcp"
)

type KeyPairDb map[string]string

type Client struct {
	Connection net.Conn
	Db         KeyPairDb
}

type ClientsMap map[string]Client

func main() {

	clientsMap := make(ClientsMap)

	fmt.Println("Server Running...")
	server, err := net.Listen(SERVER_TYPE, SERVER_HOST+":"+SERVER_PORT)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	defer server.Close()
	fmt.Println("Listening on " + SERVER_HOST + ":" + SERVER_PORT)
	fmt.Println("Waiting for client...")
	for {
		connection, err := server.Accept()

		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			// os.Exit(1)
		}

		remoteAddr := connection.RemoteAddr().String()

		fmt.Println("Client connected: ", remoteAddr)

		go processClient(&clientsMap, connection)
	}
}
func processClient(clientsMap *ClientsMap, connection net.Conn) {
	remoteAddr := connection.RemoteAddr().String()

	client := Client{
		Connection: connection,
		Db:         make(KeyPairDb),
	}

	(*clientsMap)[remoteAddr] = client

	fmt.Println("ClientsMap: ", *clientsMap)

	var buffer string

	fmt.Println("------------- START OF RECEIVING -------------")

	for {
		_, received, err := receive(connection)

		if err != nil {
			break
		}

		if received != "\n" {
			buffer += received
		} else {
			fmt.Println("------------- END OF RECEIVING -------------")
		}

		fmt.Println("totalBuffer, len: ", buffer, len(buffer))

		// What if value of key contains end ?
		if received == "\n" && buffer != "" {

			fmt.Println("totalBuffer, len: ", buffer, len(buffer))

			// FIXME: should buffer == "ecc..."
			if buffer == "QUIT" {
				fmt.Println("Closing connection, bye bye")

				connection.Close()

				break
			}

			if strings.HasPrefix(buffer, "set") {
				fmt.Println("Received SET")

				parts := strings.Split(buffer, " ")

				fmt.Println("Parts:  ", parts)

				key := parts[1]
				value := strings.Join(parts[2:], " ")

				client := getCurrentClient(clientsMap, connection)

				client.Db[key] = strings.TrimSuffix(value, "\n")

				fmt.Println("Client DB: ", client.Db)

				send(connection, fmt.Sprintf("%s = %s setted", key, value))
			} else if strings.HasPrefix(buffer, "get") {
				parts := strings.Split(buffer, " ")

				key := parts[1]

				client := getCurrentClient(clientsMap, connection)

				send(connection, client.Db[key])
			} else if strings.HasPrefix(buffer, "del") {
				parts := strings.Split(buffer, " ")

				key := parts[1]

				client := getCurrentClient(clientsMap, connection)

				delete(client.Db, key)

				fmt.Println("Client DB: ", client.Db)

				send(connection, "Deleted")
			} else {
				send(connection, "Command Unknown")
			}

			buffer = ""
		}

	}

}

func getCurrentClient(clientsMap *ClientsMap, connection net.Conn) Client {
	return (*clientsMap)[connection.RemoteAddr().String()]
}

func receive(connection net.Conn) (int, string, error) {
	remoteAddr := connection.RemoteAddr().String()

	// var totalBuffer []byte
	// var totalLen int
	// var received string
	// var buffer []byte

	buffer := make([]byte, 4)

	mLen, err := connection.Read(buffer)

	fmt.Println("receive() -> Buffer, mLen: ", string(buffer), mLen)

	if err != nil {
		fmt.Println("Error reading:", err.Error())

		fmt.Println("Closing connection, bye bye")

		connection.Close()

		return mLen, "", err
	}

	received := string(buffer[:mLen])

	// Stripping \n could create problems ?
	// received = strings.TrimSuffix(received, "\n")

	fmt.Printf("Received[%s]:\"%s\"\n", remoteAddr, received)

	return mLen, received, nil
}

func send(connection net.Conn, message string) {
	_, err := connection.Write([]byte(message + "\n"))

	if err != nil {
		fmt.Println("Error sending data: ", err.Error())

		connection.Close()
	}
}
