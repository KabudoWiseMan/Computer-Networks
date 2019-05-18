package main

import (
	"net/smtp"
    "github.com/mgutz/logxi/v1"
    "bufio"
    "os"
    "fmt"
    "crypto/tls"
    "strings"
)

type SmtpServer struct {
	host string
	port string
}

func (s *SmtpServer) ServerName() string {
	return s.host + ":" + s.port
}

func main() {
	var toIds []string
	for i := 2; i <= len(os.Args) - 1; i++ {
		toIds = append(toIds, os.Args[i])
	}
	from := "From: vsevolodmolchanov@gmail.com" + "\r\n"
	to := "To: " + strings.Join(toIds, ";") + "\r\n"
	subject := "Subject: " + os.Args[1] + "\r\n"

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Enter your text:")
    message := ""
    for {
    	scanner.Scan()
        line := scanner.Text()
        if line == "EOF" {
            break
        }
        message += "\n" + line
    }

    smtpServer := SmtpServer{host: "smtp.gmail.com", port: "465"}
    auth := smtp.PlainAuth("", "vsevolodmolchanov@gmail.com", "************", smtpServer.host)
    tlsconfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         smtpServer.host,
	}
	conn, err := tls.Dial("tcp", smtpServer.ServerName(), tlsconfig)
	if err != nil {
		log.Error("Couldn't connect to smtp server", "error", err)
	}
	client, err := smtp.NewClient(conn, smtpServer.host)
	if err != nil {
		log.Error("Couldn't create new client", "error", err)
	}
	if err = client.Auth(auth); err != nil {
		log.Error("Invalid login or password", "error", err)
	}

	if err = client.Mail("vsevolodmolchanov@gmail.com"); err != nil {
		log.Error("Couldn't issue a MAIL command", "error", err)
	}
	for i := range toIds {
		if err = client.Rcpt(toIds[i]); err != nil {
			log.Error("Couldn't issue RCPT command for " + toIds[i], "error", err)
		}
	}
	w, err := client.Data()
	if err != nil {
		log.Error("Couldn't issue DATA command", "error", err)
	}

	_, err = w.Write([]byte(from + to + subject + message))
	if err != nil {
		log.Error("Couldn't write the message", "error", err)
	}

	err = w.Close()
	if err != nil {
		log.Error("Couldn't close the writer", "error", err)
	}

	client.Quit()

}