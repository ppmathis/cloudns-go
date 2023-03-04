package cloudns

import (
	"context"
	"net"
	"strings"
)

const zoneAvailableNameserversURL = "/dns/available-name-servers.json"
const zoneListURL = "/dns/list-zones.json"
const zoneCreateUrl = "/dns/register.json"
const zoneGetURL = "/dns/get-zone-info.json"
const zoneTriggerUpdateURL = "/dns/update-zone.json"
const zoneUpdateStatusURL = "/dns/update-status.json"
const zoneIsUpdatedURL = "/dns/is-updated.json"
const zoneSetActiveURL = "/dns/change-status.json"
const zoneUsageURL = "/dns/get-zones-stats.json"
const zonePageCountURL = "/dns/get-pages-count.json"
const zoneRowsPerPage = 100

// ZoneType is an enumeration of all supported zone types
type ZoneType int

// Enumeration values for ZoneType
const (
	ZoneTypeUnknown ZoneType = iota
	ZoneTypeMaster
	ZoneTypeSlave
	ZoneTypeParked
	ZoneTypeGeoDNS
)

// ZoneKind is an enumeration of all supported zone kinds
type ZoneKind int

// Enumeration values for ZoneKind
const (
	ZoneKindUnknown ZoneKind = iota
	ZoneKindDomain
	ZoneKindIPv4
	ZoneKindIPv6
)

// ZoneService is a service object which groups all operations related to ClouDNS zone management
type ZoneService struct {
	api *Client
}

// Zone represents a ClouDNS record according to the official API docs
type Zone struct {
	Name     string   `json:"name"`
	Type     ZoneType `json:"type"`
	Kind     ZoneKind `json:"zone"`
	IsActive APIBool  `json:"status"`
}

// ZoneUsage represents the current zone usage for a ClouDNS account
type ZoneUsage struct {
	Current int `json:"count,string"`
	Limit   int `json:"limit,string"`
}

// Nameserver represents a ClouDNS nameserver according to the official API docs
type Nameserver struct {
	Type          string  `json:"type"`
	Name          string  `json:"name"`
	IPv4          net.IP  `json:"ip4"`
	IPv6          net.IP  `json:"ip6"`
	Location      string  `json:"location"`
	CountryCode   string  `json:"location_cc"`
	DDoSProtected APIBool `json:"ddos_protected"`
}

// ZoneUpdateStatus represents the current update status of a nameserver for a given zone
type ZoneUpdateStatus struct {
	Server    string  `json:"server"`
	IPv4      string  `json:"ip4"`
	IPv6      string  `json:"ip6"`
	IsUpdated APIBool `json:"updated"`
}

type CreateZone struct {
	Name     string   `json:"name"`
	Type     ZoneType `json:"type"`
	Ns       []string `json:"ns"`
	MasterIp string   `json:"master_ip"`
}

// AsParams returns the HTTP parameters for a zone to use within the create zone API method
func (zone CreateZone) AsParams() HTTPParams {
	params := HTTPParams{
		"domain-name": zone.Name,
		"ns":          zone.Ns,
	}
	switch zone.Type {
	case ZoneTypeMaster:
		params["zone-type"] = "master"
	case ZoneTypeGeoDNS:
		params["zone-type"] = "geo-dns"
	case ZoneTypeParked:
		params["zone-type"] = "parked"
	case ZoneTypeSlave:
		params["zone-type"] = "slave"
		params["master-ip"] = zone.MasterIp
	case ZoneTypeUnknown:
		params["zone-type"] = "unknown"
	}

	return params
}

// NewZone instantiates a new CreateZone which can be used within ClouDNS API methods. It does -not- add this zone
// automatically.
func NewZone(name string, zoneType ZoneType, ns []string, masterIp string) CreateZone {
	return CreateZone{
		Name:     name,
		Type:     zoneType,
		Ns:       ns,
		MasterIp: masterIp,
	}
}

// List returns all zones
// Official Docs: https://www.cloudns.net/wiki/article/50/
func (svc *ZoneService) List(ctx context.Context) ([]Zone, error) {
	return svc.Search(ctx, "", 0)
}

// Search returns all zones matching a given name and/or group ID
// Official Docs: https://www.cloudns.net/wiki/article/50/
func (svc *ZoneService) Search(ctx context.Context, search string, groupID int) ([]Zone, error) {
	var err error
	var pageCount int
	var pageResults []Zone

	// Build search parameters for zone querying
	params := HTTPParams{"rows-per-page": zoneRowsPerPage}
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

// Get returns a zone with a given name
// Official Docs: https://www.cloudns.net/wiki/article/134/
func (svc *ZoneService) Get(ctx context.Context, zoneName string) (result Zone, err error) {
	params := HTTPParams{"domain-name": zoneName}
	err = svc.api.request(ctx, "POST", zoneGetURL, params, nil, &result)
	return
}

// TriggerUpdate triggers a manual update for a given zone
// Official Docs: https://www.cloudns.net/wiki/article/135/
func (svc *ZoneService) TriggerUpdate(ctx context.Context, zoneName string) (result StatusResult, err error) {
	params := HTTPParams{"domain-name": zoneName}
	err = svc.api.request(ctx, "POST", zoneTriggerUpdateURL, params, nil, &result)
	return
}

// SetActive enables or disables a zone with the given name
// Official Docs: https://www.cloudns.net/wiki/article/55/
func (svc *ZoneService) SetActive(ctx context.Context, zoneName string, isActive bool) (result StatusResult, err error) {
	params := HTTPParams{"domain-name": zoneName}
	if isActive {
		params["status"] = 1
	} else {
		params["status"] = 0
	}

	err = svc.api.request(ctx, "POST", zoneSetActiveURL, params, nil, &result)
	return
}

// IsUpdated returns a boolean if the given zone has been updated to all ClouDNS nameservers
// Official Docs: https://www.cloudns.net/wiki/article/54/
func (svc *ZoneService) IsUpdated(ctx context.Context, zoneName string) (result bool, err error) {
	params := HTTPParams{"domain-name": zoneName}
	err = svc.api.request(ctx, "POST", zoneIsUpdatedURL, params, nil, &result)
	return
}

// GetUpdateStatus returns a list of all nameservers for the given zone with their update status
// Official Docs: https://www.cloudns.net/wiki/article/53/
func (svc *ZoneService) GetUpdateStatus(ctx context.Context, zoneName string) (result []ZoneUpdateStatus, err error) {
	params := HTTPParams{"domain-name": zoneName}
	err = svc.api.request(ctx, "POST", zoneUpdateStatusURL, params, nil, &result)
	return
}

// AvailableNameservers returns all nameservers available for the current account
// Official Docs: https://www.cloudns.net/wiki/article/47/
func (svc *ZoneService) AvailableNameservers(ctx context.Context) (result []Nameserver, err error) {
	err = svc.api.request(ctx, "POST", zoneAvailableNameserversURL, nil, nil, &result)
	return
}

// GetUsage returns the current zone usage for the current account (actual usage and maximum zones for current plan)
// Official Docs: https://www.cloudns.net/wiki/article/52/
func (svc *ZoneService) GetUsage(ctx context.Context) (result ZoneUsage, err error) {
	err = svc.api.request(ctx, "POST", zoneUsageURL, nil, nil, &result)
	return
}

// Create a new zone
func (svc *ZoneService) Create(ctx context.Context, zone CreateZone) (result StatusResult, err error) {
	params := zone.AsParams()

	err = svc.api.request(ctx, "POST", zoneCreateUrl, params, nil, &result)
	return
}

// UnmarshalJSON converts the ClouDNS zone type into the correct ZoneType enumeration value
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

// UnmarshalJSON converts the ClouDNS zone type into the correct ZoneType enumeration value
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
