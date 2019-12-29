package cloudns

import (
	"context"
	"encoding/json"
	"errors"
)

const recordSOAGetURL = "/dns/soa-details.json"
const recordSOAUpdateURL = "/dns/modify-soa.json"
const recordGetDynamicURL = "/dns/get-dynamic-url.json"
const recordDisableDynamicURL = "/dns/disable-dynamic-url.json"
const recordChangeDynamicURL = "/dns/change-dynamic-url.json"
const recordAvailableTTLsURL = "/dns/get-available-ttl.json"
const recordAvailableRecordTypesURL = "/dns/get-available-record-types.json"
const recordCopyFromZoneURL = "/dns/copy-records.json"
const recordImportURL = "/dns/records-import.json"
const recordExportURL = "/dns/records-export.json"
const recordImportTransferURL = "/dns/axfr-import.json"
const recordListURL = "/dns/records.json"
const recordCreateURL = "/dns/add-record.json"
const recordUpdateURL = "/dns/mod-record.json"
const recordDeleteURL = "/dns/delete-record.json"
const recordSetActiveURL = "/dns/change-record-status.json"

// RecordFormat is an enumeration of all supported record formats
type RecordFormat int

// Enumeration values for RecordFormat
const (
	RecordFormatBIND RecordFormat = iota
	RecordFormatTinyDNS
)

// RecordService is a service object which groups all operations related to ClouDNS record management
type RecordService struct {
	api *Client
}

// RecordMap represents a map of records indexed by the record ID
type RecordMap map[int]Record

// Record represents a ClouDNS record according to the official API docs
type Record struct {
	ID         int     `json:"id,string,omitempty"`
	Host       string  `json:"host"`
	Record     string  `json:"record"`
	RecordType string  `json:"type"`
	TTL        int     `json:"ttl,string"`
	IsActive   APIBool `json:"status"`

	Priority         uint16 `json:"priority,string,omitempty"`
	Weight           uint16 `json:"weight,string,omitempty"`
	Port             uint16 `json:"port,string,omitempty"`
	GeoDNSLocationID int    `json:"geodns-location,omitempty"`

	RP
	CAA
	NAPTR
	TLSA
	SSHFP
	WebRedirect
}

// RP represents parameters specifically for RP records
type RP struct {
	Mail string `json:"mail,omitempty"`
	TXT  string `json:"txt,omitempty"`
}

// SSHFP represents parameters specifically for SSHFP records
type SSHFP struct {
	Algorithm uint8 `json:"algorithm,string,omitempty"`
	Type      uint8 `json:"fp_type,string,omitempty"`
}

// CAA represents parameters specifically for CAA records
type CAA struct {
	Flag  uint8   `json:"caa_flag,string,omitempty"`
	Type  string `json:"caa_type,omitempty"`
	Value string `json:"caa_value,omitempty"`
}

// TLSA represents parameters specifically for TLSA records
type TLSA struct {
	Usage        uint8 `json:"tlsa_usage,string,omitempty"`
	Selector     uint8 `json:"tlsa_selector,string,omitempty"`
	MatchingType uint8 `json:"tlsa_matching_type,string,omitempty"`
}

// WebRedirect represents parameters specifically for web redirect records
type WebRedirect struct {
	MobileMeta   APIBool `json:"mobile_meta"`
	SavePath     APIBool `json:"save_path,omitempty"`
	RedirectType int     `json:"redirect_type,string,omitempty"`

	IsFrame          APIBool `json:"frame,omitempty"`
	FrameTitle       string  `json:"frame_title,omitempty"`
	FrameKeywords    string  `json:"frame_keywords,omitempty"`
	FrameDescription string  `json:"frame_description,omitempty"`
}

// NAPTR represents parameters specifically for NAPTR records
type NAPTR struct {
	Order       uint16 `json:"order,string,omitempty"`
	Preference  uint16 `json:"pref,string,omitempty"`
	Flags       string `json:"flag"`
	Service     string `json:"params"`
	Regexp      string `json:"regexp"`
	Replacement string `json:"replace"`
}

// SOA represents the SOA record of a ClouDNS zone
type SOA struct {
	Serial     int    `json:"serialNumber,string"`
	PrimaryNS  string `json:"primaryNS"`
	AdminMail  string `json:"adminMail"`
	Refresh    int    `json:"refresh,string"`
	Retry      int    `json:"retry,string"`
	Expire     int    `json:"expire,string"`
	DefaultTTL int    `json:"defaultTTL,string"`
}

// RecordsExport represents a BIND zone file export provided by the ClouDNS API
type RecordsExport struct {
	StatusResult
	Zone string `json:"zone"`
}

// DynamicURL represents a DynDNS URL for a specific zone record
type DynamicURL struct {
	Host string `json:"host"`
	URL  string `json:"url"`
}

// GetSOA returns the SOA record of the given zone
// Official Docs: https://www.cloudns.net/wiki/article/62/
func (svc *RecordService) GetSOA(ctx context.Context, zoneName string) (result SOA, err error) {
	params := HTTPParams{"domain-name": zoneName}
	err = svc.api.request(ctx, "POST", recordSOAGetURL, params, nil, &result)
	return
}

// UpdateSOA updates the SOA record of the given zone
// Official Docs: https://www.cloudns.net/wiki/article/63/
func (svc *RecordService) UpdateSOA(ctx context.Context, zoneName string, soa SOA) (result StatusResult, err error) {
	params := soa.AsParams()
	params["domain-name"] = zoneName

	err = svc.api.request(ctx, "POST", recordSOAUpdateURL, params, nil, &result)
	return
}

// List returns all the records of a given zone
// Official Docs: https://www.cloudns.net/wiki/article/57/
func (svc *RecordService) List(ctx context.Context, zoneName string) (result RecordMap, err error) {
	return svc.Search(ctx, zoneName, "", "")
}

// Search returns all records matching a given host and/or record type within the given zone
// Official Docs: https://www.cloudns.net/wiki/article/57/
func (svc *RecordService) Search(ctx context.Context, zoneName, host, recordType string) (result RecordMap, err error) {
	// Build search parameters for record querying
	params := HTTPParams{"domain-name": zoneName}
	if host != "" {
		params["host"] = host
	}
	if recordType != "" {
		params["type"] = recordType
	}

	// Fetch all DNS records with a twist: Unmarshalling to the record map fails if the zone contains no records, as
	// ClouDNS decided to return an empty array instead of a JSON object when no records have been found. In this
	// specific case, we silence the error and return an empty map instead.
	err = svc.api.request(ctx, "POST", recordListURL, params, nil, &result)
	var typeError *json.UnmarshalTypeError
	if errors.As(err, &typeError) && typeError.Value == "array" {
		return make(RecordMap), nil
	}

	return
}

// Create a new record within the given zone
// Official Docs: https://www.cloudns.net/wiki/article/58/
func (svc *RecordService) Create(ctx context.Context, zoneName string, record Record) (result StatusResult, err error) {
	params := record.AsParams()
	params["domain-name"] = zoneName

	err = svc.api.request(ctx, "POST", recordCreateURL, params, nil, &result)
	return
}

// Update modifies a specific record with a given record ID inside the given zone
// Official Docs: https://www.cloudns.net/wiki/article/60/
func (svc *RecordService) Update(ctx context.Context, zoneName string, recordID int, record Record) (result StatusResult, err error) {
	params := record.AsParams()
	params["domain-name"] = zoneName
	params["record-id"] = recordID

	err = svc.api.request(ctx, "POST", recordUpdateURL, params, nil, &result)
	return
}

// Delete modifies a specific record with a given record ID inside the given zone
// Official Docs: https://www.cloudns.net/wiki/article/59/
func (svc *RecordService) Delete(ctx context.Context, zoneName string, recordID int) (result StatusResult, err error) {
	params := HTTPParams{"domain-name": zoneName, "record-id": recordID}
	err = svc.api.request(ctx, "POST", recordDeleteURL, params, nil, &result)
	return
}

// SetActive enables or disables a given record ID within the specified zone
// Official Docs: https://www.cloudns.net/wiki/article/66/
func (svc *RecordService) SetActive(ctx context.Context, zoneName string, recordID int, isActive bool) (result StatusResult, err error) {
	params := HTTPParams{"domain-name": zoneName, "record-id": recordID}
	if isActive {
		params["status"] = 1
	} else {
		params["status"] = 0
	}

	err = svc.api.request(ctx, "POST", recordSetActiveURL, params, nil, &result)
	return
}

// CopyFromZone copies all records from one zone into another, optionally overwriting the existing records
// Official Docs: https://www.cloudns.net/wiki/article/61/
func (svc *RecordService) CopyFromZone(ctx context.Context, targetZoneName, sourceZoneName string, overwrite bool) (result StatusResult, err error) {
	params := HTTPParams{"domain-name": targetZoneName, "from-domain": sourceZoneName}
	if overwrite {
		params["delete-current-records"] = 1
	} else {
		params["delete-current-records"] = 0
	}

	err = svc.api.request(ctx, "POST", recordCopyFromZoneURL, params, nil, &result)
	return
}

// Import records with a specific format into the zone, optionally overwriting the existing records
// Official Docs: https://www.cloudns.net/wiki/article/156/
func (svc *RecordService) Import(ctx context.Context, zoneName string, format RecordFormat, content string, overwrite bool) (result StatusResult, err error) {
	params := HTTPParams{"domain-name": zoneName, "content": content}

	switch format {
	case RecordFormatBIND:
		params["format"] = "bind"
	case RecordFormatTinyDNS:
		params["format"] = "tinydns"
	default:
		return result, ErrIllegalArgument.wrap(errors.New("invalid record format"))
	}

	if overwrite {
		params["delete-existing-records"] = 1
	} else {
		params["delete-existing-records"] = 0
	}

	err = svc.api.request(ctx, "POST", recordImportURL, params, nil, &result)
	return
}

// ImportTransfer imports records from an authoritative nameserver into the zone using AXFR, overwriting all records
// Official Docs: https://www.cloudns.net/wiki/article/65/
func (svc *RecordService) ImportTransfer(ctx context.Context, zoneName, server string) (result StatusResult, err error) {
	params := HTTPParams{"domain-name": zoneName, "server": server}
	err = svc.api.request(ctx, "POST", recordImportTransferURL, params, nil, &result)
	return
}

// Export returns all records of the given zone as a BIND zone file
// Official Docs: https://www.cloudns.net/wiki/article/166/
func (svc *RecordService) Export(ctx context.Context, zoneName string) (result RecordsExport, err error) {
	params := HTTPParams{"domain-name": zoneName}
	err = svc.api.request(ctx, "POST", recordExportURL, params, nil, &result)
	return
}

// GetDynamicURL returns the current DynDNS url for the given record
// Official Docs: https://www.cloudns.net/wiki/article/64/
func (svc *RecordService) GetDynamicURL(ctx context.Context, zoneName string, recordID int) (result DynamicURL, err error) {
	params := HTTPParams{"domain-name": zoneName, "record-id": recordID}
	err = svc.api.request(ctx, "POST", recordGetDynamicURL, params, nil, &result)
	return
}

// ChangeDynamicURL creates or replaces the current DynDNS url for the given record
// Official Docs: https://www.cloudns.net/wiki/article/152/
func (svc *RecordService) ChangeDynamicURL(ctx context.Context, zoneName string, recordID int) (result DynamicURL, err error) {
	params := HTTPParams{"domain-name": zoneName, "record-id": recordID}
	err = svc.api.request(ctx, "POST", recordChangeDynamicURL, params, nil, &result)
	return
}

// DisableDynamicURL disables the current DynDNS url for the given record
// Official Docs: https://www.cloudns.net/wiki/article/152/
func (svc *RecordService) DisableDynamicURL(ctx context.Context, zoneName string, recordID int) (result StatusResult, err error) {
	params := HTTPParams{"domain-name": zoneName, "record-id": recordID}
	err = svc.api.request(ctx, "POST", recordDisableDynamicURL, params, nil, &result)
	return
}

// AvailableTTLs returns the available record TTLs for a specified zone
// Official Docs: https://www.cloudns.net/wiki/article/153/
func (svc *RecordService) AvailableTTLs(ctx context.Context, zoneName string) (result []int, err error) {
	params := HTTPParams{"domain-name": zoneName}
	err = svc.api.request(ctx, "POST", recordAvailableTTLsURL, params, nil, &result)
	return
}

// AvailableRecordTypes returns the available record types for a given zone type and kind
// Official Docs: https://www.cloudns.net/wiki/article/157/
func (svc *RecordService) AvailableRecordTypes(ctx context.Context, zoneType ZoneType, zoneKind ZoneKind) (result []string, err error) {
	params := HTTPParams{}
	isAuthoritative := zoneType == ZoneTypeMaster || zoneType == ZoneTypeGeoDNS
	isParked := zoneType == ZoneTypeParked
	isForward := zoneKind == ZoneKindDomain
	isReverse := zoneKind == ZoneKindIPv4 || zoneKind == ZoneKindIPv6

	switch {
	case isAuthoritative && isForward:
		params["zone-type"] = "domain"
	case isAuthoritative && isReverse:
		params["zone-type"] = "reverse"
	case isParked:
		params["zone-type"] = "parked"
	default:
		return nil, ErrIllegalArgument.wrap(errors.New("unsupported combination of zone type and kind"))
	}

	err = svc.api.request(ctx, "POST", recordAvailableRecordTypesURL, params, nil, &result)
	return
}

// NewRecord instantiates a new record which can be used within ClouDNS API methods. It does -not- add this record
// automatically to any given kind of zone.
func NewRecord(host, recordType, record string, ttl int) Record {
	return Record{
		Host:       host,
		Record:     record,
		RecordType: recordType,
		TTL:        ttl,
		IsActive:   true,
	}
}

// AsParams returns the HTTP parameters for the SOA record for use within the other API methods
func (soa SOA) AsParams() HTTPParams {
	return HTTPParams{
		"primary-ns":  soa.PrimaryNS,
		"admin-mail":  soa.AdminMail,
		"refresh":     soa.Refresh,
		"retry":       soa.Retry,
		"expire":      soa.Expire,
		"default-ttl": soa.DefaultTTL,
	}
}

// AsParams returns the HTTP parameters for a record for use within the other API methods
func (rec Record) AsParams() HTTPParams {
	params := HTTPParams{
		"host":        rec.Host,
		"record":      rec.Record,
		"record-type": rec.RecordType,
		"ttl":         rec.TTL,
	}

	switch rec.RecordType {
	case "MX":
		params["priority"] = rec.Priority
	case "SRV":
		params["priority"] = rec.Priority
		params["weight"] = rec.Weight
		params["port"] = rec.Port
	case "WR":
		isFrame, _ := rec.WebRedirect.IsFrame.MarshalJSON()

		params["save-path"] = rec.WebRedirect.SavePath
		params["redirect-type"] = rec.WebRedirect.RedirectType
		params["frame"] = string(isFrame)
		params["frame-title"] = rec.WebRedirect.FrameTitle
		params["frame-keywords"] = rec.WebRedirect.FrameKeywords
		params["frame-description"] = rec.WebRedirect.FrameDescription
	case "RP":
		params["mail"] = rec.RP.Mail
		params["txt"] = rec.RP.TXT
	case "SSHFP":
		params["algorithm"] = rec.SSHFP.Algorithm
		params["fptype"] = rec.SSHFP.Type
	case "TLSA":
		params["tlsa_usage"] = rec.TLSA.Usage
		params["tlsa_selector"] = rec.TLSA.Selector
		params["tlsa_matching_type"] = rec.TLSA.MatchingType
	case "CAA":
		params["caa_flag"] = rec.CAA.Flag
		params["caa_type"] = rec.CAA.Type
		params["caa_value"] = rec.CAA.Value
	case "NAPTR":
		params["order"] = rec.NAPTR.Order
		params["pref"] = rec.NAPTR.Preference
		params["flag"] = rec.NAPTR.Flags
		params["params"] = rec.NAPTR.Service
		params["regexp"] = rec.NAPTR.Regexp
		params["replace"] = rec.NAPTR.Replacement
	}

	return params
}

// AsSlice converts a RecordMap to a slice of records for easier handling
func (rm RecordMap) AsSlice() []Record {
	results := make([]Record, 0, len(rm))
	for _, value := range rm {
		results = append(results, value)
	}

	return results
}
