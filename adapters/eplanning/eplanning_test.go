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

func TestSovrnOpenRtbRequest(t *testing.T) {
	service := CreateSovrnService(adapterstest.BidOnTags(""))
	server := service.Server
	ctx := context.Background()
	req := SampleSovrnRequest(1, t)
	bidder := req.Bidders[0]
	adapter := NewSovrnAdapter(adapters.DefaultHTTPAdapterConfig, server.URL)
	adapter.Call(ctx, req, bidder)

	adapterstest.VerifyIntValue(len(service.LastBidRequest.Imp), 1, t)
	adapterstest.VerifyStringValue(service.LastBidRequest.Imp[0].TagID, "123456", t)
	adapterstest.VerifyBannerSize(service.LastBidRequest.Imp[0].Banner, 728, 90, t)
	checkHttpRequest(*service.LastHttpRequest, t)
}


func TestEPlanningOpenRtbRequest(t *testing.T) {
	service := CreateEPlanningService(adapterstest.BidOnTags(""))
	server := service.Server
}

func CreateEPlanningService(tagsToBid map[string]bool) adapterstest.OrtbMockService {
	service := adapterstest.OrtbMockService{}
	var lastBidRequest openrtb.BidRequest
	var lastHttpReq http.Request

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lastHttpReq = *r
		defer r.Body.Close()
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		var breq openrtb.BidRequest
		err = json.Unmarshal(body, &breq)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		lastBidRequest = breq
		var bids []openrtb.Bid
		for i, imp := range breq.Imp {
			if tagsToBid[imp.TagID] {
				bids = append(bids, adapterstest.SampleBid(imp.Banner.W, imp.Banner.H, imp.ID, i+1))
			}
		}

		// serialize the bids to openrtb.BidResponse
		js, _ := json.Marshal(openrtb.BidResponse{
			SeatBid: []openrtb.SeatBid{
				{
					Bid: bids,
				},
			},
		})
		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	}))

	service.Server = server
	service.LastBidRequest = &lastBidRequest
	service.LastHttpRequest = &lastHttpReq

	return service
}



}
