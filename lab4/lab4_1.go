package main

import (
 	"fmt"
 	"golang.org/x/crypto/ssh"
 	"github.com/mgutz/logxi/v1"
 	"os"
 	"strings"
 	"bufio"
 	"golang.org/x/crypto/ssh/terminal"
)

func getPass() string {
	fmt.Print(os.Args[1] + "'s password: ")
	password, err := terminal.ReadPassword(0)
	fmt.Print("\n")
	if err != nil {
		log.Error("This is not a password", err)
	}
	return string(password)
}

func main() {
	s := strings.Split(os.Args[1], "@")
	config := &ssh.ClientConfig{
		User: s[0],
		Auth: []ssh.AuthMethod{
			ssh.Password(getPass()),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	port := "22"
	if len(os.Args) >= 4 && os.Args[2] == "-p" {
		port = os.Args[3]
	}
	log.Info("Connecting to server")
	client, err := ssh.Dial("tcp", s[1] + ":" + port, config)
	if err != nil {
		log.Error("Couldn't connect to server", err)
	}

	log.Info("Executing commands")
	flag := false
	for {
		command := ""
		session, err := client.NewSession()
		if err != nil {
			log.Error("Couldn't make new session", err)
		}
		defer session.Close()
		if len(os.Args) >= 3 && (len(os.Args) != 4 && (os.Args[2] == "-p" || port == "22")) {
			if port != "22" {
				command = strings.Join(os.Args[4:], " ")
			} else {
				command = strings.Join(os.Args[2:], " ")
			}
			flag = true
		} else {
			fmt.Print("Command: ")
			in := bufio.NewScanner(os.Stdin)
			in.Scan()
			command = in.Text()
		}
		if command == "exit" {
				fmt.Println("Connection to " + s[1] + " closed.")
				break
		}
		response, err := session.CombinedOutput(command)
		if err != nil {
			log.Error("Couldn't execute the command " + command, err)
		}
		if string(response) != "" {
			fmt.Println(string(response))
		}
		if flag {
			break
		}
	}
}