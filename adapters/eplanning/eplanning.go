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
	http    *adapters.HTTPAdapter
	URI     string
	version string
}

type ePlanningRequest struct {
	id         string
	userAgent  string
	ip         string
	bidderCode string
	isSecure   bool
	referer    string
	userId     string
	adUnits    []*ePlanningAdUnit
}

type ePlanningAdUnit struct {
	Currency       string
	Bidfloor       string
	Price          float64
	IsInterstitial bool
	Type           string
	SpaceId        float64
	Client         string
}

type ePlanningBid struct {
	ResponseType string  `json:"response,omitempty"`
	Banner       string  `json:"banner,omitempty"`
	Price        float64 `json:"win_bid,omitempty"`
	Currency     string  `json:"win_cur,omitempty"`
	Width        uint64  `json:"width,omitempty"`
	Height       uint64  `json:"height,omitempty"`
	DealId       string  `json:"deal_id,omitempty"`
}

func (adapter *EPlanningAdapter) MakeRequests(request *openrtb.BidRequest) ([]*adapters.RequestData, []error) {
	adformRequest, errors := openRtbToEPlanningRequest(request)
	if len(adformRequest.adUnits) == 0 {
		return nil, errors
	}

	requestData := adapters.RequestData{
		Method:  "GET",
		Uri:     adformRequest.buildAdformUrl(a),
		Body:    nil,
		Headers: adformRequest.buildAdformHeaders(a),
	}

	requests := []*adapters.RequestData{&requestData}

	return requests, errors
}

func openRtbToEPlanningRequest(request *openrtb.BidRequest) (*ePlanningRequest, []error) {
	adUnits := make([]*ePlanningAdUnit, 0, len(request.Imp))
	errors := make([]error, 0, len(request.Imp))
	secure := false
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

		if !secure && imp.Secure != nil && *imp.Secure == 1 {
			secure = true
		}

		ePlanningAdUnit.bidId = imp.ID
		ePlanningAdUnit.adUnitCode = imp.ID
		adUnits = append(adUnits, &ePlanningAdUnit)
	}

	referer := ""
	if request.Site != nil {
		referer = request.Site.Page
	}

	tid := ""
	if request.Source != nil {
		tid = request.Source.TID
	}

	return &ePlanningRequest{
		adUnits:   adUnits,
		ip:        request.Device.IP,
		userAgent: request.Device.UA,
		isSecure:  secure,
		referer:   referer,
		userId:    request.User.BuyerUID,
		tid:       tid,
	}, errors
}
