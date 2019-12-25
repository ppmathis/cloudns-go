package cloudns

import (
	"context"
	"net"
)

type ZoneType string
type ZoneKind string

const (
	ZoneMaster ZoneType = "master"
	ZoneSlave  ZoneType = "slave"
	ZoneParked ZoneType = "parked"
	ZoneGeoDNS ZoneType = "geodns"
)

const (
	ZoneDomain ZoneKind = "domain"
	ZoneIPv4   ZoneKind = "ipv4"
	ZoneIPv6   ZoneKind = "ipv6"
)

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

func (api *API) AvailableNameservers(ctx context.Context) (results []Nameserver, err error) {
	err = api.request(ctx, "POST", "/dns/available-name-servers.json", nil, nil, &results)
	return
}

func (api *API) ListZones(ctx context.Context) (results []Zone, err error) {
	var pageCount int
	params := HttpParams{"rows-per-page": 100}

	// Get the amount of pages for the current query
	err = api.request(ctx, "POST", "/dns/get-pages-count.json", params, nil, &pageCount)
	if err != nil {
		return
	}

	// Retrieve all pages and gather the results
	for page := 1; page <= pageCount; page++ {
		var pageResults []Zone
		params["page"] = page

		err = api.request(ctx, "POST", "/dns/list-zones.json", params, nil, &pageResults)
		if err != nil {
			return
		}

		results = append(results, pageResults...)
	}

	return
}

func (api *API) ZoneUsage(ctx context.Context) (result ZoneUsage, err error) {
	err = api.request(ctx, "POST", "/dns/get-zones-stats.json", nil, nil, &result)
	return
}

func (api *API) ZoneInfo(ctx context.Context, name string) (result Zone, err error) {
	params := HttpParams{"domain-name": name}
	err = api.request(ctx, "POST", "/dns/get-zone-info.json", params, nil, &result)
	return
}

func (api *API) ZoneSetActive(ctx context.Context, name string, isActive bool) error {
	params := HttpParams{"domain-name": name, "status": isActive}
	return api.request(ctx, "POST", "/dns/change-status.json", params, nil, nil)
}
