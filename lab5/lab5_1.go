package main

import (
    "github.com/jlaffaye/ftp"
    "github.com/mgutz/logxi/v1"
    "path"
    "os"
    "io/ioutil"
)

func main() {
	connection, err := ftp.Dial("127.*.*.***:21")
	if err != nil {
		log.Error("Couldn't connect to server", "error", err)
	}
	if err := connection.Login("ft****", "************"); err != nil {
		log.Error("Invalid login or password", "error", err)
	}

	entries, err := connection.List(".")
	if err != nil {
		log.Error("Couldn't get files", "error", err)
	} else {
		for i := range entries {
			name := entries[i].Name
			if path.Ext(name) == ".txt" {
				reader, err := connection.Retr(name)
				if err != nil {
					log.Error("Couldn't get file " + name, "error", err)
				}
				buf, err := ioutil.ReadAll(reader)
				if err != nil {
					log.Error("Couldn't read file " + name, "error", err)	
				}

				file, err := os.Create(name)
				if err != nil {
					log.Error("Couldn't create file " + name, "error", err)
				}

				if _, err := file.Write(buf); err != nil {
					log.Error("Couldn't write to file " + name, "error", err)	
				}
				if err := file.Close(); err != nil {
					log.Error("Couldn't close file " + name, "error", err)
				}

				break
			}
		}
	}

	if err := connection.Quit(); err != nil {
		log.Error("Couldn't quit the server", "error", err)
	}
}