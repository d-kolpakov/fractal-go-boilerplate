package pg

import (
	"context"
	"errors"
	"fmt"
	"github.com/d-kolpakov/logger"
	"strconv"
)

type Wrapper struct {
	common *ConnWrapper
	shards map[int]*ConnWrapper
}

func NewWrapper(common *ConnWrapper, shards map[int]*ConnWrapper) *Wrapper {
	return &Wrapper{
		common: common,
		shards: shards,
	}
}

func (w *Wrapper) Common() *ConnWrapper {
	return w.common
}

func (w *Wrapper) Shard(ctx context.Context) (*ConnWrapper, error) {
	num := w.getShardFromCtx(ctx)

	if num == -1 {
		return nil, errors.New("empty shard num in ctx")
	}

	shard, ok := w.shards[num]

	if !ok {
		return nil, fmt.Errorf("shard %d not exist in shard pool", num)
	}

	return shard, nil
}

func (w *Wrapper) getShardFromCtx(ctx context.Context) int {
	var key logger.ContextUIDKey = "shard"
	res := -1

	source := ctx.Value(key)

	if source == nil {
		return res
	}

	sourceString, ok := source.(string)
	if !ok {
		return res
	}

	converted, err := strconv.Atoi(sourceString)

	if err != nil || converted < 0 {
		return res
	}

	return converted
}
