package smser

import (
	"context"

	log "github.com/gophish/gophish/logger"
	"github.com/twilio/twilio-go"
	openapi "github.com/twilio/twilio-go/rest/api/v2010"
)

type TwilioMessage struct {
	Client twilio.RestClient
	Params openapi.CreateMessageParams
}

// Smser is an interface that defines an object used to queue and
// send mailer.Sms instances.
type Smser interface {
	Start(ctx context.Context)
	Queue([]Sms)
}

// Sms is an interface that handles the common operations for sms messages
type Sms interface {
	Error(err error) error
	Success() error
	Generate(msg *TwilioMessage) error
	Backoff(err error) error
}

// SmsWorker is the worker that receives slices of sms's
type SmsWorker struct {
	queue chan []Sms
}

// NewSmsWorker returns an instance of SmsWorker with the mail queue
// initialized.
func NewSmsWorker() *SmsWorker {
	return &SmsWorker{
		queue: make(chan []Sms),
	}
}

// Start launches the mail worker to begin listening on the Queue channel
// for new slices of Sms instances to process.
func (sw *SmsWorker) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case sms := <-sw.queue:
			go func(ctx context.Context, sms []Sms) {
				sendSms(ctx, sms)
			}(ctx, sms)
		}
	}
}

// Queue sends the provided mail to the internal queue for processing.
func (sw *SmsWorker) Queue(sms []Sms) {
	sw.queue <- sms
}

// sendSms attempts to send the provided Sms instances.
// If the context is cancelled before all of the sms are sent,
// sendSms just returns and does not modify those sms's.
func sendSms(ctx context.Context, sms []Sms) {
	for _, s := range sms {
		select {
		case <-ctx.Done():
			return
		default:
			break
		}
		// Generate the message
		message := &TwilioMessage{}
		err := s.Generate(message)
		if err != nil {
			log.Warn(err)
			s.Error(err)
			continue
		}
		// Send the message
		_, err = message.Client.Api.CreateMessage(&message.Params)
		if err != nil {
			log.Warn(err)
			s.Backoff(err)
			continue
		}
		s.Success()
	}
}
