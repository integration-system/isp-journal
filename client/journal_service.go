package client

import (
	"github.com/integration-system/isp-journal/search"
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

func (c *journalServiceClient) Search(request search.SearchRequest) ([]search.SearchResponse, error) {
	response := make([]search.SearchResponse, 0)
	if err := c.client.Post(c.gateHost+searchMethod, &request, &response); err != nil {
		return nil, err
	} else {
		return response, nil
	}
}

func (c *journalServiceClient) SearchWithCursor(request search.SearchWithCursorRequest) (*search.SearchWithCursorResponse, error) {
	response := new(search.SearchWithCursorResponse)
	if err := c.client.Post(c.gateHost+searchWithCursorMethod, &request, response); err != nil {
		return nil, err
	} else {
		return response, nil
	}
}
