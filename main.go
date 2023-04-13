package main

import (
    "bytes"
    "fmt"
    "io/ioutil"
    "log"
    "net/http"
    "net/url"
    "golang.org/x/net/html"
    "path"
    "strings"
    "io"
)


func main() {
    http.HandleFunc("/", proxyHandler)
    http.HandleFunc("/proxy", proxyHandler)
    http.HandleFunc("/proxyresult", proxyResultHandler)
    log.Fatal(http.ListenAndServe(":8080", nil))
}

func proxyHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method == "GET" {
        if r.URL.Path == "/style.css" {
            cssBytes, err := ioutil.ReadFile("static/style.css")
            if err != nil {
                http.Error(w, "Failed to read style.css file", http.StatusInternalServerError)
                return
            }
            w.Header().Set("Content-Type", "text/css; charset=utf-8")
            w.Write(cssBytes)
            return
        }
        if r.URL.Path == "/404.css" {
            cssBytes, err := ioutil.ReadFile("static/404.css")
            if err != nil {
                http.Error(w, "Failed to read style.css file", http.StatusInternalServerError)
                return
            }
            w.Header().Set("Content-Type", "text/css; charset=utf-8")
            w.Write(cssBytes)
            return
        }
        if r.URL.Path == "/petah.png" {
            pngBytes, err := ioutil.ReadFile("static/petah.png")
            if err != nil {
                http.Error(w, "Failed to read petah.png file", http.StatusInternalServerError)
                return
            }
            w.Header().Set("Content-Type", "image/png")
            w.Write(pngBytes)
            return
        }
        if r.URL.Path == "/lowresstewie.jpeg" {
            pngBytes, err := ioutil.ReadFile("static/lowresstewie.jpeg")
            if err != nil {
                http.Error(w, "Failed to read stewie file", http.StatusInternalServerError)
                return
            }
            w.Header().Set("Content-Type", "image/jpeg")
            w.Write(pngBytes)
            return
        }
        if r.URL.Path == "/" || r.URL.Path == "/index.html" {
            htmlBytes, err := ioutil.ReadFile("static/index.html")
            if err != nil {
                http.Error(w, "Failed to read index.html file", http.StatusInternalServerError)
                return
            }

            w.Header().Set("Content-Type", "text/html; charset=utf-8")
            w.Write(htmlBytes)
            return
        }
      w.Header().Set("Content-Type", "text/html; charset=utf-8")
      w.WriteHeader(http.StatusNotFound)
      four04Bytes, err := ioutil.ReadFile("static/404.html")
      if err != nil {
          http.Error(w, "Failed to read 404.html file", http.StatusInternalServerError)
          return
      }
      w.Write(four04Bytes)
    } else {
        http.Redirect(w, r, "/proxyresult?url="+r.FormValue("url"), http.StatusSeeOther)
    }
}


func proxyResultHandler(w http.ResponseWriter, r *http.Request) {
    url := r.FormValue("url")
    if url == "" {
        fmt.Fprintf(w, "<html><head><title>Error</title></head><body><p>Invalid URL</p></body></html>")
        return
    }

    resp, err := http.Get(url)
    if err != nil {
        fmt.Fprintf(w, "<html><head><title>Error</title></head><body><p>%v</p></body></html>", err)
        return
    }
    defer resp.Body.Close()

    contentType := resp.Header.Get("Content-Type")
    if !strings.HasPrefix(contentType, "text/html") {
        w.Header().Set("Content-Type", contentType)
        io.Copy(w, resp.Body)
        return
    }

    doc, err := html.Parse(resp.Body)
    if err != nil {
        fmt.Fprintf(w, "<html><head><title>Error</title></head><body><p>%v</p></body></html>", err)
        return
    }

    // Rewrite hrefs
    rewriteHrefs(doc, resp.Request.URL)

    var buf bytes.Buffer
    if err := html.Render(&buf, doc); err != nil {
        fmt.Fprintf(w, "<html><head><title>Error</title></head><body><p>%v</p></body></html>", err)
        return
    }

    w.Header().Set("Content-Type", contentType)
    w.Write(buf.Bytes())
}

// Also does src

func rewriteHrefs(doc *html.Node, proxyUrl *url.URL) {
    if doc.Type == html.ElementNode && doc.Data == "a" || doc.Data == "link" || doc.Data == "img" {
        for i := range doc.Attr {
            if doc.Attr[i].Key == "href" || doc.Attr[i].Key == "src" {
                hrefUrl, err := url.Parse(doc.Attr[i].Val)
                if err == nil {
                    if hrefUrl.IsAbs() {
                        // Replace absolute URL with proxied URL
                        doc.Attr[i].Val = "/proxyresult?url=" + proxyUrl.String() + hrefUrl.Path
                    } else {
                        // Replace relative URL with proxied URL
                        newUrl := *proxyUrl
                        newUrl.Path = path.Join(proxyUrl.Path, hrefUrl.Path)
                        doc.Attr[i].Val = "/proxyresult?url=" + newUrl.String()
                    }
                }
            }
        }
    }

    for c := doc.FirstChild; c != nil; c = c.NextSibling {
        rewriteHrefs(c, proxyUrl)
    }
}