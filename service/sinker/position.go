package sinker

import (
	"context"
	"database/sql"
	"sync"

	"github.com/frozenpine/rhmonitor4go/service"
)

type PositionSinker struct {
	posSinker func(*SinkAccountPosition) (sql.Result, error)

	ctx     context.Context
	source  service.RohonMonitor_SubInvestorPositionClient
	output  chan *SinkAccountPosition
	posPool sync.Pool

	settlements *sync.Map
	posCache    map[string]*SinkAccountPosition
}

func NewPositionSinker(
	ctx context.Context, settlements *sync.Map,
	src service.RohonMonitor_SubInvestorPositionClient,
) (*PositionSinker, error) {
	if src == nil || settlements == nil {
		return nil, ErrInvalidArgs
	}

	posSinker, err := InsertDB[SinkAccountPosition](
		ctx, "operation_account_position",
	)
	if err != nil {
		return nil, err
	}

	sinker := &PositionSinker{
		ctx:         ctx,
		posSinker:   posSinker,
		source:      src,
		settlements: settlements,
		output:      make(chan *SinkAccountPosition, 1),
		posPool:     sync.Pool{New: func() any { return new(SinkAccountPosition) }},
		posCache:    make(map[string]*SinkAccountPosition),
	}

	// go sinker.run()

	return sinker, nil
}
