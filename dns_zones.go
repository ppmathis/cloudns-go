package cloudns

import (
	"context"
	"net"
	"strings"
)

const zoneAvailableNameserversURL = "/dns/available-name-servers.json"
const zoneListURL = "/dns/list-zones.json"
const zoneGetURL = "/dns/get-zone-info.json"
const zoneTriggerUpdateURL = "/dns/update-zone.json"
const zoneUpdateStatusURL = "/dns/update-status.json"
const zoneIsUpdatedURL = "/dns/is-updated.json"
const zoneSetActiveURL = "/dns/change-status.json"
const zoneUsageURL = "/dns/get-zones-stats.json"
const zonePageCountURL = "/dns/get-pages-count.json"
const zoneRowsPerPage = 100

type ZoneType int
type ZoneKind int

const (
	ZoneTypeUnknown ZoneType = iota
	ZoneTypeMaster
	ZoneTypeSlave
	ZoneTypeParked
	ZoneTypeGeoDNS
)
const (
	ZoneKindUnknown ZoneKind = iota
	ZoneKindDomain
	ZoneKindIPv4
	ZoneKindIPv6
)

type zoneService struct {
	api *Client
}

type Zone struct {
	Name     string   `json:"name"`
	Type     ZoneType `json:"type"`
	Kind     ZoneKind `json:"zone"`
	IsActive APIBool  `json:"status"`
}

type ZoneUsage struct {
	Current int `json:"count,string"`
	Limit   int `json:"limit,string"`
}

type Nameserver struct {
	Type          string  `json:"type"`
	Name          string  `json:"name"`
	IPv4          net.IP  `json:"ip4"`
	IPv6          net.IP  `json:"ip6"`
	Location      string  `json:"location"`
	CountryCode   string  `json:"location_cc"`
	DDoSProtected APIBool `json:"ddos_protected"`
}

type ZoneUpdateStatus struct {
	Server    string  `json:"server"`
	IPv4      string  `json:"ip4"`
	IPv6      string  `json:"ip6"`
	IsUpdated APIBool `json:"updated"`
}

func (svc *zoneService) List(ctx context.Context) ([]Zone, error) {
	return svc.Search(ctx, "", 0)
}

func (svc *zoneService) Search(ctx context.Context, search string, groupID int) ([]Zone, error) {
	var err error
	var pageCount int
	var pageResults []Zone

	// Build search parameters for zone querying
	params := HttpParams{"rows-per-page": zoneRowsPerPage}
	if search != "" {
		params["search"] = search
	}
	if groupID != 0 {
		params["group-id"] = groupID
	}

	// Fetch number of available pages
	err = svc.api.request(ctx, "POST", zonePageCountURL, params, nil, &pageCount)
	if err != nil {
		return nil, err
	}

	// Fetch all pages iteratively and gather the results together
	results := make([]Zone, 0, pageCount*zoneRowsPerPage)
	for pageIndex := 1; pageIndex <= pageCount; pageIndex++ {
		params["page"] = pageIndex
		err = svc.api.request(ctx, "POST", zoneListURL, params, nil, &pageResults)
		if err != nil {
			return nil, err
		}

		results = append(results, pageResults...)
	}

	return results, nil
}

func (svc *zoneService) Get(ctx context.Context, zoneName string) (result Zone, err error) {
	params := HttpParams{"domain-name": zoneName}
	err = svc.api.request(ctx, "POST", zoneGetURL, params, nil, &result)
	return
}

func (svc *zoneService) TriggerUpdate(ctx context.Context, zoneName string) (result BaseResult, err error) {
	params := HttpParams{"domain-name": zoneName}
	err = svc.api.request(ctx, "POST", zoneTriggerUpdateURL, params, nil, &result)
	return
}

func (svc *zoneService) SetActive(ctx context.Context, zoneName string, isActive bool) (result BaseResult, err error) {
	params := HttpParams{"domain-name": zoneName}
	if isActive {
		params["status"] = 1
	} else {
		params["status"] = 0
	}

	err = svc.api.request(ctx, "POST", zoneSetActiveURL, params, nil, &result)
	return
}

func (svc *zoneService) IsUpdated(ctx context.Context, zoneName string) (result bool, err error) {
	params := HttpParams{"domain-name": zoneName}
	err = svc.api.request(ctx, "POST", zoneIsUpdatedURL, params, nil, &result)
	return
}

func (svc *zoneService) GetUpdateStatus(ctx context.Context, zoneName string) (result []ZoneUpdateStatus, err error) {
	params := HttpParams{"domain-name": zoneName}
	err = svc.api.request(ctx, "POST", zoneUpdateStatusURL, params, nil, &result)
	return
}

func (svc *zoneService) AvailableNameservers(ctx context.Context) (result []Nameserver, err error) {
	err = svc.api.request(ctx, "POST", zoneAvailableNameserversURL, nil, nil, &result)
	return
}

func (svc *zoneService) GetUsage(ctx context.Context) (result ZoneUsage, err error) {
	err = svc.api.request(ctx, "POST", zoneUsageURL, nil, nil, &result)
	return
}

func (zt *ZoneType) UnmarshalJSON(data []byte) error {
	switch strings.Trim(string(data), `"`) {
	case "master":
		*zt = ZoneTypeMaster
	case "slave":
		*zt = ZoneTypeSlave
	case "parked":
		*zt = ZoneTypeParked
	case "geodns":
		*zt = ZoneTypeGeoDNS
	default:
		*zt = ZoneTypeUnknown
	}

	return nil
}
func (zk *ZoneKind) UnmarshalJSON(data []byte) error {
	switch strings.Trim(string(data), `"`) {
	case "domain":
		*zk = ZoneKindDomain
	case "ipv4":
		*zk = ZoneKindIPv4
	case "ipv6":
		*zk = ZoneKindIPv6
	default:
		*zk = ZoneKindUnknown
	}

	return nil
}
