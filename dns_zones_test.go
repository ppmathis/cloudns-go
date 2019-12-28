package cloudns

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestZoneService_AvailableNameservers(t *testing.T) {
	teardown := setup(t)
	defer teardown()

	nameservers, err := client.Zones.AvailableNameservers(ctx)
	assert.NoError(t, err, "should not fail")
	assert.NotEmpty(t, nameservers, "should return at least one nameserver")
}

func TestZoneService_List(t *testing.T) {
	teardown := setup(t)
	defer teardown()

	zones, err := client.Zones.List(ctx)
	assert.NoError(t, err, "should not fail")
	assert.NotEmpty(t, zones, "should return at least one zone")
}

func TestZoneService_Search(t *testing.T) {
	teardown := setup(t)
	defer teardown()

	zones, err := client.Zones.Search(ctx, testDomain, 0)
	assert.NoError(t, err, "should not fail")
	assert.Len(t, zones, 1, "should return exactly one zone")
	assert.Equal(t, testDomain, zones[0].Name, "first result should match the test zone")
}

func TestZoneService_SetActive(t *testing.T) {
	var err error

	teardown := setup(t)
	defer teardown()

	_, err = client.Zones.SetActive(ctx, testDomain, false)
	assert.NoError(t, err, "disabling test zone should not fail")
	_, err = client.Zones.SetActive(ctx, testDomain, true)
	assert.NoError(t, err, "enabling test zone should not fail")
}

func TestZoneService_IsUpdated(t *testing.T) {
	teardown := setup(t)
	defer teardown()

	_, err := client.Zones.IsUpdated(ctx, testDomain)
	assert.NoError(t, err, "should not fail")
}

func TestZoneService_TriggerUpdate(t *testing.T) {
	teardown := setup(t)
	defer teardown()

	_, err := client.Zones.TriggerUpdate(ctx, testDomain)
	assert.NoError(t, err, "TriggerUpdate() should not fail")

	isUpdated, err := client.Zones.IsUpdated(ctx, testDomain)
	assert.NoError(t, err, "IsUpdated() should not fail")
	assert.False(t, isUpdated, "zone update status should be false due to manual trigger")
}

func TestZoneService_GetUpdateStatus(t *testing.T) {
	teardown := setup(t)
	defer teardown()

	updateStatus, err := client.Zones.GetUpdateStatus(ctx, testDomain)
	assert.NoError(t, err, "should not fail")
	assert.NotEmpty(t, updateStatus, "should contain at least one result")
}

func TestZoneService_Get(t *testing.T) {
	teardown := setup(t)
	defer teardown()

	zone, err := client.Zones.Get(ctx, testDomain)
	assert.NoError(t, err, "should not fail")
	assert.Equal(t, testDomain, zone.Name, "zone name of result should match test zone")
}

func TestZoneService_GetUsage(t *testing.T) {
	teardown := setup(t)
	defer teardown()

	_, err := client.Zones.GetUsage(ctx)
	assert.NoError(t, err, "should not fail")
}
