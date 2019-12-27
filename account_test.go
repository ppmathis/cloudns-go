package cloudns

import (
	"testing"
)

func TestAccountService_Login(t *testing.T) {
	teardown := setup(t)
	defer teardown()

	_, err := client.Account.Login(ctx)
	if err != nil {
		t.Fatalf("Account.Login() returned error: %v", err)
	}
}

func TestAccountService_GetBalance(t *testing.T) {
	teardown := setup(t)
	defer teardown()

	_, err := client.Account.GetBalance(ctx)
	if err != nil {
		t.Fatalf("Account.GetBalance() returned error: %v", err)
	}
}

func TestAccountService_GetCurrentIP(t *testing.T) {
	teardown := setup(t)
	defer teardown()

	_, err := client.Account.GetCurrentIP(ctx)
	if err != nil {
		t.Fatalf("Account.GetCurrentIP() returned error: %v", err)
	}
}
