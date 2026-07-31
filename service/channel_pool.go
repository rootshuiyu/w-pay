package service

import (
	"context"

	"wpay/common"
	"wpay/dao"
	"wpay/model"
)

// ChannelPoolService 多商户码轮询代收
type ChannelPoolService struct {
	channelDAO *dao.PayChannelDAO
	storeDAO   *dao.StoreDAO
}

func NewChannelPoolService() *ChannelPoolService {
	return &ChannelPoolService{
		channelDAO: dao.NewPayChannelDAO(),
		storeDAO:   dao.NewStoreDAO(),
	}
}

func (s *ChannelPoolService) refreshQuota(ch *model.PayChannel) *model.PayChannel {
	if ch == nil {
		return nil
	}
	if err := s.channelDAO.ResetDailyUsedIfNeeded(ch); err != nil {
		return ch
	}
	refreshed, err := s.channelDAO.FindByID(ch.ID)
	if err != nil || refreshed == nil {
		return ch
	}
	return refreshed
}

// Select 从平台码池选取商户码（platformID=0 为未分配池；storeID 可选进一步限定门店）
func (s *ChannelPoolService) Select(platformID, storeID uint64, payType model.PayType, amount int64) (*model.Store, *model.PayChannel, error) {
	candidates, err := s.channelDAO.ListPoolCandidates(payType, platformID, storeID)
	if err != nil {
		return nil, nil, err
	}

	eligible := filterEligible(s, candidates, amount)
	if len(eligible) == 0 && platformID > 0 {
		singles, err := s.channelDAO.ListPlatformSingleChannels(payType, platformID, storeID)
		if err != nil {
			return nil, nil, err
		}
		eligible = filterEligible(s, singles, amount)
	}

	if len(candidates) == 0 && len(eligible) == 0 {
		if platformID > 0 {
			return nil, nil, common.ErrInvalidInput("该平台暂无可用支付通道")
		}
		return nil, nil, common.ErrInvalidInput("轮询池无可用商户码，请先在渠道管理中启用「参与轮询」")
	}
	if len(eligible) == 0 {
		return nil, nil, common.ErrInvalidInput("所有商户码已达额度上限或单笔超限，请调整额度或稍后再试")
	}

	idx, err := nextRotateIndex(int8(payType), platformID, storeID, totalWeight(eligible))
	if err != nil {
		common.Log.Warn("channel pool rotate fallback: %v", err)
		idx = 0
	}
	ch := selectWeightedChannel(eligible, idx)

	store, err := s.storeDAO.FindByID(ch.StoreID)
	if err != nil || store == nil {
		return nil, nil, common.ErrInvalidInput("商户码关联门店不存在")
	}
	if store.Status != model.StoreStatusNormal {
		return nil, nil, common.ErrInvalidInput("商户码关联门店已停用")
	}
	return store, &ch, nil
}

func filterEligible(s *ChannelPoolService, list []model.PayChannel, amount int64) []model.PayChannel {
	eligible := make([]model.PayChannel, 0, len(list))
	for i := range list {
		refreshed := s.refreshQuota(&list[i])
		if refreshed == nil || !refreshed.CanAccept(amount) {
			continue
		}
		eligible = append(eligible, *refreshed)
	}
	return eligible
}

func nextRotateIndex(payType int8, platformID, storeID uint64, totalWeight int) (int, error) {
	if totalWeight <= 0 {
		return 0, common.ErrInvalidInput("empty pool")
	}
	ctx := context.Background()
	key := common.ChannelRotateKey(payType, platformID, storeID)
	val, err := dao.RDB.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	return int((val - 1) % int64(totalWeight)), nil
}

func totalWeight(channels []model.PayChannel) int {
	weight := 0
	for _, ch := range channels {
		if ch.RotateWeight > 0 {
			weight += ch.RotateWeight
		} else {
			weight++
		}
	}
	if weight == 0 {
		weight = len(channels)
	}
	return weight
}

func selectWeightedChannel(channels []model.PayChannel, idx int) model.PayChannel {
	if len(channels) == 0 {
		return model.PayChannel{}
	}
	idx = idx % totalWeight(channels)
	for _, ch := range channels {
		w := ch.RotateWeight
		if w <= 0 {
			w = 1
		}
		if idx < w {
			return ch
		}
		idx -= w
	}
	return channels[len(channels)-1]
}

func (s *ChannelPoolService) PoolStats(payType *model.PayType) ([]model.PayChannel, error) {
	return s.poolStatsForPlatform(payType, nil)
}

func (s *ChannelPoolService) PoolStatsForPlatform(payType *model.PayType, platformID *uint64) ([]model.PayChannel, error) {
	return s.poolStatsForPlatform(payType, platformID)
}

func (s *ChannelPoolService) poolStatsForPlatform(payType *model.PayType, platformID *uint64) ([]model.PayChannel, error) {
	list, err := s.channelDAO.ListPoolStats(payType, platformID)
	if err != nil {
		return nil, err
	}
	for i := range list {
		refreshed := s.refreshQuota(&list[i])
		if refreshed != nil {
			list[i] = *refreshed
		}
	}
	return list, err
}
