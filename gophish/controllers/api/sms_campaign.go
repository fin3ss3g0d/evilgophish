package api

import (
	"encoding/json"
	"net/http"

	ctx "github.com/gophish/gophish/context"
	log "github.com/gophish/gophish/logger"
	"github.com/gophish/gophish/models"
)

// Campaigns returns a list of campaigns if requested via GET.
// If requested via POST, APICampaigns creates a new campaign and returns a reference to it.
func (as *Server) SMSCampaigns(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == "GET":
		cs, err := models.GetCampaigns(ctx.Get(r, "user_id").(int64))
		if err != nil {
			log.Error(err)
		}
		JSONResponse(w, cs, http.StatusOK)
	//POST: Create a new campaign and return it as JSON
	case r.Method == "POST":
		c := models.Campaign{}
		// Put the request into a campaign
		err := json.NewDecoder(r.Body).Decode(&c)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "Invalid JSON structure"}, http.StatusBadRequest)
			return
		}
		err = models.PostSMSCampaign(&c, ctx.Get(r, "user_id").(int64))
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusBadRequest)
			return
		}
		// If the campaign is scheduled to launch immediately, send it to the worker.
		// Otherwise, the worker will pick it up at the scheduled time
		if c.Status == models.CampaignInProgress {
			go as.smsworker.LaunchCampaign(c)
		}
		JSONResponse(w, c, http.StatusCreated)
	}
}
