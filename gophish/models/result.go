package models

import (
	"crypto/rand"
	"encoding/json"
	"math/big"
	"net"
	"time"

	log "github.com/gophish/gophish/logger"
	"github.com/gorilla/websocket"
	"github.com/jinzhu/gorm"
	"github.com/oschwald/maxminddb-golang"
)

type mmCity struct {
	GeoPoint mmGeoPoint `maxminddb:"location"`
}

type mmGeoPoint struct {
	Latitude  float64 `maxminddb:"latitude"`
	Longitude float64 `maxminddb:"longitude"`
}

// Result contains the fields for a result object,
// which is a representation of a target in a campaign.
type Result struct {
	Id           int64     `json:"-"`
	CampaignId   int64     `json:"-"`
	UserId       int64     `json:"-"`
	RId          string    `json:"id"`
	Status       string    `json:"status" sql:"not null"`
	IP           string    `json:"ip"`
	Latitude     float64   `json:"latitude"`
	Longitude    float64   `json:"longitude"`
	SendDate     time.Time `json:"send_date"`
	Reported     bool      `json:"reported" sql:"not null"`
	ModifiedDate time.Time `json:"modified_date"`
	BaseRecipient
	SMSTarget bool `json:"sms_target"`
}

type FeedEvent struct {
	Event   string `json:"event"`
	Time    string `json:"time"`
	Message string `json:"message"`
}

func (r *Result) NotifyEmailSent() error {
	c, _, err := websocket.DefaultDialer.Dial("ws://localhost:1337/ws", nil)
	if err != nil {
		return err
	}
	defer c.Close()

	fe := FeedEvent{}
	fe.Event = "Email Sent"
	fe.Message = "Email has been sent to victim: <strong>" + r.Email + "</strong>"
	fe.Time = r.ModifiedDate.String()
	data, _ := json.Marshal(fe)

	err = c.WriteMessage(websocket.TextMessage, []byte(string(data)))
	if err != nil {
		return err
	}
	return err
}

func (r *Result) NotifySMSSent() error {
	c, _, err := websocket.DefaultDialer.Dial("ws://localhost:1337/ws", nil)
	if err != nil {
		return err
	}
	defer c.Close()

	fe := FeedEvent{}
	fe.Event = "SMS Sent"
	fe.Message = "SMS has been sent to victim: <strong>" + r.Email + "</strong>"
	fe.Time = r.ModifiedDate.String()
	data, _ := json.Marshal(fe)

	err = c.WriteMessage(websocket.TextMessage, []byte(string(data)))
	if err != nil {
		return err
	}
	return err
}

func (r *Result) createEvent(status string, details interface{}) (*Event, error) {
	e := &Event{Email: r.Email, Message: status}
	if details != nil {
		dj, err := json.Marshal(details)
		if err != nil {
			return nil, err
		}
		e.Details = string(dj)
	}
	AddEvent(e, r.CampaignId)
	return e, nil
}

func (r *Result) HandleSMSSent() error {
	event, err := r.createEvent(EventSent, nil)
	if err != nil {
		return err
	}
	r.SendDate = event.Time
	r.Status = EventSent
	r.ModifiedDate = event.Time
	r.SMSTarget = true

	if conf.FeedEnabled {
		err = r.NotifySMSSent()
		if err != nil {
			log.Error("Error sending websocket message: %v", err)
		}
	}

	return db.Save(r).Error
}

// HandleEmailSent updates a Result to indicate that the email has been
// successfully sent to the remote SMTP server
func (r *Result) HandleEmailSent() error {
	event, err := r.createEvent(EventSent, nil)
	if err != nil {
		return err
	}
	r.SendDate = event.Time
	r.Status = EventSent
	r.ModifiedDate = event.Time
	r.SMSTarget = false

	if conf.FeedEnabled {
		err = r.NotifyEmailSent()
		if err != nil {
			log.Error("Error sending websocket message: %v", err)
		}
	}

	return db.Save(r).Error
}

// HandleEmailError updates a Result to indicate that there was an error when
// attempting to send the email to the remote SMTP server.
func (r *Result) HandleEmailError(err error) error {
	event, err := r.createEvent(EventSendingError, EventError{Error: err.Error()})
	if err != nil {
		return err
	}
	r.Status = Error
	r.ModifiedDate = event.Time
	return db.Save(r).Error
}

// HandleEmailBackoff updates a Result to indicate that the email received a
// temporary error and needs to be retried
func (r *Result) HandleEmailBackoff(err error, sendDate time.Time) error {
	event, err := r.createEvent(EventSendingError, EventError{Error: err.Error()})
	if err != nil {
		return err
	}
	r.Status = StatusRetry
	r.SendDate = sendDate
	r.ModifiedDate = event.Time
	return db.Save(r).Error
}

// HandleEmailOpened updates a Result in the case where the recipient opened the
// email.
func (r *Result) HandleEmailOpened(details EventDetails) error {
	event, err := r.createEvent(EventOpened, details)
	if err != nil {
		return err
	}
	// Don't update the status if the user already clicked the link
	// or submitted data to the campaign
	if r.Status == EventClicked || r.Status == EventDataSubmit {
		return nil
	}
	r.Status = EventOpened
	r.ModifiedDate = event.Time
	return db.Save(r).Error
}

func (r *Result) HandleSMSOpened(details EventDetails) error {
	event, err := r.createEvent(EventOpened, details)
	if err != nil {
		return err
	}
	// Don't update the status if the user already clicked the link
	// or submitted data to the campaign
	if r.Status == EventClicked || r.Status == EventDataSubmit {
		return nil
	}
	r.Status = EventOpened
	r.ModifiedDate = event.Time

	return db.Save(r).Error
}

// HandleClickedLink updates a Result in the case where the recipient clicked
// the link in an email.
func (r *Result) HandleClickedLink(details EventDetails) error {
	event, err := r.createEvent(EventClicked, details)
	if err != nil {
		return err
	}
	// Don't update the status if the user has already submitted data via the
	// landing page form.
	if r.Status == EventDataSubmit {
		return nil
	}
	r.Status = EventClicked
	r.ModifiedDate = event.Time
	return db.Save(r).Error
}

// HandleFormSubmit updates a Result in the case where the recipient submitted
// credentials to the form on a Landing Page.
func (r *Result) HandleFormSubmit(details EventDetails) error {
	event, err := r.createEvent(EventDataSubmit, details)
	if err != nil {
		return err
	}
	r.Status = EventDataSubmit
	r.ModifiedDate = event.Time
	return db.Save(r).Error
}

func (r *Result) HandleCapturedSession(details EventDetails) error {
	event, err := r.createEvent(EventCapturedSession, details)
	if err != nil {
		return err
	}
	r.Status = EventCapturedSession
	r.ModifiedDate = event.Time
	return db.Save(r).Error
}

// HandleEmailReport updates a Result in the case where they report a simulated
// phishing email using the HTTP handler.
func (r *Result) HandleEmailReport(details EventDetails) error {
	event, err := r.createEvent(EventReported, details)
	if err != nil {
		return err
	}
	r.Reported = true
	r.ModifiedDate = event.Time
	return db.Save(r).Error
}

// UpdateGeo updates the latitude and longitude of the result in
// the database given an IP address
func (r *Result) UpdateGeo(addr string) error {
	// Open a connection to the maxmind db
	mmdb, err := maxminddb.Open("static/db/geolite2-city.mmdb")
	if err != nil {
		log.Fatal(err)
	}
	defer mmdb.Close()
	ip := net.ParseIP(addr)
	var city mmCity
	// Get the record
	err = mmdb.Lookup(ip, &city)
	if err != nil {
		return err
	}
	// Update the database with the record information
	r.IP = addr
	r.Latitude = city.GeoPoint.Latitude
	r.Longitude = city.GeoPoint.Longitude
	return db.Save(r).Error
}

func generateResultId() (string, error) {
	const alphaNum = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	// Generate a random length between 8 and 32
	length, err := rand.Int(rand.Reader, big.NewInt(25)) // Generates a number between 0 and 24
	if err != nil {
		return "", err
	}
	finalLength := int(length.Int64()) + 8 // Ensure length is between 8 and 32

	k := make([]byte, finalLength)
	for i := range k {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphaNum))))
		if err != nil {
			return "", err
		}
		k[i] = alphaNum[idx.Int64()]
	}
	return string(k), nil
}

// GenerateId generates a unique key to represent the result
// in the database
func (r *Result) GenerateId(tx *gorm.DB) error {
	// Keep trying until we generate a unique key (shouldn't take more than one or two iterations)
	for {
		rid, err := generateResultId()
		if err != nil {
			return err
		}
		r.RId = rid
		err = tx.Table("results").Where("r_id=?", r.RId).First(&Result{}).Error
		if err == gorm.ErrRecordNotFound {
			break
		}
	}
	return nil
}

// GetResult returns the Result object from the database
// given the ResultId
func GetResult(rid string) (Result, error) {
	r := Result{}
	err := db.Where("r_id=?", rid).First(&r).Error
	return r, err
}
