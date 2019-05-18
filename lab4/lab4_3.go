package main

import (
 	"fmt"
 	"golang.org/x/crypto/ssh"
 	"github.com/mgutz/logxi/v1"
 	"os"
 	"strings"
 	"bufio"
 	"time"
 	"golang.org/x/crypto/ssh/terminal"
)

func getPass(i int) string {
	fmt.Print(os.Args[i] + "'s password: ")
	password, err := terminal.ReadPassword(0)
	fmt.Print("\n")
	if err != nil {
		log.Error("This is not a password", err)
	}
	return string(password)
}

type Server struct {
	User, Password, Host, Port string
}

func executeCommand(command string, config *ssh.ClientConfig, server Server) {
	log.Info("Connecting to server " + server.User + "@" + server.Host)
	client, err := ssh.Dial("tcp", server.Host + ":" + server.Port, config)
	if err != nil {
		log.Error("Couldn't connect to server " + server.User + "@" + server.Host, err)
	}

	log.Info("Executing command on " + command + "on " + server.User + "@" + server.Host)
	session, err := client.NewSession()
	if err != nil {
		log.Error("Couldn't make new session on " + server.User + "@" + server.Host, err)
	}
	defer session.Close()
	response, err := session.CombinedOutput(command)
	if err != nil {
		log.Error("Couldn't execute the command " + command + "on " + server.User + "@" + server.Host, err)
	}
	if string(response) != "" {
		fmt.Println(server.User + "@" + server.Host + " says:" + "\n" + string(response))
	}
}

func main() {
	command := ""
	flag := false
	var servers []Server
	for i := 1; i <= len(os.Args) - 1; i++ {
		s := strings.Split(os.Args[i], "@")
		if s[0] == os.Args[i] {
			command = strings.Join(os.Args[i:], " ")
			flag = true
			break
		}
		if i + 2 <= len(os.Args) - 1 {
			if os.Args[i + 1] == "-p" {
				server := Server{
					User: s[0],
					Password: getPass(i),
					Host: s[1],
					Port: os.Args[i + 2],
				}
				servers = append(servers, server)
				i += 2
				continue
			}
		}
		server := Server{
			User: s[0],
			Password: getPass(i),
			Host: s[1],
			Port: "22",
		}
		servers = append(servers, server)
	}
	var configs []*ssh.ClientConfig
	for i := range servers {
		config := &ssh.ClientConfig{
			User: servers[i].User,
			Auth: []ssh.AuthMethod{
				ssh.Password(servers[i].Password),
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}
		configs = append(configs, config)
	}

	for {
		if !flag {
			fmt.Print("Command: ")
			in := bufio.NewScanner(os.Stdin)
			in.Scan()
			command = in.Text()
		}
		if command == "exit" {
			fmt.Println("Connection to all servers closed.")
			break
		}

		for i := range configs {
			go executeCommand(command, configs[i], servers[i])
		}
		amt := time.Duration(300 * len(servers))
        time.Sleep(time.Millisecond * amt)
        if flag {
			break
		}
	}
}