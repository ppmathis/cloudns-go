package cloudns

import (
	"errors"
	"math/rand"
	"net/url"
	"testing"
	"time"
)

var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))
var randomCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func createRandomString(length int) string {
	result := make([]byte, length)
	for i := range result {
		result[i] = randomCharset[seededRand.Intn(len(randomCharset))]
	}
	return string(result)
}

func createTestRecord(t *testing.T) Record {
	recordName := createRandomString(16)
	record := NewRecord(recordName, "A", "127.0.0.1", testTTL)

	_, err := client.Records.Create(ctx, testDomain, record)
	if err != nil {
		t.Fatalf("could not create test record: %v", err)
	}

	records, err := client.Records.Search(ctx, testDomain, recordName, "A")
	if err != nil {
		t.Fatalf("could not list records for determining test record: %v", err)
	}

	if len(records) < 1 {
		t.Fatalf("could not find created test record")
	}

	return records.AsSlice()[0]
}

func setupWithRecord(t *testing.T) (Record, func()) {
	baseTeardown := setup(t)
	testRecord := createTestRecord(t)

	return testRecord, func() {
		_, _ = client.Records.Delete(ctx, testDomain, testRecord.ID)
		baseTeardown()
	}
}

func TestRecordService_GetSOA(t *testing.T) {
	teardown := setup(t)
	defer teardown()

	_, err := client.Records.GetSOA(ctx, testDomain)
	if err != nil {
		t.Fatalf("Records.GetSOA() returned error: %v", err)
	}
}

func TestRecordService_UpdateSOA(t *testing.T) {
	teardown := setup(t)
	defer teardown()

	oldSOA, err := client.Records.GetSOA(ctx, testDomain)
	if err != nil {
		t.Fatalf("Records.GetSOA() returned error: %v", err)
	}

	newSOA := oldSOA
	newSOA.AdminMail = createRandomString(16) + "@" + testDomain
	_, err = client.Records.UpdateSOA(ctx, testDomain, newSOA)
	if err != nil {
		t.Fatalf("Records.UpdateSOA() returned error: %v", err)
	}
}

func TestRecordService_List(t *testing.T) {
	_, teardown := setupWithRecord(t)
	defer teardown()

	records, err := client.Records.List(ctx, testDomain)
	if err != nil {
		t.Fatalf("Records.List() returned error: %v", err)
	}
	if len(records) < 1 {
		t.Fatalf("Records.List() returned %d elements, expected >= 1", len(records))
	}
}

func TestRecordService_Create(t *testing.T) {
	teardown := setup(t)
	defer teardown()

	record := NewRecord("localhost", "A", "127.0.0.1", testTTL)
	_, err := client.Records.Create(ctx, testDomain, record)
	if err != nil {
		t.Fatalf("Records.Create() returned error: %v", err)
	}
}

func TestRecordService_Update(t *testing.T) {
	record, teardown := setupWithRecord(t)
	defer teardown()

	record.Record = "127.0.0.2"
	_, err := client.Records.Update(ctx, testDomain, record.ID, record)
	if err != nil {
		t.Fatalf("Records.Update() returned error: %v", err)
	}
}

func TestRecordService_Delete(t *testing.T) {
	record, teardown := setupWithRecord(t)
	defer teardown()

	_, err := client.Records.Delete(ctx, testDomain, record.ID)
	if err != nil {
		t.Fatalf("Records.Delete() returned error: %v", err)
	}
}

func TestRecordService_SetActive(t *testing.T) {
	record, teardown := setupWithRecord(t)
	defer teardown()

	_, err := client.Records.SetActive(ctx, testDomain, record.ID, false)
	if err != nil {
		t.Fatalf("Records.SetActive() returned error: %v", err)
	}
}

func TestRecordService_Import_BIND(t *testing.T) {
	teardown := setup(t)
	defer teardown()

	_, err := client.Records.Import(ctx, testDomain, RecordFormatBIND, "@ 3600 IN A 1.2.3.4", false)
	if err != nil {
		t.Fatalf("Records.Import() returned error: %v", err)
	}
}

func TestRecordService_Import_Overwrite_BIND(t *testing.T) {
	_, teardown := setupWithRecord(t)
	defer teardown()

	_, err := client.Records.Import(ctx, testDomain, RecordFormatBIND, "@ 3600 IN A 1.2.3.4", true)
	if err != nil {
		t.Fatalf("Records.Import() returned error: %v", err)
	}

	records, err := client.Records.List(ctx, testDomain)
	if err != nil {
		t.Fatalf("Records.List() returned error: %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("Expected single result from Records.List(), got: %d", len(records))
	}
}

func TestRecordService_Import_TinyDNS(t *testing.T) {
	teardown := setup(t)
	defer teardown()

	_, err := client.Records.Import(ctx, testDomain, RecordFormatTinyDNS, "=:1.2.3.4", false)
	if err != nil {
		t.Fatalf("Records.Import() returned error: %v", err)
	}
}

func TestRecordService_Import_Overwrite_TinyDNS(t *testing.T) {
	_, teardown := setupWithRecord(t)
	defer teardown()

	_, err := client.Records.Import(ctx, testDomain, RecordFormatTinyDNS, "=:1.2.3.4", true)
	if err != nil {
		t.Fatalf("Records.Import() returned error: %v", err)
	}

	records, err := client.Records.List(ctx, testDomain)
	if err != nil {
		t.Fatalf("Records.List() returned error: %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("Expected single result from Records.List(), got: %d", len(records))
	}
}

func TestRecordService_Import_Invalid(t *testing.T) {
	teardown := setup(t)
	defer teardown()

	_, err := client.Records.Import(ctx, testDomain, -1, "", false)
	if err == nil || !errors.Is(err, ErrIllegalArgument) {
		t.Fatalf("Expected ErrIllegalArgument from Records.Import() with invalid format, got: %v", err)
	}
}

func TestRecordService_Export(t *testing.T) {
	teardown := setup(t)
	defer teardown()

	_, err := client.Records.Export(ctx, testDomain)
	if err != nil {
		t.Fatalf("Records.Export() returned error: %v", err)
	}
}

func TestRecordService_GetDynamicURL(t *testing.T) {
	record, teardown := setupWithRecord(t)
	defer teardown()

	result, err := client.Records.GetDynamicURL(ctx, testDomain, record.ID)
	if err != nil {
		t.Fatalf("Records.GetDynamicURL() returned error: %v", err)
	}

	expectedHost := record.Host + "." + testDomain
	if result.Host != expectedHost {
		t.Fatalf("Records.GetDynamicURL() returned host [%s], expected [%s]", result.Host, expectedHost)
	}
	if _, err := url.Parse(result.URL); err != nil {
		t.Fatalf("could not parse URL from Records.GetDynamicURL(): %s", result.URL)
	}
}

func TestRecordService_ChangeDynamicURL(t *testing.T) {
	record, teardown := setupWithRecord(t)
	defer teardown()

	oldResult, err := client.Records.GetDynamicURL(ctx, testDomain, record.ID)
	if err != nil {
		t.Fatalf("Records.GetDynamicURL() returned error: %v", err)
	}

	newResult, err := client.Records.ChangeDynamicURL(ctx, testDomain, record.ID)
	if err != nil {
		t.Fatalf("Records.ChangeDynamicURL() returned error: %v", err)
	}

	if oldResult.Host != newResult.Host {
		t.Fatalf("Unexpected host diffence, got [%s], expected [%s]", oldResult.Host, newResult.Host)
	}

	if oldResult.URL == newResult.URL {
		t.Fatalf("URL has not changed after Records.ChangeDynamicURL(), still got [%s]", newResult.URL)
	}
}

func TestRecordService_DisableDynamicURL(t *testing.T) {
	record, teardown := setupWithRecord(t)
	defer teardown()

	_, err := client.Records.DisableDynamicURL(ctx, testDomain, record.ID)
	if err != nil {
		t.Fatalf("Records.DisableDynamicURL() returned error: %v", err)
	}
}

func TestRecordService_AvailableTTLs(t *testing.T) {
	teardown := setup(t)
	defer teardown()

	ttlValues, err := client.Records.AvailableTTLs(ctx, testDomain)
	if err != nil {
		t.Fatalf("Records.AvailableTTLs() returned error: %v", err)
	}
	if len(ttlValues) < 1 {
		t.Fatalf("Expected at least one result from Records.AvailableTTLs()")
	}
}

func TestRecordService_AvailableRecordTypes_Valid(t *testing.T) {
	teardown := setup(t)
	defer teardown()

	test := func(zt ZoneType, zk ZoneKind) {
		recordTypes, err := client.Records.AvailableRecordTypes(ctx, zt, zk)
		if err != nil {
			t.Fatalf("Records.AvailableRecordTypes(%d, %d) returned error: %v", zt, zk, err)
		}
		if len(recordTypes) < 1 {
			t.Fatalf("Expected at least one result from Records.AvailableTTLs(%d, %d)", zt, zk)
		}
	}

	test(ZoneTypeMaster, ZoneKindDomain)
	test(ZoneTypeMaster, ZoneKindIPv4)
	test(ZoneTypeMaster, ZoneKindIPv6)
	test(ZoneTypeGeoDNS, ZoneKindDomain)
	test(ZoneTypeParked, ZoneKindDomain)
}

func TestRecordService_AvailableRecordTypes_Invalid(t *testing.T) {
	teardown := setup(t)
	defer teardown()

	_, err := client.Records.AvailableRecordTypes(ctx, ZoneTypeSlave, ZoneKindDomain)
	if err == nil || !errors.Is(err, ErrIllegalArgument) {
		t.Fatalf("Expected ErrIllegalArgument from Records.AvailableRecordTypes() with invalid zone type/kind, got: %v", err)
	}
}
