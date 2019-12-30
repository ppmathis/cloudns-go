package cloudns

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"net/url"
	"strings"
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
	record := NewRecord(RecordTypeA, recordName, "127.0.0.1", testTTL)

	_, err := client.Records.Create(ctx, testDomain, record)
	if err != nil {
		t.Fatalf("could not create test record: %v", err)
	}

	records, err := client.Records.Search(ctx, testDomain, recordName, RecordTypeA)
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

	record := NewRecord(RecordTypeA, "localhost", "127.0.0.1", testTTL)
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

func TestRecordService_RecordTypes(t *testing.T) {
	teardown := setup(t)
	defer teardown()

	testRecordUpdate := func(initial, updated Record) {
		var err error

		// Override test-specific values for initial and updated record
		initial.TTL = 60
		initial.Host = initial.Host + strings.ToLower(createRandomString(16))
		updated.Host = initial.Host
		updated.TTL = 300

		// Create the test record using the ClouDNS API
		_, err = client.Records.Create(ctx, testDomain, initial)
		assert.NoError(t, err, "creating test record should not fail")

		// Search for the created test record
		initialRecords, err := client.Records.Search(ctx, testDomain, initial.Host, initial.RecordType)
		assert.NoError(t, err, "searching for initial test record should not fail")
		assert.Len(t, initialRecords, 1, "result should contain a single record")
		initialRecord := initialRecords.AsSlice()[0]

		// Ensure to cleanup test record when test ends
		defer func() {
			_, err := client.Records.Delete(ctx, testDomain, initialRecord.ID)
			assert.NoError(t, err, "deleting test record should not fail during cleanup")
		}()

		// Update the provided initial and record structs with API-provided values
		initial.ID = initialRecord.ID
		initial.IsActive = initialRecord.IsActive
		initial.Host = initialRecord.Host
		updated.ID = initialRecord.ID
		updated.IsActive = initialRecord.IsActive
		updated.Host = initialRecord.Host

		// Ensure the current record matches the provided initial record
		assert.EqualValues(t, initial, initialRecord, "created test record should match provided initial data")

		// Update the record using the provided struct
		_, err = client.Records.Update(ctx, testDomain, initialRecord.ID, updated)
		assert.NoError(t, err, "updating test record should not fail")

		// Search for the updated test record
		updatedRecords, err := client.Records.Search(ctx, testDomain, updated.Host, updated.RecordType)
		assert.NoError(t, err, "searching for updated test record should not fail")
		assert.Len(t, updatedRecords, 1, "result should contain a single record")
		updatedRecord := updatedRecords.AsSlice()[0]

		// Ensure the current record matches the provided updated record
		assert.EqualValues(t, initialRecord.ID, updatedRecord.ID, "updated test record should have same ID as initial test record")
		assert.EqualValues(t, updated, updatedRecord, "updated test record should match provided update data")
	}

	testRecordUpdate(
		NewRecordA("", "192.0.2.100", 0),
		NewRecordA("", "192.0.2.200", 0),
	)
	testRecordUpdate(
		NewRecordAAAA("", "2001:db8::100", 0),
		NewRecordAAAA("", "2001:db8::200", 0),
	)
	testRecordUpdate(
		NewRecordMX("", 100, "mx1.local", 0),
		NewRecordMX("", 200, "mx2.local", 0),
	)
	testRecordUpdate(
		NewRecordCNAME("", "server1.local", 0),
		NewRecordCNAME("", "server2.local", 0),
	)
	testRecordUpdate(
		NewRecordTXT("", "Hello", 0),
		NewRecordTXT("", "World", 0),
	)
	testRecordUpdate(
		NewRecordNS("", "ns1.local", 0),
		NewRecordNS("", "ns2.local", 0),
	)
	testRecordUpdate(
		NewRecordSRV("_test._tcp.", 10, 20, 30, "srv1.local", 0),
		NewRecordSRV("_test._tcp.", 40, 50, 60, "srv2.local", 0),
	)
	testRecordUpdate(
		NewRecordWebRedirect("", "http://www1.local", WebRedirect{
			IsFrame: true, FrameTitle: "T", FrameKeywords: "K", FrameDescription: "D",
		}, 0),
		NewRecordWebRedirect("", "http://www2.local", WebRedirect{
			IsFrame: false, SavePath: true, RedirectType: 302,
		}, 0),
	)
	testRecordUpdate(
		NewRecordALIAS("", "www1.local", 0),
		NewRecordALIAS("", "www2.local", 0),
	)
	testRecordUpdate(
		NewRecordRP("", "admin1@mail.local", "txt1.local", 0),
		NewRecordRP("", "admin2@mail.local", "txt2.local", 0),
	)
	testRecordUpdate(
		NewRecordSSHFP("", 1, 1, "4fca1fe60ec4fca4053504f4fcab0d5d7c99bd0f", 0),
		NewRecordSSHFP("", 3, 2, "1357acf64348f3f7bd0942ba75878ebd3a75af979007f059741d29f95c4a0b80", 0),
	)
	testRecordUpdate(
		NewRecordNAPTR("", 10, 20, "U", "svc1.local", "Hello", "", 0),
		NewRecordNAPTR("", 30, 40, "S", "svc2.local", "", "World", 0),
	)
	testRecordUpdate(
		NewRecordCAA("", 0, "issue", "ca1.local", 0),
		NewRecordCAA("", 128, "issuewild", "ca2.local", 0),
	)
	testRecordUpdate(
		NewRecordTLSA("_443._tcp.", 0, 1, 2, "53472ce4477a2b2f17085ff9f0b07ecad2091d25a9ec02dda622c741e62c75e4ee8dea577db822caa32935a86e52827c51bf000a65506d4f760dab7712ed9a25", 0),
		NewRecordTLSA("_443._tcp.", 2, 0, 1, "078a656e3670499c991bb0274682058af7bdc05fc462c605f0f8958179816cd7", 0),
	)
}
