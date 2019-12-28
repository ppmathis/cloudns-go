package cloudns

import (
	"context"
	"net"
)

// AccountService is a service object which groups all operations related to ClouDNS account management
type AccountService struct {
	api *Client
}

// Login attempts authentication against the ClouDNS backend with the configured set of credentials.
// Official Docs: https://www.cloudns.net/wiki/article/45/
func (svc *AccountService) Login(ctx context.Context) (result StatusResult, err error) {
	err = svc.api.request(ctx, "POST", "/dns/login.json", nil, nil, &result)
	return
}

// GetCurrentIP returns the IP address which the ClouDNS API backend sees while connecting to it.
// Official Docs: https://www.cloudns.net/wiki/article/307/
func (svc *AccountService) GetCurrentIP(ctx context.Context) (net.IP, error) {
	var result struct {
		IP net.IP `json:"ip"`
	}

	err := svc.api.request(ctx, "POST", "/ip/get-my-ip.json", nil, nil, &result)
	return result.IP, err
}

// GetBalance returns the current account balance / funds for the configured credentials
// Official Docs: https://www.cloudns.net/wiki/article/354/
func (svc *AccountService) GetBalance(ctx context.Context) (float64, error) {
	var result struct {
		Funds float64 `json:"funds,string"`
	}

	err := svc.api.request(ctx, "POST", "/account/get-balance.json", nil, nil, &result)
	return result.Funds, err
}
