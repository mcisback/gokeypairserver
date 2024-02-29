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

	for {
		_, received, _ := receive(connection)

		if received == "QUIT" {
			fmt.Println("Closing connection, bye bye")

			connection.Close()

			break
		}

		if strings.HasPrefix(received, "SET") {
			fmt.Println("Received SET")

			parts := strings.Split(received, " ")

			key := parts[1]
			value := parts[2]

			client := getCurrentClient(clientsMap, connection)

			client.Db[key] = value

			fmt.Println("Client DB: ", client.Db)

			send(connection, "Thanks! Got your message:"+received)
		} else if strings.HasPrefix(received, "GET") {
			parts := strings.Split(received, " ")

			key := parts[1]

			client := getCurrentClient(clientsMap, connection)

			send(connection, client.Db[key])
		}

	}

}

func getCurrentClient(clientsMap *ClientsMap, connection net.Conn) Client {
	return (*clientsMap)[connection.RemoteAddr().String()]
}

func receive(connection net.Conn) (int, string, error) {
	buffer := make([]byte, 1024)

	mLen, err := connection.Read(buffer)

	if err != nil {
		fmt.Println("Error reading:", err.Error())

		fmt.Println("Closing connection, bye bye")

		connection.Close()

		return mLen, "", err
	}

	received := string(buffer[:mLen])
	received = strings.TrimSuffix(received, "\n")

	fmt.Printf("Received:\"%s\"\n", received)

	return mLen, received, nil
}

func send(connection net.Conn, message string) {
	_, err := connection.Write([]byte(message + "\n"))

	if err != nil {
		fmt.Println("Error sending data: ", err.Error())

		connection.Close()
	}
}
