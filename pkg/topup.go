package pkg

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"time"

	logging "github.com/ipfs/go-log/v2"
)

var (
	log = logging.Logger("stampkeeper/pkg")
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

func newTopupTask(ctx context.Context, name, batchId, url, balanceEndpoint string, minAmount, topAmount *big.Int, interval time.Duration) (*Topup, error) {
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
	}, nil
}

func (t *Topup) Execute(context.Context) error {
	var resp *http.Response
	var err error
	defer func() {
		if resp != nil {
			resp.Body.Close()
		}
	}()
	for {
		// get balance
		resp, err = http.Get(fmt.Sprintf("%s/stamps/%s", t.url, t.batchId))
		if err != nil {
			log.Error(err)
			return err
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Error(err)
			return err
		}
		err = resp.Body.Close()
		if err != nil {
			log.Error(err)
			return err
		}
		s := &stamp{}
		err = json.Unmarshal(body, s)
		if err != nil {
			log.Error(err)
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
				log.Error(err)
				return err
			}

			resp, err = client.Do(req)
			if err != nil {
				log.Error(err)
				return err
			}
			if resp.StatusCode != 202 {
				log.Errorf("failed to top up %s. got code %d\n", t.batchId, resp.StatusCode)
				return fmt.Errorf("failed to top up %s. got code %d\n", t.batchId, resp.StatusCode)
			}
			err = resp.Body.Close()
			if err != nil {
				log.Error(err)
				return err
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
