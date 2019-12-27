package cloudns

import (
	"math/rand"
	"testing"
	"time"
)

var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))
var randomCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

const testDomain string = "api-example.com"
const testTTL int = 3600

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
