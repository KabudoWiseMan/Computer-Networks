package main

import (
    "github.com/mgutz/logxi/v1"
    "golang.org/x/net/html"
    "fmt" // пакет для форматированного ввода вывода
    "net/http" // пакет для поддержки HTTP протокола
    "strings" // пакет для работы с UTF-8 строками
    "html/template"
    "strconv"
)

type Item struct {
    Ticker, Company, Price, Change, Percent string
}

type ViewData struct{
    Col1, Col2, Col3, Col4, Col5 string
    T1, T2, T3, T4, T5 string
    C1, C2, C3, C4, C5 string
    P1, P2, P3, P4, P5 string
    Ch1, Ch2, Ch3, Ch4, Ch5 string
    Per1, Per2, Per3, Per4, Per5 string
}

var (
    t, c, p, ch, per [5]string
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

func numbers(node *html.Node) (string, string, string) {
    if isDiv(node, "emet_index") {
        prS := strings.Replace(getChildren(node)[0].Data, ",", ".", -1)
        pr, _ := strconv.ParseFloat(prS, 64)
        perS := strings.Replace(getChildren((node.NextSibling.NextSibling))[0].Data, ",", ".", -1)
        per, _ := strconv.ParseFloat(perS[:len(perS) - 1], 64)
        start, _ := strconv.ParseFloat(strconv.FormatFloat(pr * 100 / (100 + per), 'f', len(prS) - 2, 64), 64)
        chS := strconv.FormatFloat(pr - start, 'f', len(prS)-2, 64)
        return prS, chS, perS
    }
    for c := node.FirstChild; c != nil; c = c.NextSibling {
        if price, change, percent := numbers(c); price != "nil" {
            return price, change, percent
        }
    }
    return "nil", "nil", "nil"
}

func findNumbers(site string) (string, string, string) {
    if response, err := http.Get(site); err != nil {
            log.Error("request to " + site + " failed", "error", err)
        } else {
            defer response.Body.Close()
            status := response.StatusCode
            if status == http.StatusOK {
                if doc, err := html.Parse(response.Body); err != nil {
                    log.Error("invalid HTML from " + site, "error", err)
                } else {
                    return numbers(doc)
                }
            }
        }
    return "nil", "nil", "nil"
}

func readItem(item *html.Node) *Item {
    if isElem(item, "a") {
        ticker := getAttr(item, "data-symbol")[5:]
        company := item.FirstChild.Data
        site := "https://bcs-express.ru/kotirovki-i-grafiki/" + ticker
        price, change, percent := findNumbers(site)
        return &Item{
            Ticker: ticker,
            Company: company,
            Price: price,
            Change: change,
            Percent: percent,
        }
    }
    return nil
}

func downloadNews() []*Item {
        log.Info("sending request to ru.tradingview.com")
        if response, err := http.Get("https://ru.tradingview.com"); err != nil {
                log.Error("request to ru.tradingview.com failed", "error", err)
        } else {
                defer response.Body.Close()
                status := response.StatusCode
                log.Info("got response from ru.tradingview.com", "status", status)
                if status == http.StatusOK {
                        if doc, err := html.Parse(response.Body); err != nil {
                                log.Error("invalid HTML from ru.tradingview.com", "error", err)
                        } else {
                                log.Info("HTML from ru.tradingview.com parsed successfully")
                                return search(doc)
                        }
                }
        }
        return nil
}

func search(node *html.Node) []*Item {
        if isDiv(node, "tv-mainmenu") {
            var items []*Item
            shares := getChildren(getChildren(getChildren(getChildren(node.FirstChild)[1])[1])[1])[2]
            count := 0
            for sh, i := shares.FirstChild, 0; sh != nil && i < 9; sh = sh.NextSibling {
                i++         
                if item := readItem(sh); item != nil {   
                    items = append(items, item)
                    t[count] = item.Ticker
                    c[count] = item.Company
                    p[count] = item.Price
                    ch[count] = item.Change
                    per[count] = item.Percent
                    count++
                }               
            }
        }
        for c := node.FirstChild; c != nil; c = c.NextSibling {
                if items := search(c); items != nil {
                        return items
                }
        }
        return nil
}

func parser(w http.ResponseWriter, r *http.Request) {    
    tmpl, _ := template.ParseFiles("index.html")
        log.Info("Downloader started")
        downloadNews()
        var color [5]string
        for i := range color {
            if (ch[i])[0] == '-' {
                color[i] = "red"
            } else {
                color[i] = "green"
            }
        }
        data := ViewData{
            Col1: color[0],
            Col2: color[1],
            Col3: color[2],
            Col4: color[3],
            Col5: color[4],
            T1: t[0],
            T2: t[1],
            T3: t[2],
            T4: t[3],
            T5: t[4],
            C1: c[0],
            C2: c[1],
            C3: c[2],
            C4: c[3],
            C5: c[4],
            P1: p[0],
            P2: p[1],
            P3: p[2],
            P4: p[3],
            P5: p[4],
            Ch1: ch[0],
            Ch2: ch[1],
            Ch3: ch[2],
            Ch4: ch[3],
            Ch5: ch[4],
            Per1: per[0],
            Per2: per[1],
            Per3: per[2],
            Per4: per[3],
            Per5: per[4],
        }
        tmpl.Execute(w, data)
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
    http.HandleFunc("/", HomeRouterHandler) // установим роутер
    err := http.ListenAndServe(":10073", nil) // задаем слушать порт
    if err != nil {
        log.Fatal("ListenAndServe: ", err)
    }
}