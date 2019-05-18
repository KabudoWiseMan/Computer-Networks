package main

import (
        "github.com/mgutz/logxi/v1"
        "golang.org/x/net/html"
        "net/http"
        "strconv"
        "strings"
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

type Item struct {
        Ticker, Company, Price, Change, Percent string
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
			for sh, i := shares.FirstChild, 0; sh != nil && i < 9; sh = sh.NextSibling {
				i++			
				if item := readItem(sh); item != nil {
					log.Info("Share", 
						"Ticker", item.Ticker, 
						"Company", item.Company, 
						"Price", item.Price,
						"Change", item.Change,
						"Percent", item.Percent)	
					items = append(items, item)
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





//===================================================================================================



func main() {

        log.Info("Downloader started")
        downloadNews()
}