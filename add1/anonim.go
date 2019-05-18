package main

import (
    "github.com/mgutz/logxi/v1"
    "golang.org/x/net/html"
    "fmt" // пакет для форматированного ввода вывода
    "net/http" // пакет для поддержки HTTP протокола
    "strings" // пакет для работы с  UTF-8 строками
    // "strconv"
    "io"
    "bytes"
    "net/url"
    "html/template"
    "path"
    "os"
    // "io/ioutil"
)

func isElem(node *html.Node, tag string) bool {
    return node != nil && node.Type == html.ElementNode && node.Data == tag
}

func getAttrIndex(node *html.Node, key string) (int) {
    for i, attr := range node.Attr {
        if attr.Key == key {
            return i
        }
    }
    return -1
}

func isFile(link string, base *url.URL) bool {
    s := ".json.aif.cda.mid.midi.mp3.mpa.ogg.wav.wma.wpl.bin.dmg.iso.toast.vcd.csv.dat.db.dbf.log.mdb.sav.sql.tar.xml.apk.bat.bin.cgi.pl.exe.gadget.jar.py.wsf.go.fnt.fon.otf.ttf.ai.bmp.gif.ico.jpeg.jpg.png.ps.psd.svg.tif.tiff.rss.xhtml.php.jsp.js.css.htm.html.asp.aspx.key.odp.pps.ppt.pptx.c.cpp.cs.h.java.sh.swift.vb.ods.xlr.xls.xlsx.3g2.3gp.avi.flv.h264.m4v.mkv.mov.mp4.mpg.mpeg.rm.swf.vob.wmv.doc.docx.odt.pdf.rtf.tex.txt.wks.wps.wpd"
    if strings.Contains(link, base.Host) && path.Ext(link) != "" && strings.Contains(s, path.Ext(link)) {
        return true 
    }
    return false
}

func absLink(link string, base *url.URL) string {
    u, err := url.Parse(link)
    if err != nil {
        return link
    }
    return base.ResolveReference(u).String()
}

func newFileLink(link string, base *url.URL) string {
    if strings.Contains(link, base.Host) {
        name := path.Base(link)
        downloadFile(link, name)
        return "/files/" + name
    }
    return link
}

func downloadFile(link string, name string) {
    _, err := os.Stat("./files/" + name)
    if os.IsNotExist(err) {
        out, err := os.Create("./files/" + name)
        if err != nil {
            log.Error("Couldn't create a file", err)
        }
        defer out.Close()
        resp, err := http.Get(link)
        if err != nil {
            log.Error("Couldn't connect to link", err)
        }
        defer resp.Body.Close()
        _, err = io.Copy(out, resp.Body)
        if err != nil {
            log.Error("Couldn't download a file", err)
        }
    }
}

// func deleteAllFiles() {
//     dir, err := ioutil.ReadDir("./files")
//     if err != nil {
//         log.Error("Directory doesn't exist", err)
//     }
//     for _, d := range dir {
//         os.RemoveAll(path.Join([]string{"files", d.Name()}...))
//     }
// }

// func isUrl(s string) bool {
//     _, err := url.ParseRequestURI(s)
//     if err != nil {
//         return false
//     }
//     return true
// }

func newLink(link string, base *url.URL) string {
    if strings.Contains(link, base.Host) {
        return "http://lab.posevin.com:10073/" + link
    }
    return link
}

func search(w http.ResponseWriter, node *html.Node, base *url.URL) {
    if isElem(node, "link") || 
    isElem(node, "a") ||
    isElem(node, "area") || 
    isElem(node, "base") {
        i := getAttrIndex(node, "href")
        if i != -1 {
            aLink := absLink(node.Attr[i].Val, base)
            if isFile(aLink, base) {
                node.Attr[i].Val = newFileLink(aLink, base)
            } else {
                node.Attr[i].Val = newLink(aLink, base)
            }
        }
    }
    if isElem(node, "script") || 
    isElem(node, "iframe") || 
    isElem(node, "img") || 
    isElem(node, "frame") || 
    isElem(node, "audio") ||  
    isElem(node, "embed") || 
    isElem(node, "input") ||  
    isElem(node, "source") ||  
    isElem(node, "track") || 
    isElem(node, "video") {
        i := getAttrIndex(node, "src")
        if i != -1 {
            log.Info("LINK: ", node.Attr[i].Val)
            aLink := absLink(node.Attr[i].Val, base)
            node.Attr[i].Val = newFileLink(aLink, base)
        }
    }
    for c := node.FirstChild; c != nil; c = c.NextSibling {
        search(w, c, base)
    }
}

func renderNode(n *html.Node) string {
    var buf bytes.Buffer
    w := io.Writer(&buf)
    html.Render(w, n)
    return buf.String()
}

func open(w http.ResponseWriter, doc *html.Node) {
    htm := renderNode(doc)
    tmpl, _ := template.New("i").Parse(htm)
    tmpl.Execute(w, nil)
}

func parser(w http.ResponseWriter, r *http.Request) {    
    log.Info("Downloader started")
    // log.Info("sending request to https://190.24.10.4:8080 proxy")
    // proxyUrl, err := url.Parse("https://190.24.10.4:8080")
    // if err != nil {
    //     log.Error("request to https://190.24.10.4:8080 proxy failed", "error", err)
    // }
    // myClient := &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)}}
    ourPath := strings.Replace(string(r.URL.Path)[1:], ":/", "://", 1)
    log.Info("sending request to " + ourPath)
    // if response, err := myClient.Get(strings.Replace(string(r.URL.Path)[1:], ":/", "://", 1)); err != nil {
    if response, err := http.Get(ourPath); err != nil {
        log.Error("request to " + ourPath + " failed", "error", err)
    } else {
        defer response.Body.Close()
        status := response.StatusCode
        log.Info("got response from " + ourPath, "status", status)
        if status == http.StatusOK {
            if doc, err := html.Parse(response.Body); err != nil {
                log.Error("invalid HTML from " + ourPath, "error", err)
            } else {
                log.Info("HTML from " + ourPath + " parsed successfully")
                base, err := url.Parse(ourPath)
                if err != nil {
                    log.Error(ourPath + " not a path", err)
                }
                search(w, doc, base)
                open(w, doc)
            }
        }
    }
}

func HomeRouterHandler(w http.ResponseWriter, r *http.Request) {
    r.ParseForm() //анализ аргументов,
    fmt.Println(r.Form)  // ввод информации о форме на стороне сервера
    fmt.Println("path", r.URL.Path)
    fmt.Println("scheme", r.URL.Scheme)
    fmt.Println(r.Form["url_long"])
    for k, v := range r.Form {
        fmt.Println("key:", k)
        fmt.Println("val:", strings.Join(v, ""))
    }
    if r.URL.Path == "/" {
        tmpl, _ := template.ParseFiles("index.html")
        tmpl.Execute(w, nil)
    } else {
        parser(w, r) // отправляем данные на клиентскую сторону
        // deleteAllFiles()
    }
}

func main() {
    http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
    http.Handle("/files/", http.StripPrefix("/files/", http.FileServer(http.Dir("./files"))))
    http.HandleFunc("/", HomeRouterHandler) // установим роутер
    err := http.ListenAndServe(":10073", nil) // задаем слушать порт
    if err != nil {
        log.Fatal("ListenAndServe: ", err)
    }
}