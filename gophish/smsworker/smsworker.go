package smsworker

import (
	"context"

	"time"

	log "github.com/gophish/gophish/logger"
	"github.com/gophish/gophish/models"
	"github.com/gophish/gophish/smser"
	"github.com/sirupsen/logrus"
)

// Worker is an interface that defines the operations needed for a background worker
type Worker interface {
	Start()
	LaunchCampaign(c models.Campaign)
}

// DefaultWorker is the background worker that handles watching for new campaigns and sending sms's appropriately.
type DefaultWorker struct {
	smser smser.Smser
}

// New creates a new worker object to handle the creation of campaigns
func New(options ...func(Worker) error) (Worker, error) {
	defaultSmser := smser.NewSmsWorker()
	w := &DefaultWorker{
		smser: defaultSmser,
	}
	for _, opt := range options {
		if err := opt(w); err != nil {
			return nil, err
		}
	}
	return w, nil
}

// WithSmser sets the smser for a given worker.
// By default, workers use a standard, default smsworker.
func WithSmser(s smser.Smser) func(*DefaultWorker) error {
	return func(w *DefaultWorker) error {
		w.smser = s
		return nil
	}
}

// processCampaigns loads smslogs scheduled to be sent before the provided
// time and sends them to the smser.
func (w *DefaultWorker) processCampaigns(t time.Time) error {
	sms, err := models.GetQueuedSmsLogs(t.UTC())
	if err != nil {
		log.Error(err)
		return err
	}
	// Lock the SmsLogs (they will be unlocked after processing)
	err = models.LockSmsLogs(sms, true)
	if err != nil {
		return err
	}
	campaignCache := make(map[int64]models.Campaign)
	// We'll group the smslogs by campaign ID to (roughly) group
	// them by sending profile.
	msg := make(map[int64][]smser.Sms)
	for _, s := range sms {
		// We cache the campaign here to greatly reduce the time it takes to
		// generate the message (ref #1726)
		c, ok := campaignCache[s.CampaignId]
		if !ok {
			c, err = models.GetCampaignSMSContext(s.CampaignId, s.UserId)
			if err != nil {
				return err
			}
			campaignCache[c.Id] = c
		}
		s.CacheCampaign(&c)
		msg[s.CampaignId] = append(msg[s.CampaignId], s)
	}

	// Next, we process each group of smslogs in parallel
	for cid, msc := range msg {
		go func(cid int64, msc []smser.Sms) {
			c := campaignCache[cid]
			if c.Status == models.CampaignQueued {
				err := c.UpdateStatus(models.CampaignInProgress)
				if err != nil {
					log.Error(err)
					return
				}
			}
			log.WithFields(logrus.Fields{
				"num_sms's": len(msc),
			}).Info("Sending sms's to smser for processing")
			w.smser.Queue(msc)
		}(cid, msc)
	}
	return nil
}

// Start launches the worker to poll the database every minute for any pending smslogs
// that need to be processed.
func (w *DefaultWorker) Start() {
	log.Info("Background SMS Worker Started Successfully - Waiting for Campaigns")
	go w.smser.Start(context.Background())
	for t := range time.Tick(1 * time.Minute) {
		err := w.processCampaigns(t)
		if err != nil {
			log.Error(err)
			continue
		}
	}
}

// LaunchCampaign starts a campaign
func (w *DefaultWorker) LaunchCampaign(c models.Campaign) {
	sms, err := models.GetSmsLogsByCampaign(c.Id)
	if err != nil {
		log.Error(err)
		return
	}
	models.LockSmsLogs(sms, true)
	// This is required since you cannot pass a slice of values
	// that implements an interface as a slice of that interface.
	smsEntries := []smser.Sms{}
	currentTime := time.Now().UTC()
	campaignSMSCtx, err := models.GetCampaignSMSContext(c.Id, c.UserId)
	if err != nil {
		log.Error(err)
		return
	}
	for _, s := range sms {
		// Only send the sms's scheduled to be sent for the past minute to
		// respect the campaign scheduling options
		if s.SendDate.After(currentTime) {
			s.Unlock()
			continue
		}
		err = s.CacheCampaign(&campaignSMSCtx)
		if err != nil {
			log.Error(err)
			return
		}
		smsEntries = append(smsEntries, s)
	}
	w.smser.Queue(smsEntries)
}
