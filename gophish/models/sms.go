package models

import (
	"errors"
	"time"

	log "github.com/gophish/gophish/logger"
)

// SMS contains the attributes needed to handle the sending of campaign SMS messages
type SMS struct {
	Id               int64     `json:"id" gorm:"column:id; primary_key:yes"`
	UserId           int64     `json:"-" gorm:"column:user_id"`
	Name             string    `json:"name"`
	TwilioAccountSid string    `json:"account_sid"`
	TwilioAuthToken  string    `json:"auth_token"`
	SMSFrom          string    `json:"sms_from"`
	ModifiedDate     time.Time `json:"modified_date"`
}

var ErrAccountSidNotSpecified = errors.New("No Twilio Account SID Specified")
var ErrAuthTokenNotSpecified = errors.New("No Twilio Auth Token Specified")

func (s *SMS) Validate() error {
	var err error

	if s.TwilioAccountSid == "" {
		err = ErrAccountSidNotSpecified
	} else if s.TwilioAuthToken == "" {
		err = ErrAuthTokenNotSpecified
	}

	return err
}

// TableName specifies the database tablename for Gorm to use
func (s SMS) TableName() string {
	return "sms"
}

// GetSMSs returns the SMSs owned by the given user.
func GetSMSs(uid int64) ([]SMS, error) {
	ss := []SMS{}
	err := db.Where("user_id=?", uid).Find(&ss).Error
	if err != nil {
		log.Error(err)
		return ss, err
	}

	return ss, nil
}

// GetSMS returns the SMS, if it exists, specified by the given id and user_id.
func GetSMS(id int64, uid int64) (SMS, error) {
	s := SMS{}
	err := db.Where("user_id=? and id=?", uid, id).Find(&s).Error
	if err != nil {
		log.Error(err)
		return s, err
	}

	return s, err
}

// GetSMSByName returns the SMS, if it exists, specified by the given name and user_id.
func GetSMSByName(n string, uid int64) (SMS, error) {
	s := SMS{}
	err := db.Where("user_id=? and name=?", uid, n).Find(&s).Error
	if err != nil {
		log.Error(err)
		return s, err
	}
	return s, err
}

// PostSMS creates a new SMS in the database.
func PostSMS(s *SMS) error {
	err := s.Validate()
	if err != nil {
		log.Error(err)
		return err
	}
	// Insert into the DB
	err = db.Save(s).Error
	if err != nil {
		log.Error(err)
	}

	return err
}

// PutSMS edits an existing SMTP in the database.
// Per the PUT Method RFC, it presumes all data for a SMS is provided.
func PutSMS(s *SMS) error {
	err := s.Validate()
	if err != nil {
		log.Error(err)
		return err
	}
	err = db.Where("id=?", s.Id).Save(s).Error
	if err != nil {
		log.Error(err)
	}

	return err
}

// DeleteSMS deletes an existing SMS in the database.
// An error is returned if a SMS with the given user id and SMS id is not found.
func DeleteSMS(id int64, uid int64) error {
	var err = db.Where("user_id=?", uid).Delete(SMS{Id: id}).Error
	if err != nil {
		log.Error(err)
	}
	return err
}
