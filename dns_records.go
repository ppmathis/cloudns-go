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

type RecordFormat int

const (
	RecordFormatBIND RecordFormat = iota
	RecordFormatTinyDNS
)

type recordService struct {
	api *API
}

type RecordMap map[int]Record
type Record struct {
	ID         int     `json:"id,string,omitempty"`
	Host       string  `json:"host"`
	Record     string  `json:"record"`
	RecordType string  `json:"type"`
	TTL        int     `json:"ttl,string"`
	IsActive   APIBool `json:"status"`

	Priority int `json:"priority,omitempty"`
	Weight   int `json:"weight,omitempty"`
	Port     int `json:"port,omitempty"`
}

type SOA struct {
	Serial     int    `json:"serialNumber,string"`
	PrimaryNS  string `json:"primaryNS"`
	AdminMail  string `json:"adminMail"`
	Refresh    int    `json:"refresh,string"`
	Retry      int    `json:"retry,string"`
	Expire     int    `json:"expire,string"`
	DefaultTTL int    `json:"defaultTTL,string"`
}

type RecordsExport struct {
	BaseResult
	Zone string `json:"zone"`
}

type DynamicURL struct {
	Host string `json:"host"`
	URL  string `json:"url"`
}

func (svc *recordService) GetSOA(ctx context.Context, zoneName string) (result SOA, err error) {
	params := HttpParams{"domain-name": zoneName}
	err = svc.api.request(ctx, "POST", recordSOAGetURL, params, nil, &result)
	return
}

func (svc *recordService) UpdateSOA(ctx context.Context, zoneName string, soa SOA) (result BaseResult, err error) {
	params := soa.AsParams()
	params["domain-name"] = zoneName

	err = svc.api.request(ctx, "POST", recordSOAUpdateURL, params, nil, &result)
	return
}

func (svc *recordService) List(ctx context.Context, zoneName string) (result RecordMap, err error) {
	return svc.Search(ctx, zoneName, "", "")
}

func (svc *recordService) Search(ctx context.Context, zoneName, host, recordType string) (result RecordMap, err error) {
	// Build search parameters for record querying
	params := HttpParams{"domain-name": zoneName}
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

func (svc *recordService) Create(ctx context.Context, zoneName string, record Record) (result BaseResult, err error) {
	params := record.AsParams()
	params["domain-name"] = zoneName

	err = svc.api.request(ctx, "POST", recordCreateURL, params, nil, &result)
	return
}

func (svc *recordService) Update(ctx context.Context, zoneName string, recordID int, record Record) (result BaseResult, err error) {
	params := record.AsParams()
	params["domain-name"] = zoneName
	params["record-id"] = recordID

	err = svc.api.request(ctx, "POST", recordUpdateURL, params, nil, &result)
	return
}

func (svc *recordService) Delete(ctx context.Context, zoneName string, recordID int) (result BaseResult, err error) {
	params := HttpParams{"domain-name": zoneName, "record-id": recordID}
	err = svc.api.request(ctx, "POST", recordDeleteURL, params, nil, &result)
	return
}

func (svc *recordService) SetActive(ctx context.Context, zoneName string, recordID int, isActive bool) (result BaseResult, err error) {
	params := HttpParams{"domain-name": zoneName, "record-id": recordID}
	if isActive {
		params["status"] = 1
	} else {
		params["status"] = 0
	}

	err = svc.api.request(ctx, "POST", recordSetActiveURL, params, nil, &result)
	return
}

func (svc *recordService) CopyFromZone(ctx context.Context, zoneName, sourceZoneName string, overwrite bool) (result BaseResult, err error) {
	params := HttpParams{"domain-name": zoneName, "from-domain": sourceZoneName}
	if overwrite {
		params["delete-current-records"] = 1
	} else {
		params["delete-current-records"] = 0
	}

	err = svc.api.request(ctx, "POST", recordCopyFromZoneURL, params, nil, &result)
	return
}

func (svc *recordService) Import(ctx context.Context, zoneName string, format RecordFormat, content string, overwrite bool) (result BaseResult, err error) {
	params := HttpParams{"domain-name": zoneName, "content": content}

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

func (svc *recordService) ImportTransfer(ctx context.Context, zoneName, server string) (result BaseResult, err error) {
	params := HttpParams{"domain-name": zoneName, "server": server}
	err = svc.api.request(ctx, "POST", recordImportTransferURL, params, nil, &result)
	return
}

func (svc *recordService) Export(ctx context.Context, zoneName string) (result RecordsExport, err error) {
	params := HttpParams{"domain-name": zoneName}
	err = svc.api.request(ctx, "POST", recordExportURL, params, nil, &result)
	return
}

func (svc *recordService) GetDynamicURL(ctx context.Context, zoneName string, recordID int) (result DynamicURL, err error) {
	params := HttpParams{"domain-name": zoneName, "record-id": recordID}
	err = svc.api.request(ctx, "POST", recordGetDynamicURL, params, nil, &result)
	return
}

func (svc *recordService) ChangeDynamicURL(ctx context.Context, zoneName string, recordID int) (result DynamicURL, err error) {
	params := HttpParams{"domain-name": zoneName, "record-id": recordID}
	err = svc.api.request(ctx, "POST", recordChangeDynamicURL, params, nil, &result)
	return
}

func (svc *recordService) DisableDynamicURL(ctx context.Context, zoneName string, recordID int) (result BaseResult, err error) {
	params := HttpParams{"domain-name": zoneName, "record-id": recordID}
	err = svc.api.request(ctx, "POST", recordDisableDynamicURL, params, nil, &result)
	return
}

func (svc *recordService) AvailableTTLs(ctx context.Context, zoneName string) (result []int, err error) {
	params := HttpParams{"domain-name": zoneName}
	err = svc.api.request(ctx, "POST", recordAvailableTTLsURL, params, nil, &result)
	return
}

func (svc *recordService) AvailableRecordTypes(ctx context.Context, zoneType ZoneType, zoneKind ZoneKind) (result []string, err error) {
	params := HttpParams{}
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

func NewRecord(host, recordType, record string, ttl int) Record {
	return Record{
		Host:       host,
		Record:     record,
		RecordType: recordType,
		TTL:        ttl,
		IsActive:   true,
	}
}

func (soa SOA) AsParams() HttpParams {
	return HttpParams{
		"primary-ns":  soa.PrimaryNS,
		"admin-mail":  soa.AdminMail,
		"refresh":     soa.Refresh,
		"retry":       soa.Retry,
		"expire":      soa.Expire,
		"default-ttl": soa.DefaultTTL,
	}
}

func (rec Record) AsParams() HttpParams {
	return HttpParams{
		"host":        rec.Host,
		"record":      rec.Record,
		"record-type": rec.RecordType,
		"ttl":         rec.TTL,
		"priority":    rec.Priority,
		"weight":      rec.Weight,
		"port":        rec.Port,
	}
}

func (rm RecordMap) AsSlice() []Record {
	results := make([]Record, 0, len(rm))
	for _, value := range rm {
		results = append(results, value)
	}

	return results
}
