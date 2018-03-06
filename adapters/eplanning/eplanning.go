package eplanning

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/buger/jsonparser"
	"github.com/mxmCherry/openrtb"
	"github.com/prebid/prebid-server/adapters"
	"github.com/prebid/prebid-server/openrtb_ext"
	"github.com/prebid/prebid-server/pbs"
	"golang.org/x/net/context/ctxhttp"
)

type EPlanningAdapter struct {
	http *adapters.HTTPAdapter
	URI  string
}

type ePlanningRequest struct {
	id      string
	user    *ePlanningUser
	adUnits []*ePlanningAdUnit
}

type ePlanningUser struct {
	userId     string
	clientIp   string
	urlId      string //preguntar como se saca
	locationId string //preguntar como se saca
	connType   string //preguntar como se saca
}

type ePlanningBid struct {
	Id       string  `json:"id,omitempty"`
	BidId    string  `json:"bidid,omitempty"`
	Price    float64 `json:"price,omitempty"`
	Currency string  `json:"cur,omitempty"`
	Width    uint64  `json:"w,omitempty"`
	Height   uint64  `json:"h,omitempty"`
	DealId   string  `json:"dealid,omitempty"`
	Seat     string  `seat:"seat,omitempty"`
}
type ePlanningAdUnit struct {
	id             string
	Currency       string
	Bidfloor       string
	Price          float64
	IsInterstitial bool
	Type           string
	SpaceId        float64
	Client         string
	Video          *ePlanningVideo
	Banner         *ePlanningBanner
}

type ePlanningVideo struct {
	Weight         int
	Height         int
	ScreenPosition string
}

type ePlanningBanner struct {
	Weight         int
	Height         int
	ScreenPosition string
}

func (adapter *EPlanningAdapter) MakeRequests(request *openrtb.BidRequest) ([]*adapters.RequestData, []error) {
	adformRequest, errors := openRtbToEPlanningRequest(request)
	if len(adformRequest.adUnits) == 0 {
		return nil, errors
	}

	requestData := adapters.RequestData{
		Method: "POST",
		Uri:    adapter.URI,
		Body:   adformRequest,
	}

	requests := []*adapters.RequestData{&requestData}

	return requests, errors
}

func openRtbToEPlanningRequest(request *openrtb.BidRequest) (*ePlanningRequest, []error) {

	adUnits := make([]*ePlanningAdUnit, 0, len(request.Imp))
	errors := make([]error, 0, len(request.Imp))
	for _, imp := range request.Imp {

		params, _, _, err := jsonparser.Get(imp.Ext, "bidder")
		if err != nil {
			errors = append(errors, err)
			continue
		}
		var ePlanningAdUnit ePlanningAdUnit
		if err := json.Unmarshal(params, &ePlanningAdUnit); err != nil {
			errors = append(errors, err)
			continue
		}
		adUnits = append(adUnits, &ePlanningAdUnit)
	}
	return &ePlanningRequest{
		agent:   request.Device,
		adUnits: adUnits,
		user:    request.User,
	}, errors
}

func (adapter *EPlanningAdapter) MakeBids(internalRequest *openrtb.BidRequest, externalRequest *adapters.RequestData, response *adapters.ResponseData) ([]*adapters.TypedBid, []error) {
	if response.StatusCode == http.StatusNoContent {
		return nil, nil
	}

	if response.StatusCode != http.StatusOK {
		return nil, []error{fmt.Errorf("unexpected status code: %d. Run with request.debug = 1 for more info", response.StatusCode)}
	}

	ePlanningOutput, err := parseEPlanningBids(response.Body)
	if err != nil {
		return nil, []error{err}
	}

	bids := toOpenRtbBids(ePlanningOutput, internalRequest)

	return bids, nil
}

func NewEPlanningBidder(client *http.Client, endpoint string) *EPlanningAdapter {
	adapter := &adapters.HTTPAdapter{Client: client}

	return &EPlanningAdapter{
		http: a,
		URI:  endpoint,
	}
}

func parseEPlanningBids(response []byte) ([]*ePlanningBid, error) {
	var bids []*ePlanningBid
	if err := json.Unmarshal(response, &bids); err != nil {
		return nil, err
	}

	return bids, nil
}

func toOpenRtbBids(ePlanningBids []*ePlanningBid, r *openrtb.BidRequest) []*adapters.TypedBid {
	bids := make([]*adapters.TypedBid, 0, len(ePlanningBids))

	for i, bid := range ePlanningBids {
		openRtbBid := openrtb.Bid{
			ID:       r.Imp[i].ID,
			ImpID:    r.Imp[i].ID,
			Price:    bid.Price,
			AdM:      bid.Banner,
			W:        bid.Width,
			H:        bid.Height,
			DealID:   bid.DealId,
			BidId:    bid.BidId,
			Currency: bid.Currency,
		}

		bids = append(openRtbBid)
	}

	return bids
}
