package main

import (
        "github.com/mgutz/logxi/v1"
        "golang.org/x/net/html"
        "net/http"
)

var (
        is_second bool
)

func getChildren(node *html.Node) []*html.Node {
        var children []*html.Node
        for c := node.FirstChild; c != nil; c = c.NextSibling {
                children = append(children, c)
        }
        return children
}

func getAttr(node *html.Node, key string) string {
        for _, attr := range node.Attr {
                if attr.Key == key {
                        return attr.Val
                }
        }
        return ""
}

func isText(node *html.Node) bool {
        return node != nil && node.Type == html.TextNode
}

func isElem(node *html.Node, tag string) bool {
        return node != nil && node.Type == html.ElementNode && node.Data == tag
}

func isDiv(node *html.Node, class string) bool {
        return isElem(node, "div") && getAttr(node, "class") == class
}

func readItem(item *html.Node) *Item {
        if a := item.FirstChild; isElem(a, "a") {
                if cs := getChildren(a); len(cs) == 2 && isElem(cs[0], "time") && isText(cs[1]) {
                        return &Item{
                                Ref:   getAttr(a, "href"),
                                Time:  getAttr(cs[0], "title"),
                                Title: cs[1].Data,
                        }
                }
                if cs := getChildren(a); len(cs) == 1 && isText(cs[0]) {
                        return &Item{
                                Ref:   getAttr(a, "href"),
                                Title: cs[0].Data,
                        }
                }
        }
        return nil
}

type Item struct {
        Ref, Time, Title string
}


func downloadNews() []*Item {
        log.Info("sending request to lenta.ru")
        if response, err := http.Get("http://lenta.ru"); err != nil {
                log.Error("request to lenta.ru failed", "error", err)
        } else {
                defer response.Body.Close()
                status := response.StatusCode
                log.Info("got response from lenta.ru", "status", status)
                if status == http.StatusOK {
                        if doc, err := html.Parse(response.Body); err != nil {
                                log.Error("invalid HTML from lenta.ru", "error", err)
                        } else {
                                log.Info("HTML from lenta.ru parsed successfully")
                                return search(doc)
                        }
                }
        }
        return nil
}

func search(node *html.Node) []*Item {
        if isDiv(node, "span4") {
                if is_second {
                        log.Info("------------span4_2------------")
                        is_second = false
                }
                var items []*Item
                for c := node.FirstChild; c != nil; c = c.NextSibling {
                        if isDiv(c, "first-item") {
                                log.Info("------------span4_1------------")
                                is_second = true
                                cs := getChildren(c)
                                if item := readItem(cs[1]); item != nil {
                                        log.Info("title", "val", item.Title)
                                        items = append(items, item)
                                }
                        }
                        if isDiv(c, "item") {
                                if item := readItem(c); item != nil {
                                        log.Info("title", "val", item.Title)
                                        items = append(items, item)
                                }
                        }
                }
                // return items
        }
        if isDiv(node, "b-yellow-box__wrap") {
                log.Info("------------b-yellow-box__wrap------------")
                var items []*Item
                for c := node.FirstChild; c != nil; c = c.NextSibling {
                        if isDiv(c, "item") {
                                if item := readItem(c); item != nil {
                                        log.Info("title", "val", item.Title)
                                        items = append(items, item)
                                }
                        }
                }
                return items
        }
        for c := node.FirstChild; c != nil; c = c.NextSibling {
                if items := search(c); items != nil {
                        return items
                }
        }
        return nil
}





//===================================================================================================



func main() {
        is_second = false
        log.Info("Downloader started")
        downloadNews()


}