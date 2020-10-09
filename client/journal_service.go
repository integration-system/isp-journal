package client

import (
	"github.com/integration-system/isp-lib/v2/http"
)

const (
	scheme                 = "http://"
	searchMethod           = "/api/journal/log/search"
	searchWithCursorMethod = "/api/journal/log/search_with_cursor"
)

func NewJournalServiceClient(restClient http.RestClient) *journalServiceClient {
	return &journalServiceClient{client: restClient}
}

type journalServiceClient struct {
	client   http.RestClient
	gateHost string
}

func (c *journalServiceClient) ReceiveConfiguration(gateHost string) {
	c.gateHost = scheme + gateHost
}
