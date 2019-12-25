package cloudns

import (
	"context"
	"net"
)

func (api *API) Login(ctx context.Context) error {
	return api.request(ctx, "POST", "/dns/login.json", nil, nil, nil)
}

func (api *API) CurrentIP(ctx context.Context) (net.IP, error) {
	var result struct {
		IP net.IP `json:"ip"`
	}

	err := api.request(ctx, "POST", "/ip/get-my-ip.json", nil, nil, &result)
	return result.IP, err
}

func (api *API) AccountBalance(ctx context.Context) (float64, error) {
	var result struct {
		Funds float64 `json:"funds,string"`
	}

	err := api.request(ctx, "POST", "/account/get-balance.json", nil, nil, &result)
	return result.Funds, err
}
