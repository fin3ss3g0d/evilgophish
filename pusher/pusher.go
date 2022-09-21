package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
	pusher "github.com/pusher/pusher-http-go/v5"
)

func main() {

	flag.Parse()

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	appID := os.Getenv("PUSHER_APP_ID")
	appKey := os.Getenv("PUSHER_APP_KEY")
	appSecret := os.Getenv("PUSHER_APP_SECRET")
	appCluster := os.Getenv("PUSHER_APP_CLUSTER")
	appIsSecure := os.Getenv("PUSHER_APP_SECURE")

	var isSecure bool
	if appIsSecure == "1" {
		isSecure = true
	}

	client := &pusher.Client{
		AppID:               appID,
		Key:                 appKey,
		Secret:              appSecret,
		Cluster:             appCluster,
		Secure:              isSecure,
		EncryptionMasterKeyBase64: os.Getenv("PUSHER_CHANNELS_ENCRYPTION_KEY"),
		HTTPClient: &http.Client{
			Timeout: time.Minute * 2,
		},
	}

	mux := http.NewServeMux()

	f := &feed{
		mu:   &sync.RWMutex{},
		data: make(map[string]string, 0),
	}

	mux.Handle("/feed", createFeedTitle(client, f))
	mux.Handle("/pusher/auth", authenticateUsers(client))
	mux.Handle("/", http.FileServer(http.Dir("./client")))

	log.Println("Starting local Pusher HTTP server on port 1400")
	log.Println("Start viewing the live feed at: http://localhost:1400/")
	log.Fatal(http.ListenAndServe("127.0.0.1:1400", mux))
}

type feed struct {
	data map[string]string

	mu *sync.RWMutex
}

func (f *feed) exists(title string) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	_, ok := f.data[title]
	return ok
}

func (f *feed) Add(title, content string) error {
	if f.exists(title) {
		return errors.New("title already exists")
	}

	f.mu.Lock()
	defer f.mu.Unlock()
	f.data[title] = content
	return nil
}

const (
	successMsg = "success"
	errorMsg   = "error"
)

func createFeedTitle(client *pusher.Client, f *feed) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if r.Method == http.MethodOptions {
			return
		}

		writer := json.NewEncoder(w)

		type respose struct {
			Message   string `json:"message"`
			Status    string `json:"status"`
			Timestamp int64  `json:"timestamp"`
		}

		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			writer.Encode(&respose{
				Message:   http.StatusText(http.StatusMethodNotAllowed),
				Status:    errorMsg,
				Timestamp: time.Now().Unix(),
			})

			return
		}

		var request struct {
			Title   string `json:"title"`
			Content string `json:"content"`
		}

		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			writer.Encode(&respose{
				Message:   "Invalid request body",
				Status:    errorMsg,
				Timestamp: time.Now().Unix(),
			})
			return
		}

		if len(strings.TrimSpace(request.Title)) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			writer.Encode(&respose{
				Message:   "Title field is empty",
				Status:    errorMsg,
				Timestamp: time.Now().Unix(),
			})
			return
		}

		if len(strings.TrimSpace(request.Content)) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			writer.Encode(&respose{
				Message:   "Content field is empty",
				Status:    errorMsg,
				Timestamp: time.Now().Unix(),
			})
			return
		}

		if err := f.Add(request.Title, request.Content); err != nil {
			w.WriteHeader(http.StatusAlreadyReported)
			writer.Encode(&respose{
				Message:   err.Error(),
				Status:    errorMsg,
				Timestamp: time.Now().Unix(),
			})
			return
		}

		go func() {

			err := client.Trigger("private-encrypted-feeds", "items", map[string]string{
				"title":     request.Title,
				"content":   request.Content,
				"createdAt": time.Now().String(),
			})

			if err != nil {
				fmt.Println(err)
			}

		}()

		w.WriteHeader(http.StatusOK)
		writer.Encode(&respose{
			Message:   "Feed item was successfully added",
			Status:    errorMsg,
			Timestamp: time.Now().Unix(),
		})
	}
}

func authenticateUsers(client *pusher.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if r.Method == http.MethodOptions {
			return
		}

		params, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		presenceData := pusher.MemberData{
			UserID: "10",
			UserInfo: map[string]string{
				"random": "random",
			},
		}

		response, err := client.AuthenticatePresenceChannel(params, presenceData)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Printf("[-] Error authenticating to Pusher! %s\n", err)
			return
		}

		w.Write(response)
	}
}
