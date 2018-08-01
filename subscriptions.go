package licensing

import (
	"context"
	"net/url"

	"github.com/docker/licensing/lib/go-clientlib"
	"github.com/docker/licensing/model"
)

// RequestParams holds request parameters
type RequestParams struct {
	DockerID         string
	PartnerAccountID string
	Origin           string
}

func (c *client) createSubscription(ctx context.Context, request *model.SubscriptionCreationRequest) (response *model.SubscriptionDetail, err error) {
	url := c.baseURI
	url.Path += "/api/billing/v4/subscriptions"
	response = new(model.SubscriptionDetail)
	_, _, err = c.doReq(ctx, "POST", &url, clientlib.SendJSON(request), clientlib.RecvJSON(response))
	return
}

func (c *client) getSubscription(ctx context.Context, id string) (response *model.SubscriptionDetail, err error) {
	url := c.baseURI
	url.Path += "/api/billing/v4/subscriptions/" + id
	response = new(model.SubscriptionDetail)
	_, _, err = c.doReq(ctx, "GET", &url, clientlib.RecvJSON(response))
	return
}

func (c *client) listSubscriptions(ctx context.Context, params map[string]string) (response []*model.SubscriptionDetail, err error) {
	values := url.Values{}
	values.Set("docker_id", params["docker_id"])
	values.Set("partner_account_id", params["partner_account_id"])
	values.Set("origin", params["origin"])

	url := c.baseURI
	url.Path += "/api/billing/v4/subscriptions"
	url.RawQuery = values.Encode()

	response = make([]*model.SubscriptionDetail, 0)
	_, _, err = c.doReq(ctx, "GET", &url, clientlib.RecvJSON(&response))
	return
}
