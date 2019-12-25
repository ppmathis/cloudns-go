package cloudns

import (
	"context"
	"encoding/json"
	"errors"
)

type RecordID int
type RecordType string
type RecordContainer map[RecordID]*Record

type Record struct {
	ID         RecordID   `json:"id,string,omitempty"`
	Host       string     `json:"host"`
	Record     string     `json:"record"`
	RecordType RecordType `json:"record-type"`
	TTL        int        `json:"ttl,string"`
	IsActive   APIBool    `json:"status"`

	Priority int `json:"priority,omitempty"`
	Weight   int `json:"weight,omitempty"`
	Port     int `json:"port,omitempty"`
}

func (api *API) ZoneAvailableRecordTypes(ctx context.Context, zoneType ZoneType, zoneKind ZoneKind) (result []string, err error) {
	params := HttpParams{}
	switch {
	case (zoneType == ZoneMaster || zoneType == ZoneGeoDNS) && (zoneKind == ZoneDomain):
		params["zone-type"] = "domain"
	case (zoneType == ZoneMaster || zoneType == ZoneGeoDNS) && (zoneKind == ZoneIPv4 || zoneKind == ZoneIPv6):
		params["zone-type"] = "reverse"
	case zoneType == ZoneParked:
		params["zone-type"] = "parked"
	default:
		return nil, ErrIllegalArgument.wrap(errors.New("unsupported zone type or kind"))
	}

	err = api.request(ctx, "POST", "/dns/get-available-record-types.json", params, nil, &result)
	return
}

func (api *API) ZoneAvailableTTLs(ctx context.Context, zone string) (result []int, err error) {
	params := HttpParams{"domain-name": zone}
	err = api.request(ctx, "POST", "/dns/get-available-ttl.json", params, nil, &result)
	return
}

func (api *API) ListRecords(ctx context.Context, zone string) (results RecordContainer, err error) {
	params := HttpParams{"domain-name": zone}
	err = api.request(ctx, "POST", "/dns/records.json", params, nil, &results)

	// If unmarshalling failed due to an invalid type, the zone has no records, as for some reason ClouDNS
	// suddenly returns an array instead of a dictionary when a zone is empty.
	var typeError *json.UnmarshalTypeError
	if errors.As(err, &typeError) && typeError.Value == "array" {
		return make(RecordContainer), nil
	}

	return
}

func (api *API) CreateRecord(ctx context.Context, zone string, record *Record) error {
	params, err := record.GetParams()
	if err != nil {
		return err
	}

	params["domain-name"] = zone
	return api.request(ctx, "POST", "/dns/add-record.json", params, nil, nil)
}

func (api *API) UpdateRecord(ctx context.Context, zone string, record *Record) error {
	params, err := record.GetParams()
	if err != nil {
		return err
	}

	params["domain-name"] = zone
	params["record-id"] = record.ID
	return api.request(ctx, "POST", "/dns/mod-record.json", params, nil, nil)
}

func (api *API) DeleteRecord(ctx context.Context, zone string, recordID RecordID) error {
	params := HttpParams{
		"domain-name": zone,
		"record-id":   recordID,
	}
	return api.request(ctx, "POST", "/dns/delete-record.json", params, nil, nil)
}

func NewRecord(recordType RecordType, host, record string, ttl int) *Record {
	return &Record{
		Host:       host,
		Record:     record,
		RecordType: recordType,
		TTL:        ttl,
		IsActive:   true,
	}
}

func (rec *Record) GetParams() (params HttpParams, err error) {
	jsonBody, err := json.Marshal(rec)
	if err != nil {
		return nil, ErrIllegalArgument.wrap(err)
	}

	if err := json.Unmarshal(jsonBody, &params); err != nil {
		return nil, ErrIllegalArgument.wrap(err)
	}

	return
}

func (rc RecordContainer) AsSlice() (records []*Record) {
	records = make([]*Record, 0, len(rc))
	for _, value := range rc {
		records = append(records, value)
	}

	return
}
