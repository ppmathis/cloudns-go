package cloudns

import (
	"errors"
	"fmt"
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

func TestRecordService_RecordTypes(t *testing.T) {
	teardown := setup(t)
	defer teardown()

	testRecordType := func(recordType string, initial, updated Record) {
		var err error

		// Override test-specific values for initial and updated record
		initial.RecordType = recordType
		initial.TTL = 60
		if initial.Host == "" {
			initial.Host = "%s"
		}
		initial.Host = fmt.Sprintf(initial.Host, strings.ToLower(createRandomString(16)))

		updated.RecordType = initial.RecordType
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

	testRecordType("A",
		Record{Record: "192.0.2.100"},
		Record{Record: "192.0.2.200"},
	)
	testRecordType("AAAA",
		Record{Record: "2001:db8::100"},
		Record{Record: "2001:db8::200"},
	)
	testRecordType("MX",
		Record{Record: "mx1.local", Priority: 100},
		Record{Record: "mx2.local", Priority: 200},
	)
	testRecordType("CNAME",
		Record{Record: "server1.local"},
		Record{Record: "server2.local"},
	)
	testRecordType("TXT",
		Record{Record: "Hello"},
		Record{Record: "World"},
	)
	testRecordType("NS",
		Record{Record: "ns1.local"},
		Record{Record: "ns2.local"},
	)
	testRecordType("SRV",
		Record{Host: "_%s._tcp", Record: "srv1.local", Priority: 10, Weight: 20, Port: 30},
		Record{Record: "srv2.local", Priority: 40, Weight: 50, Port: 60},
	)
	testRecordType("WR",
		Record{Record: "http://www1.local", WebRedirect: WebRedirect{
			IsFrame: true, FrameTitle: "T", FrameKeywords: "K", FrameDescription: "D",
		}},
		Record{Record: "http://www2.local", WebRedirect: WebRedirect{
			IsFrame: false, SavePath: true, RedirectType: 302,
		}},
	)
	testRecordType("ALIAS",
		Record{Record: "www1.local"},
		Record{Record: "www2.local"},
	)
	testRecordType("RP",
		Record{RP: RP{Mail: "admin1@mail.local", TXT: "txt1.local"}},
		Record{RP: RP{Mail: "admin2@mail.local", TXT: "txt2.local"}},
	)
	testRecordType("SSHFP",
		Record{Record: "4fca1fe60ec4fca4053504f4fcab0d5d7c99bd0f", SSHFP: SSHFP{
			Algorithm: 1, Type: 1,
		}},
		Record{Record: "1357acf64348f3f7bd0942ba75878ebd3a75af979007f059741d29f95c4a0b80", SSHFP: SSHFP{
			Algorithm: 3, Type: 2,
		}},
	)
	testRecordType("NAPTR",
		Record{NAPTR: NAPTR{
			Order: 10, Preference: 20, Flags: "U", Service: "svc1.local", Regexp: "Hello",
		}},
		Record{NAPTR: NAPTR{
			Order: 30, Preference: 40, Flags: "S", Service: "svc2.local", Replacement: "World",
		}},
	)
	testRecordType("CAA",
		Record{CAA: CAA{
			Flag: 0, Type: "issue", Value: "ca1.local",
		}},
		Record{CAA: CAA{
			Flag: 128, Type: "issuewild", Value: "ca2.local",
		}},
	)
	testRecordType("TLSA",
		Record{Host: "_443._tcp.%s", Record: "53472ce4477a2b2f17085ff9f0b07ecad2091d25a9ec02dda622c741e62c75e4ee8dea577db822caa32935a86e52827c51bf000a65506d4f760dab7712ed9a25", TLSA: TLSA{
			Usage: 0, Selector: 1, MatchingType: 2,
		}},
		Record{Record: "078a656e3670499c991bb0274682058af7bdc05fc462c605f0f8958179816cd7", TLSA: TLSA{
			Usage: 2, Selector: 0, MatchingType: 1,
		}},
	)
}
