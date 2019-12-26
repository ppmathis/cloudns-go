package cloudns

import (
	"context"
	"encoding/json"
	"errors"
)

const recordAvailableTTLsURL = "/dns/get-available-ttl.json"
const recordAvailableRecordTypesURL = "/dns/get-available-record-types.json"
const recordListURL = "/dns/records.json"
const recordCreateURL = "/dns/add-record.json"
const recordUpdateURL = "/dns/mod-record.json"
const recordDeleteURL = "/dns/delete-record.json"

type recordService struct {
	api *API
}

type RecordMap map[int]Record
type Record struct {
	ID         int     `json:"id,string,omitempty"`
	Host       string  `json:"host"`
	Record     string  `json:"record"`
	RecordType string  `json:"record-type"`
	TTL        int     `json:"ttl,string"`
	IsActive   APIBool `json:"status"`

	Priority int `json:"priority,omitempty"`
	Weight   int `json:"weight,omitempty"`
	Port     int `json:"port,omitempty"`
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
	params, err := record.AsParams()
	if err != nil {
		return
	}

	params["domain-name"] = zoneName
	err = svc.api.request(ctx, "POST", recordCreateURL, params, nil, &result)
	return
}

func (svc *recordService) Update(ctx context.Context, zoneName string, recordID int, record Record) (result BaseResult, err error) {
	params, err := record.AsParams()
	if err != nil {
		return
	}

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

func (rec Record) AsParams() (params HttpParams, err error) {
	jsonBody, err := json.Marshal(rec)
	if err != nil {
		return nil, ErrIllegalArgument.wrap(err)
	}

	if err := json.Unmarshal(jsonBody, &params); err != nil {
		return nil, ErrIllegalArgument.wrap(err)
	}

	return
}

func (rm RecordMap) AsSlice() []Record {
	results := make([]Record, 0, len(rm))
	for _, value := range rm {
		results = append(results, value)
	}

	return results
}
