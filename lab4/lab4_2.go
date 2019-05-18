package main

import (
    "github.com/gliderlabs/ssh"
    "github.com/mgutz/logxi/v1"
    "strings"
    "golang.org/x/crypto/ssh/terminal"
    "os/exec"
)

var (
    users map[string]string
)

func HomeRouterHandler(s ssh.Session) {
    flag := false
    term := terminal.NewTerminal(s, "$ ")
    for {
        command := ""
        name := ""
        args := []string{}
        if len(s.Command()) > 0 {
            command = strings.Join(s.Command(), " ")
            name = s.Command()[0]
            args = s.Command()[1:]
            flag = true
        } else {
            command, _ = term.ReadLine()
            name = strings.Split(command, " ")[0]
            args = strings.Split(command, " ")[1:]
        }
        if command == "exit" {
            break
        }
        log.Info("User says: ", command)
        cmd := exec.Command(name, args...)
        response, err := cmd.CombinedOutput()
        if err != nil {
            log.Error("Couldn't execute the command " + command, err)
        }
        if string(response) != "" {
            term.Write(response)
        }
        if flag {
            break
        }
    }
    log.Info("terminal closed")
}

func Auth(ctx ssh.Context, pass string) bool {
    if p, ok := users[ctx.User()]; ok && p == pass {
        return true
    }
    return false
}

func main() {
    users = make(map[string]string)
    users["user"] = "password"
    ssh.Handle(HomeRouterHandler) // установим роутер
    err := ssh.ListenAndServe(":443", nil, ssh.HostKeyFile("keys/id_rsa"), ssh.PasswordAuth(Auth)) // задаем слушать порт
    if err != nil {
        log.Fatal("ListenAndServe: ", err)
    }
}