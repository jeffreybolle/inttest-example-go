package creditscore

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

type CreditStore struct {
	url    string
	client http.Client
}

func NewCreditScore(url string) *CreditStore {
	return &CreditStore{
		url: url,
		client: http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

type creditScoreReq struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type creditScoreResp struct {
	Score string `json:"score"`
}

func (cs *CreditStore) GetScore(ctx context.Context, firstName, lastName string) (float64, error) {
	csreq := creditScoreReq{
		FirstName: firstName,
		LastName:  lastName,
	}
	body, err := json.Marshal(&csreq)
	if err != nil {
		return 0, fmt.Errorf("error while encoding to JSON: %v", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cs.url, bytes.NewReader(body))
	if err != nil {
		return 0, fmt.Errorf("error while creating HTTP request: %v", err)
	}
	resp, err := cs.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("error while doing HTTP request: %v", err)
	}
	defer resp.Body.Close()
	var csresp creditScoreResp
	err = json.NewDecoder(resp.Body).Decode(&csresp)
	if err != nil {
		return 0, fmt.Errorf("error while decoding to JSON: %v", err)
	}
	score, err := strconv.ParseFloat(csresp.Score, 64)
	if err != nil {
		return 0, fmt.Errorf("error while converting string -> float64: %v", err)
	}
	return score, nil
}
