package usersync

import (
	"fmt"
	"net/url"

	"github.com/prebid/prebid-server/pbs"
)

func NewEPlanningSyncer(externalURL string) Usersyncer {
	redirectUri := fmt.Sprintf("%s/setuid?bidder=eplanning&uid=$UID", externalURL)
	usersyncURL := "http://sync.e-planning.net/um?uid"
	info := &pbs.UsersyncInfo{
		URL:         fmt.Sprintf("%s%s", usersyncURL, url.QueryEscape(redirectUri)),
		Type:        "redirect",
		SupportCORS: false,
	}

	return &syncer{
		familyName: "eplanning",
		syncInfo:   info,
	}
}
