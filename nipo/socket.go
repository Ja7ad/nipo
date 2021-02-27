package main

import (
	"net"
	"os"
	"fmt"
	"bufio"
	"strings"
	"encoding/json"
)

func Login(config *Config, connection net.Conn) bool {
	strRemoteAddr := connection.RemoteAddr().String()
	connection.Write([]byte("Welcome to NIPO"+"\n"))
	connection.Write([]byte("You are connecting from "+strRemoteAddr+"\n"))
	connection.Write([]byte("Enter username : "))
	username, err := bufio.NewReader(connection).ReadString('\n')
	authorized := false
	if err != nil {
		fmt.Println(err)
		return false
	}
	connection.Write([]byte("Enter password : "))
	password, err := bufio.NewReader(connection).ReadString('\n')
	if err != nil {
		fmt.Println(err)
		return false
	}
	username = strings.TrimSuffix(username, "\n")
	password = strings.TrimSuffix(password, "\n")
	for _, user := range config.Users {
		if username == user.Username {
			if password == user.Password {
				authorized = true
				return authorized
			} 
		} 
	}
	if authorized == false {
		connection.Write([]byte("nipo > wrong user or password"))
		connection.Write([]byte("\n"))
	}
	return authorized
}

func (database *Database) HandelSocket(config *Config, connection net.Conn) {
	defer connection.Close()
	strRemoteAddr := connection.RemoteAddr().String()
	for {
		connection.Write([]byte("nipo > "))
		input, err := bufio.NewReader(connection).ReadString('\n')
		if err != nil {
				fmt.Println(err)
				return
		}
		if strings.TrimSpace(string(input)) == "exit" {
				config.logger("Client closed the connection from "+strRemoteAddr)
				return
		}
		if strings.TrimSpace(string(input)) == "EOF" {
			config.logger("Client terminated the connection from "+strRemoteAddr)
			return
		}
		returneddb := database.cmd(string(input), config)
		jsondb, err := json.Marshal(returneddb.items)
		if err != nil {
			config.logger("Error in converting to json")
		}
		if len(jsondb) > 2 {
			connection.Write(jsondb)
			connection.Write([]byte("\n"))
		}
	}
}

func (database *Database) OpenSocket(config *Config) {
	config.logger("Opennig Socket on "+config.Listen.Ip+":"+config.Listen.Port+"/"+config.Listen.Protocol)
	socket,err := net.Listen(config.Listen.Protocol, config.Listen.Ip+":"+config.Listen.Port)
	if err != nil {
        config.logger("Error listening: "+err.Error())
		os.Exit(1)
	}
	defer socket.Close()
	for {
		connection, err := socket.Accept()
		strRemoteAddr := connection.RemoteAddr().String()
		if err != nil {
			config.logger("Error accepting socket: "+err.Error())
		}
		if Login(config,connection) {
			connection.Write([]byte("Here you go with NIPO"+"\n"))
			go database.HandelSocket(config, connection)
		} else {
			config.logger("Wrong user pass from "+strRemoteAddr)
			connection.Close()
		}
	}
}

func ConnectSocket(config *Config) {
	socket,err := net.Dial(config.Listen.Protocol, config.Listen.Ip+":"+config.Listen.Port)
	if err != nil {
        config.logger("Error Connecting: "+err.Error())
		os.Exit(1)
	}
	defer socket.Close()
	clientReader := bufio.NewReader(os.Stdin)
	serverReader := bufio.NewReader(socket)
 
	for {
		serverResponse, err := serverReader.ReadString('\n')
		if err != nil {
			config.logger("Server did not responsed to request")
		}
		socket.Write([]byte(serverResponse + "\n"))
		clientRequest, err := clientReader.ReadString('\n')
		if err != nil {
			config.logger("Client request was not correct")
		}
		socket.Write([]byte(clientRequest + "\n"))
	}
}