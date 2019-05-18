package main
import (
    "net"
    "os"
    "github.com/RealJK/rss-parser-go"
)

func main() {
    // strEcho := "test"
    // servAddr := "lab.posevin.com:10049"
    // tcpAddr, err := net.ResolveTCPAddr("tcp", servAddr)
    // if err != nil {
    //     println("ResolveTCPAddr failed:", err.Error())
    //     os.Exit(1)
    // }

    // conn, err := net.DialTCP("tcp", nil, tcpAddr)
    // if err != nil {
    //     println("Dial failed:", err.Error())
    //     os.Exit(1)
    // }

    // _, err = conn.Write([]byte(strEcho))
    // if err != nil {
    //     println("Write to server failed:", err.Error())
    //     os.Exit(1)
    // }

    rssObject, err := rss.ParseRSS("https://lenta.ru/rss")
    if err != nil {
        for v := range rssObject.Channel.Items {
            item := rssObject.Channel.Items[v]
            // fmt.Println()
            // fmt.Printf("Item Number : %d\n", v)
            strEcho := string(item.Title)

            servAddr := "lab.posevin.com:10049"
            tcpAddr, err := net.ResolveTCPAddr("tcp", servAddr)
            if err != nil {
                println("ResolveTCPAddr failed:", err.Error())
                os.Exit(1)
            }

            conn, err := net.DialTCP("tcp", nil, tcpAddr)
            if err != nil {
                println("Dial failed:", err.Error())
                os.Exit(1)
            }

            _, err = conn.Write([]byte(strEcho))
            if err != nil {
                println("Write to server failed:", err.Error())
                os.Exit(1)
            }
            println("write to server = ", strEcho)

            reply := make([]byte, 1024)

            _, err = conn.Read(reply)
            if err != nil {
                println("Write to server failed:", err.Error())
                os.Exit(1)
            }

            println("reply from server=", string(reply))

            conn.Close()

            // fmt.Printf("Title       : %s\n", item.Title)
            // fmt.Printf("Link        : %s\n", item.Link)
            // fmt.Printf("Description : %s\n", item.Description)
            // fmt.Printf("Guid        : %s\n", item.Guid.Value)
        }
    }
    // strEcho := string(rssObject.Channel.Items[0].Title

    // println("write to server = ", strEcho)

    // reply := make([]byte, 1024)

    // _, err = conn.Read(reply)
    // if err != nil {
    //     println("Write to server failed:", err.Error())
    //     os.Exit(1)
    // }

    // println("reply from server=", string(reply))

    // conn.Close()
}
