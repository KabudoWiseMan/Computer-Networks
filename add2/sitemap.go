package main

import (
    "github.com/mgutz/logxi/v1"
    "golang.org/x/net/html"
    "fmt" // пакет для форматированного ввода вывода
    "net/http" // пакет для поддержки HTTP протокола
    "strings" // пакет для работы с  UTF-8 строками
    "net/url"
    "path"
)

var (
    links map[string]Link
)

type Link struct {
    Title, Description, Keywords string
}

func isDiv(node *html.Node, class string) bool {
    return isElem(node, "div") && getAttr(node, "class") == class
}

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

func getAttr(node *html.Node, key string) string {
    for _, attr := range node.Attr {
            if attr.Key == key {
                    return attr.Val
            }
    }
    return ""
}

func getChildren(node *html.Node) []*html.Node {
    var children []*html.Node
    for c := node.FirstChild; c != nil; c = c.NextSibling {
            children = append(children, c)
    }
    return children
}

func readItem(item string) (string, string, string) {
    log.Info("sending request to " + item)
    if response, err := http.Get(item); err != nil {
            log.Error("request to " + item + " failed", "error", err)
    } else {
        defer response.Body.Close()
        status := response.StatusCode
        if status == http.StatusOK {
            if doc, err := html.Parse(response.Body); err != nil {
                log.Error("invalid HTML from " + item, "error", err)
            } else {
                t, d, k := ttl(doc), desc(doc), kw(doc)
                if t == "" || t == "no" {
                	t = "no title"
                }
                if d == "" || d == "no" {
                	d = "no description"
                }
                if k == "" || k == "no" {
                	k = "no keywords"
                }
                return t, d, k
            }
        }
    }
    return "no title", "no description", "no keywords"
}

func ttl(node *html.Node) string {
    if isElem(node, "title") {
        return node.FirstChild.Data
    }
    if isElem(node, "body") {
    	return "no"
    }
    for c := node.FirstChild; c != nil; c = c.NextSibling {
        if title := ttl(c); title != "" {
            return title
        }
    }
    return ""
}

func desc(node *html.Node) string {
    if isElem(node, "meta") {
        if getAttr(node, "name") == "description" {
            return getAttr(node, "content")
        }
    }
    if isElem(node, "body") {
    	return "no"
    }
    for c := node.FirstChild; c != nil; c = c.NextSibling {
        if description := desc(c); description != "" {
            return description
        }
    }
    return ""
}

func kw(node *html.Node) string {
    if isElem(node, "meta") {
        if getAttr(node, "name") == "keywords" {            
            return getAttr(node, "content")
        }
    }
    if isElem(node, "body") {
    	return "no"
    }
    for c := node.FirstChild; c != nil; c = c.NextSibling {
        if keywords := kw(c); keywords != "" {
            return keywords
        }
    }
    return ""
}

func search2(link string, base *url.URL) {
	log.Info("sending request to " + link)
    if response, err := http.Get(link); err != nil {
            log.Error("request to " + link + " failed", "error", err)
    } else {
        defer response.Body.Close()
        status := response.StatusCode
        if status == http.StatusOK {
            if doc, err := html.Parse(response.Body); err != nil {
                log.Error("invalid HTML from " + link, "error", err)
            } else {
                search(doc, base)
            }
        }
    }
}

func isFile(link string, base *url.URL) bool {
    s := ".json.aif.cda.mid.midi.mp3.mpa.ogg.wav.wma.wpl.bin.dmg.iso.toast.vcd.csv.dat.db.dbf.log.mdb.sav.sql.tar.apk.bat.bin.cgi.pl.exe.gadget.jar.py.wsf.go.fnt.fon.otf.ttf.ai.bmp.gif.ico.jpeg.jpg.png.ps.psd.svg.tif.tiff.rss.php.jsp.js.css.asp.aspx.key.odp.pps.ppt.pptx.c.cpp.cs.h.java.sh.swift.vb.ods.xlr.xls.xlsx.3g2.3gp.avi.flv.h264.m4v.mkv.mov.mp4.mpg.mpeg.rm.swf.vob.wmv.doc.docx.odt.pdf.rtf.tex.txt.wks.wps.wpd"
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

func search(node *html.Node, base *url.URL) {
    if isElem(node, "link") || 
    isElem(node, "a") ||
    isElem(node, "area") || 
    isElem(node, "base") {
        i := getAttrIndex(node, "href")
        isStyle := getAttr(node, "rel")
        if i != -1 && isStyle != "stylesheet" && node.Attr[i].Val[0] != '#' {
            aLink := absLink(node.Attr[i].Val, base)
            _, ok := links[aLink]
            if !ok && strings.Contains(aLink, base.Host) && strings.Contains(aLink, "http") && !isFile(aLink, base) {
                title, description, keywords := readItem(aLink)
                link := Link{ 
                    Title: title, 
                    Description: description, 
                    Keywords: keywords,
                }
                log.Info("Link", "Url", aLink, "Title", title, "Description", description, "Keywords", keywords)
                links[aLink] = link
                search2(aLink, base)
            }
        }
    }
    for c := node.FirstChild; c != nil; c = c.NextSibling {
    	search(c, base)
    }
}

func write(w http.ResponseWriter, encoding string) {
	if strings.Contains(encoding, "text/html") {
		w.Header().Set("Content-type", encoding)
	} else {
		w.Header().Set("Content-type", "text/html; charset=" + encoding)
	}
    fmt.Fprintf(w, "<!DOCTYPE html>\n<html>\n\t<head>\n\t</head>\n\t<body>\n")
    for key, value := range links {
        fmt.Fprintf(w, "\t\t<div>")
        fmt.Fprintf(w, "\t\t\t<p>URL: " + key + "</p>\n")
        fmt.Fprintf(w, "\t\t\t<p>TITLE: " + value.Title + "\n</p>")
        fmt.Fprintf(w, "\t\t\t<p>DESCRIPTION: " + value.Description + "\n</p>")
        fmt.Fprintf(w, "\t\t\t<p>KEYWORDS: " + value.Keywords + "</p>\n")
        fmt.Fprintf(w, "\t\t</div>\n")
        fmt.Fprintf(w, "\t\t<br>\n")
    }
    fmt.Fprintf(w, "\t</body>\n")
    fmt.Fprintf(w, "</html>")
}

func countRec(node *html.Node) int {
	if isElem(node, "urlset") {
		return len(getChildren(node))
	}
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if num := countRec(c); num != -1 {
            return num
        }
    }
    return -1
}

func count(site string) int {
	log.Info("sending request to " + site + " sitemap")
    if response, err := http.Get(site); err != nil {
        log.Error("request to " + site + " failed", "error", err)
    } else {
        defer response.Body.Close()
        status := response.StatusCode
        log.Info("got response from " + site, "status", status)
        if status == http.StatusOK {
            if doc, err := html.Parse(response.Body); err != nil {
                log.Error("invalid HTML from " + site, "error", err)
            } else {
                log.Info("HTML from " + site + " parsed successfully")
                return countRec(doc)
            }
        }
    }
    return -1
}

func enc(node *html.Node) string {
    if isElem(node, "meta") {
        if getAttr(node, "http-equiv") == "Content-Type" || getAttr(node, "http-equiv") == "content-type" {            
            return getAttr(node, "content")
        }
    }
    if isElem(node, "meta") {
        if getAttr(node, "charset") != "" {            
            return getAttr(node, "charset")
        }
    }
    if isElem(node, "body") {
    	return "no"
    }
    for c := node.FirstChild; c != nil; c = c.NextSibling {
        if encoding := enc(c); encoding != "" {
            return encoding
        }
    }
    return ""
}

func parser(w http.ResponseWriter, r *http.Request) {    
    log.Info("Downloader started")
    ourPath := strings.Replace(string(r.URL.Path)[1:], ":/", "://", 1)
    log.Info("sending request to " + ourPath)
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
                encoding := enc(doc)
                if encoding == "" || encoding == "no" {
                	encoding = "utf-8"
                } 
                search(doc, base)                
                write(w, encoding)
                log.Info("Count", "Sitemap", count(ourPath + "/sitemap.xml"), "My program", len(links))
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
    parser(w, r) // отправляем данные на клиентскую сторону
}

func main() {
    links = make(map[string]Link)
    http.HandleFunc("/", HomeRouterHandler) // установим роутер
    err := http.ListenAndServe(":10073", nil) // задаем слушать порт
    if err != nil {
        log.Fatal("ListenAndServe: ", err)
    }
}