package licensing

import (
	"context"

	"github.com/docker/licensing/lib/go-clientlib"
	"github.com/docker/licensing/model"
)

func (c *client) createAccount(ctx context.Context, dockerID string, request *model.AccountCreationRequest) (response *model.Account, err error) {
	url := c.baseURI
	url.Path += "/api/billing/v4/accounts/" + dockerID

	response = new(model.Account)
	_, _, err = c.doReq(ctx, "PUT", &url, clientlib.SendJSON(request), clientlib.RecvJSON(response))
	if err != nil {
		return nil, err
	}

	return
}

func (c *client) getAccount(ctx context.Context, dockerID string) (response *model.Account, err error) {
	url := c.baseURI
	url.Path += "/api/billing/v4/accounts/" + dockerID

	response = new(model.Account)
	_, _, err = c.doReq(ctx, "GET", &url, clientlib.RecvJSON(response))
	if err != nil {
		return nil, err
	}

	return
}
