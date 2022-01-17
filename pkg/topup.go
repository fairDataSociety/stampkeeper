/*
MIT License

Copyright (c) 2021 Fair Data Society

*/
package pkg

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"math/big"
	"net/http"
	"time"
)

type Topup struct {
	ctx    context.Context
	cancel context.CancelFunc

	batchId         string
	interval        time.Duration
	url             string
	balanceEndpoint string
	name            string
	minAmount       *big.Int
	topupAmount     *big.Int
	active          bool
	logger          Logger

	startedAt int64
	stoppedAt int64

	callBack func(*TopupAction) error
}

type TopupAction struct {
	Name            string
	BatchID         string
	PreviousBalance string
	CurrentBalance  string
	PreviousDepth   int
	CurrentDepth    int
	DoneAt          int64
	DepthAdded      int
	AmountTopped    string
}

type stamp struct {
	BatchID       string `json:"batchID"`
	Utilization   int    `json:"utilization"`
	Usable        bool   `json:"usable"`
	Label         string `json:"label"`
	Depth         int    `json:"depth"`
	Amount        string `json:"amount"`
	BucketDepth   int    `json:"bucketDepth"`
	BlockNumber   int    `json:"blockNumber"`
	ImmutableFlag bool   `json:"immutableFlag"`
	Exists        bool   `json:"exists"`
	BatchTTL      int    `json:"batchTTL"`
}

func newTopupTask(ctx context.Context, name, batchId, url, balanceEndpoint string, minAmount, topAmount *big.Int, interval time.Duration, cb func(*TopupAction) error, logger Logger) (*Topup, error) {
	if len(batchId) != 64 {
		return nil, fmt.Errorf("invalid batchID")
	}
	_, err := hex.DecodeString(batchId)
	if err != nil {
		return nil, fmt.Errorf("invalid batchID")
	}
	ctx2, cancel := context.WithCancel(ctx)
	return &Topup{
		batchId:         batchId,
		url:             url,
		balanceEndpoint: balanceEndpoint,
		name:            name,
		minAmount:       minAmount,
		topupAmount:     topAmount,
		interval:        interval,
		active:          true,
		ctx:             ctx2,
		cancel:          cancel,
		logger:          logger,

		callBack: cb,
	}, nil
}

func (t *Topup) Execute(context.Context) error {
	var resp *http.Response
	defer func() {
		t.stoppedAt = time.Now().Unix()
		if resp != nil {
			resp.Body.Close()
		}
	}()
	t.startedAt = time.Now().Unix()
	for {
		// get balance
		s, err := t.getStamp()
		if err != nil {
			t.logger.Error(err)
			return err
		}

		// check balance
		amount := &big.Int{}
		amount.SetString(s.Amount, 10)
		if amount.Cmp(t.minAmount) == -1 {
			// topup
			client := &http.Client{}
			url := fmt.Sprintf("%s/stamps/topup/%s/%d", t.url, t.batchId, t.topupAmount)

			req, err := http.NewRequest(http.MethodPatch, url, nil)
			req.Header.Set("Content-Type", "application/json")
			if err != nil {
				t.logger.Error(err)
				return err
			}

			resp, err = client.Do(req)
			if err != nil {
				t.logger.Error(err)
				return err
			}
			if resp.StatusCode != 202 {
				t.logger.Errorf("failed to top up %s. got code %d\n", t.batchId, resp.StatusCode)
				continue
			}

			err = resp.Body.Close()
			if err != nil {
				t.logger.Error(err)
				return err
			}
			sNew, err := t.getStamp()
			if err != nil {
				t.logger.Error(err)
				return err
			}
			action := &TopupAction{
				BatchID:         t.batchId,
				Name:            "topup",
				PreviousBalance: s.Amount,
				PreviousDepth:   s.Depth,
				CurrentBalance:  sNew.Amount,
				CurrentDepth:    sNew.Depth,
				DoneAt:          time.Now().Unix(),
				AmountTopped:    t.topupAmount.String(),
			}
			// callback with action
			err = t.callBack(action)
			if err != nil {
				t.logger.Error("failed to run callback for batchId %s Action %+v: %s", t.batchId, action, err.Error())
			}
		}

		// check depth
		d := math.Exp2(float64(s.Depth - s.BucketDepth))
		var used = float64(s.Utilization) / d
		if used > 0.8 {
			s, err = t.getStamp()
			if err != nil {
				t.logger.Error(err)
				return err
			}

			client := &http.Client{}
			url := fmt.Sprintf("%s/stamps/dilute/%s/%d", t.url, t.batchId, s.Depth+2)

			req, err := http.NewRequest(http.MethodPatch, url, nil)
			req.Header.Set("Content-Type", "application/json")
			if err != nil {
				t.logger.Error(err)
				return err
			}

			resp, err = client.Do(req)
			if err != nil {
				t.logger.Error(err)
				return err
			}
			if resp.StatusCode != 202 {
				t.logger.Errorf("failed to dilute %s. got code %d\n", t.batchId, resp.StatusCode)
				continue
			}
			err = resp.Body.Close()
			if err != nil {
				t.logger.Error(err)
				return err
			}
			sNew, err := t.getStamp()
			if err != nil {
				t.logger.Error(err)
				return err
			}
			action := &TopupAction{
				BatchID:         t.batchId,
				Name:            "dilute",
				PreviousBalance: s.Amount,
				PreviousDepth:   s.Depth,
				CurrentBalance:  sNew.Amount,
				CurrentDepth:    sNew.Depth,
				DoneAt:          time.Now().Unix(),
				DepthAdded:      2,
			}

			// callback with action
			err = t.callBack(action)
			if err != nil {
				t.logger.Error("failed to run callback for batchId %s Action %+v: %s", t.batchId, action, err.Error())
			}
		}
		select {
		case <-time.After(t.interval):
		case <-t.ctx.Done():
			return nil
		}
	}
}

func (t *Topup) Name() string {
	return t.name
}

func (t *Topup) Stop() {
	t.cancel()
}

func (t *Topup) getStamp() (*stamp, error) {
	resp, err := http.Get(fmt.Sprintf("%s/stamps/%s", t.url, t.batchId))
	if err != nil {
		t.logger.Error(err)
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.logger.Error(err)
		return nil, err
	}
	err = resp.Body.Close()
	if err != nil {
		t.logger.Error(err)
		return nil, err
	}

	if resp.StatusCode != 200 {
		t.logger.Error(string(body))
		return nil, fmt.Errorf("%s %s", body, t.batchId)
	}
	s := &stamp{}
	err = json.Unmarshal(body, s)
	if err != nil {
		t.logger.Error(err)
		return nil, err
	}
	return s, nil
}
