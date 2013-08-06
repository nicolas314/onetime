// One-time Sharing
// Select a local file for sharing through a URL served by the same host
// The URL is meant for a one-time download only. In effect:
// The first time the URL is clicked, it is stamped as 'activated'. Further
// requests on the same URL will be honored for the next 4 hours, after
// which the host will refuse to serve it with a 404.
package main

import (
    "crypto/rand"
    "encoding/base64"
    "encoding/json"
    "errors"
    "fmt"
    "io"
    "io/ioutil"
    "log"
    "net/http"
    "os"
    "path"
    "path/filepath"
    "strings"
    "time"
)

const (
    ONETIME_SZ = 8   // Length of a one-time token
    // Token validity once clicked, in seconds
    TOKEN_VAL  = time.Duration(4*60*60) * time.Second
    CNF_NAME   = "/onetime.json"
)

type Config struct {
    TOKEN_DB        string
    BASE_ADDR       string
    LOG_FILE        string
    CRT             string
    KEY             string
    path            string
}

// Yeah, global. So what?
var cnf Config

// Return an ISO8601 time repr
func isotime(t time.Time) string {
    if t.Year()<=1970 {
        return "no"
    }
    return fmt.Sprintf("%04d-%02d-%02d %02d:%02d:%02d",
                       t.Year(), t.Month(), t.Day(),
                       t.Hour(), t.Minute(), t.Second())
}

// Pretty-print a file size: comma-separate digits
// Not really a smart implementation
func prettySize(sz int64) string {
    ssz:=fmt.Sprintf("%d", sz)
    var pr []string
    switch i:=len(ssz); i {
        case 1, 2, 3:
        pr = []string{ssz}
        case 4, 5, 6:
        pr = []string{string(ssz[0:i-3]),
                       string(ssz[i-3:])}
        case 7, 8, 9:
        pr = []string{string(ssz[0:i-6]),
                       string(ssz[i-6:i-3]),
                       string(ssz[i-3:])}
        case 10, 11, 12:
        pr = []string{string(ssz[0:i-9]),
                       string(ssz[i-9:i-6]),
                       string(ssz[i-6:i-3]),
                       string(ssz[i-3:])}
        default:
        pr= []string{""}
    }
    return strings.Join(pr, ",")
}


// Generate a one-time token of length sz
func GenerateOnetime(sz int) string {
    // Character set used to create one-time tokens
    letterSet := "1234567890" +
                 "abcdefghijklmnopqrstuvwxyz" +
                 "1234567890"

    // Get sz random bytes
    pick := make([]byte, sz)
    n, err := io.ReadFull(rand.Reader, pick)
    if n!=sz || err!=nil {
        // Cannot do much in case of random generator failure. Bailout
        panic(err)
    }
    // Pick sz characters at random
    ott:=""
    for i:=0 ; i<sz ; i++ {
        ott=ott+string(letterSet[int(pick[i]) % len(letterSet)])
    }
    return ott
}

// A Token is a path (served) and creation/activation times
type Token struct {
    Path        string
    Created     time.Time
    Activated   time.Time
}

// List of Tokens as an object
type LTokens map[string] Token

// Save a list of Tokens
func (ltok LTokens) Save(filename string) {
    js, _ := json.Marshal(ltok)
    ioutil.WriteFile(filename, js, 0644)
}

// Load a list of Tokens
func (ltok LTokens) Load(filename string) {
    js, _ := ioutil.ReadFile(filename)
    json.Unmarshal(js, &ltok)
}

// Add a Token to a list
func (ltok LTokens) Add(filename string) {
    // Add leading path if it was not provided
    ffilename, _ := filepath.Abs(filename)
    // Check file exists and is readable
    sta, err := os.Stat(ffilename)
    if err != nil {
        fmt.Println("cannot find file: %s", ffilename)
        return
    }
    if sta.IsDir() {
        fmt.Println("cannot send directories")
        return
    }
    ott := GenerateOnetime(ONETIME_SZ)
    now := time.Now()
    ltok[ott] = Token{ffilename, now, time.Unix(0,0)}
    fmt.Printf(`

A file is ready for download
Name: %s
Size: %s bytes
URL: %s/%s

`,  sta.Name(),
    prettySize(sta.Size()),
    cnf.BASE_ADDR, ott)
}

// Delete a Token from a list
func (ltok LTokens) Del(ott string) {
    fmt.Printf("removing token: %s\n", ott)
    delete(ltok, ott)
}

// Show all Tokens in the list
func (ltok LTokens) List() {
    for k, v := range ltok {
        fmt.Printf(`

    token: %s
      url: %s/%s
     file: %s
  created: %s
activated: %s
 validity: %s

`, k, cnf.BASE_ADDR, k, v.Path, isotime(v.Created), isotime(v.Activated),
   isotime(v.Activated.Add(TOKEN_VAL)))
    }
}

// Purge expired tokens
func (ltok LTokens) Purge() {
    now := time.Now()
    for k, v := range ltok {
        if isotime(v.Activated)!="no" && now.Sub(v.Activated) > TOKEN_VAL {
            ltok.Del(k)
        }
    }
}


// Return a hardcoded favicon
// Seems stupid to hardcode this but avoids having to locate
// the damn file and a file read for each request
func Favicon(w http.ResponseWriter, req * http.Request) {
    fav64:=`
AAABAAEAEBAAAAAAAABoBAAAFgAAACgAAAAQAAAAIAAAAAEAIAAAAAAAAAQAAAAAAAAAAAAAAAAA
AAAAAAD///8A////AP///wD///8A////AP///wD///8A////AP///wD///8A////AP///wD///8A
////AP///wD///8A////AP///wB/wIYAgtGKAI/flwCe7qUAq/qyAMn/zgCu+bUAnOykAI7dlQCA
zIgAf7OEAH+kgwD///8A////AP///wD///8AAIINAAakFQAgvy8APdxMOUPhTraT/50BXfRrADrZ
SQAdvCwAApoRAABoCgAASggA////AP///wD///8A////AACCDQAGpBUAIb8wN1TOW/+N/5j/dup+
uVz0awE62UkAHbwsAAKaEQAAaAoAAEoIAP///wD///8A////AP///wAAgg0AB6QWN0a9Tv9b+mr/
YP5v/2v/ef883Ui5OtlJAR28LAACmhEAAGgKAABKCAD///8A////AP///wD///8AAYEOOES4TP9D
4lL/h/CR/4HyjP9S8WH/VvVl/yvINrkdvCwBApoRAABoCgAASggA////AP///wD///8AFXcfNk2v
Vv8oxzf/jOSU/4Xljv9/5oj/eeiD/0PiUv9H51b/FKQguAKaEQEAaAoAAEoIAP///wD///8ABnoR
KnG8ef+V2pv/ltyd/47clf9VyF//h9qP+Xfcgf9x3Xv/M9JC/zfXRv8Chw65AGgKAQBKCAD///8A
////AAODEA2FyIz/rNiw/57XpP9Uv1z/Dq0dHBi3JwIRsCD5cNN5/2jTcv8jwjL/JsY1/wNfC7QA
SggB////AP///wD///8AAJcPDQ2kF/9dwGX/A6ESHQ6tHQAYtycAD64eAgWkFPloyXH/X8dp/xCv
H/8UsyP/A0wLsz93RQH///8A////AP///wAKqRkLC6gaFQKhEQAOrR0AGLcnAA+uHgADohICAJIO
+Wm5cP9buWX/AJkP/wCfD/85cT+1P3VFAf///wD///8ACqkZAAyqGgACoREADq0dABi3JwAPrh4A
A6ISAACYDwIAmA75bK1x/1+nZv8Adwz/Pp1H/wBIBx7///8A////AAqpGQAMqhoAAqERAA6tHQAY
tycAD64eAAOiEgAAmA8AAJ0PAgCZDvlvqXX/d658/wBjC0f///8A////AP///wAKqRkADKoaAAKh
EQAOrR0AGLcnAA+uHgADohIAAJgPAACdDwAAng8CAJcO+BqNJUf///8A////AP///wD///8AhNSM
AIXUjACA0IgAhtaOAIvbkwCH1o4AgdCIAH/LhwB/zocAf86HAH/KhgB/wYYA////AP///wD///8A
////AP///wD///8A////AP///wD///8A////AP///wD///8A////AP///wD///8A////AP///wD/
//8A//8AAP//AAD9/wAA+P8AAPB/AADgPwAAwB8AAIAPAACGBwAAzwMAAP+BAAD/wQAA/+MAAP/3
AAD//wAA//8AAA==`

    enc := base64.StdEncoding
    fav,_ := enc.DecodeString(fav64)
    // log.Println(req.RemoteAddr, req.URL, "favicon")
    w.Write(fav)
}


// Send a web page showing download links
func Show(w http.ResponseWriter, req * http.Request) {
    reqpath:=req.URL.Path[1:]
    // log.Println("GET", req.RemoteAddr, req.URL)
    ltok := make(LTokens)
    ltok.Load(cnf.TOKEN_DB)
    tok, err := ltok[reqpath]
    if err==false {
        log.Println("404", req.RemoteAddr, req.URL)
        http.NotFound(w, req)
        return
    }
    name := path.Base(tok.Path)
    sta, s_err := os.Stat(tok.Path)
    if s_err!=nil {
        log.Println("NOFILE", req.RemoteAddr, req.URL)
        http.NotFound(w, req)
        return
    }
    validity_period:=""
    if tok.Activated.Year()>1970 {
        validity_period="<dt>Valid until</dt><dd>"+
                         isotime(tok.Activated.Add(TOKEN_VAL))+
                        "</dd>"
    }
    log.Println("DISP", req.RemoteAddr, req.URL)
    fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
<link href='http://fonts.googleapis.com/css?family=Ubuntu' rel='stylesheet' type='text/css'>
<style type="text/css">
body {
    margin: 5%%;
    max-width: 768px;
    background-color: #9999ff;
    font-family: 'Ubuntu', sans-serif;
}
#main {
    background-color: #6666cc;
    color: white;
    padding: 10px;
    border-radius: 15px;
}
#top {
    font-weight: bold;
}
#disclaimer {
    font-style: italic;
}
a {
    color: white;
}
</style>
<meta http-equiv="Content-Type" content="text/html; charset=UTF-8" />
<title>
Download
</title>
</head>
<body>
    <div id="main">
    <p id="top">A file is ready to be retrieved:</p>
    <dl>
        <dt>Name</dt>
        <dd>%s</dd>
        <dt>Size</dt>
        <dd>%s bytes</dd>
        %s
        <dt>Link</dt>
        <dd><a href="/d/%s">Click here to start downloading</a></dd>
    </dl>
    </div>
    <p id="disclaimer">
    This link is only valid once. It will remain valid up to four hours
    after it has first been clicked.
    </p>
</body>
</html>`, name, prettySize(sta.Size()), validity_period, reqpath)
}

// Send the real data
func Distribute(w http.ResponseWriter, req * http.Request) {
    reqpath:=req.URL.Path[3:]
    // log.Println(req.RemoteAddr, req.URL)
    ltok := make(LTokens)
    ltok.Load(cnf.TOKEN_DB)
    tok, err := ltok[reqpath]
    if err==false {
        log.Println("404", req.RemoteAddr, req.URL)
        http.NotFound(w, req)
        return
    }
    if tok.Activated.Year()>1970 {
        if time.Now().Sub(tok.Activated) > TOKEN_VAL {
            log.Println("EXPIRED", req.RemoteAddr, req.URL)
            http.NotFound(w, req)
            return
        }
    }
    ltok[reqpath] = Token{tok.Path, tok.Created, time.Now()}
    ltok.Save(cnf.TOKEN_DB)
    name := path.Base(tok.Path)
    log.Println("SEND", req.RemoteAddr, req.URL)
    w.Header().Set("Content-disposition",
                   fmt.Sprintf("attachment; filename=\"%s\"", name))
    http.ServeFile(w, req, tok.Path)
    log.Println("DONE", req.RemoteAddr, reqpath)
}

// Server configure and start
func Serve() {
    fmt.Printf(`

      config: %s
    TOKEN_DB: %s
    LOG_FILE: %s
   BASE_ADDR: %s
         CRT: %s
         KEY: %s

`, cnf.path, cnf.TOKEN_DB, cnf.LOG_FILE, cnf.BASE_ADDR, cnf.CRT, cnf.KEY)
    logf, _ := os.OpenFile(cnf.LOG_FILE,
                           os.O_WRONLY|os.O_APPEND|os.O_CREATE,
                           0666)
    log.SetOutput(logf)
    defer logf.Close()
    http.HandleFunc("/favicon.ico", Favicon)
    http.HandleFunc("/d/", Distribute)
    http.HandleFunc("/", Show)

    log.Println("START", cnf.BASE_ADDR)
    // Choose http or https depending on BASE_ADDR
    var err error
    if strings.HasPrefix(cnf.BASE_ADDR, "https") {
        err = http.ListenAndServeTLS(cnf.BASE_ADDR[8:],
                                     cnf.CRT,
                                     cnf.KEY,
                                     nil)
    } else if strings.HasPrefix(cnf.BASE_ADDR, "http") {
        err = http.ListenAndServe(cnf.BASE_ADDR[7:], nil)
    } else {
        err = errors.New("unknown protocol in BASE_ADDR")
    }

    if err!=nil {
        log.Fatal(err)
        return
    }
}

// Create a default configuration file
func setConfiguration() {
    name,_ := os.Readlink("/proc/self/exe")
    cname  := path.Dir(name)+CNF_NAME

    fo, err := os.Create(cname)
    if err!=nil {
        fmt.Println("cannot create config file: ", cname)
        return
    }
    defer fo.Close()

    fmt.Fprintf(fo,
                     `{
    "TOKEN_DB": "token.db",
    "LOG_FILE": "onetime.log",
   "BASE_ADDR": "http://localhost:2500",
         "CRT": "server.crt",
         "KEY": "server.key"
}
`)
    fmt.Println("Config file created: ", cname)
    fmt.Println("Edit this file before launching the server")
}

// Read configuration from file
func readConfiguration() error {
    // Locate config file if it exists
    name,_ := os.Readlink("/proc/self/exe")
    cpath := path.Dir(name)
    cnf.path = cpath+CNF_NAME

    // Load config file
    js, err := ioutil.ReadFile(cnf.path)
    if err!=nil {
        return err
    }
    json.Unmarshal(js, &cnf)
    // Check all required values are there
    if len(cnf.TOKEN_DB)>0 {
        if cnf.TOKEN_DB[0]!='/' {
            cnf.TOKEN_DB = cpath+"/"+cnf.TOKEN_DB
        }
    } else {
        return errors.New("TOKEN_DB undefined in "+cnf.path)
    }
    if len(cnf.LOG_FILE)>0 {
        if cnf.LOG_FILE[0]!='/' {
            cnf.LOG_FILE = cpath+"/"+cnf.LOG_FILE
        }
    } else {
        return errors.New("LOG_FILE undefined in "+cnf.path)
    }
    if len(cnf.BASE_ADDR)<1 {
        return errors.New("BASE_ADDR undefined in "+cnf.path)
    }
    if len(cnf.CRT)>0 {
        if cnf.CRT[0]!='/' {
            cnf.CRT = cpath+"/"+cnf.CRT
        }
    }
    if len(cnf.KEY)>0 {
        if cnf.KEY[0]!='/' {
            cnf.KEY = cpath+"/"+cnf.KEY
        }
    }
    return nil
}

//----------------- main
func main() {
    if len(os.Args)<2 {
        fmt.Println(`
        
    use:
    onetime config          Configure server
    onetime serve           Serve onetime requests
    onetime add path        Create onetime request for path
    onetime ls              List existing requests
    onetime del token       Delete onetime request
    onetime purge           Delete all expired tokens

`)
        return
    }

    err := readConfiguration()
    if err!=nil && os.Args[1]!="config" {
        fmt.Println(err)
        return
    }
    ltok := make(LTokens)
    switch os.Args[1] {
        case "config":
        setConfiguration()
        case "serve", "server":
        Serve()
        case "add", "create":
        if len(os.Args)>=3 {
            ltok.Load(cnf.TOKEN_DB)
            ltok.Add(os.Args[2])
            ltok.Save(cnf.TOKEN_DB)
        }
        case "ls", "list":
        ltok.Load(cnf.TOKEN_DB)
        ltok.List()
        case "del", "delete", "rm":
        if len(os.Args)>=2 {
            ltok.Load(cnf.TOKEN_DB)
            for i:=2 ; i<len(os.Args) ; i++ {
                ltok.Del(os.Args[i])
            }
            ltok.Save(cnf.TOKEN_DB)
        }
        case "purge":
        ltok.Load(cnf.TOKEN_DB)
        ltok.Purge()
        ltok.Save(cnf.TOKEN_DB)
    }
    return
}

