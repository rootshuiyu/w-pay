package service

import (
	"fmt"
	"time"

	"wpay/dao"
	"wpay/model"

	"github.com/xuri/excelize/v2"
)

type ReconcileService struct {
	orderDAO    *dao.OrderDAO
	storeDAO    *dao.StoreDAO
	platformDAO *dao.PayPlatformDAO
}

func NewReconcileService() *ReconcileService {
	return &ReconcileService{
		orderDAO:    dao.NewOrderDAO(),
		storeDAO:    dao.NewStoreDAO(),
		platformDAO: dao.NewPayPlatformDAO(),
	}
}

type ReconcileQuery struct {
	StoreIDs     []uint64
	PlatformIDs  []uint64
	PayType      *model.PayType
	Status       *model.OrderStatus
	StartTime    time.Time
	EndTime      time.Time
	GroupByDay   bool
	Dimension    string // store | platform
}

type ReconcileStatResponse struct {
	Stats     []dao.OrderStat      `json:"stats"`
	Stores    []model.Store        `json:"stores,omitempty"`
	Platforms []model.PayPlatform  `json:"platforms,omitempty"`
}

func (s *ReconcileService) Stat(q ReconcileQuery) (*ReconcileStatResponse, error) {
	var stats []dao.OrderStat
	var err error
	if q.Dimension == "platform" {
		stats, err = s.orderDAO.StatByPlatforms(q.PlatformIDs, q.StartTime, q.EndTime, q.GroupByDay)
	} else {
		stats, err = s.orderDAO.StatByStores(q.StoreIDs, q.StartTime, q.EndTime, q.GroupByDay)
	}
	if err != nil {
		return nil, err
	}
	resp := &ReconcileStatResponse{Stats: stats}
	if q.Dimension == "platform" {
		ids := q.PlatformIDs
		if len(ids) == 0 {
			for _, st := range stats {
				if st.PlatformID > 0 {
					ids = append(ids, st.PlatformID)
				}
			}
		}
		if len(ids) > 0 {
			platforms, err := s.platformDAO.ListByIDs(ids)
			if err != nil {
				return nil, err
			}
			resp.Platforms = platforms
		}
	} else if len(q.StoreIDs) > 0 {
		stores, err := s.storeDAO.ListByIDs(q.StoreIDs)
		if err != nil {
			return nil, err
		}
		resp.Stores = stores
	}
	return resp, nil
}

func (s *ReconcileService) ExportExcel(q dao.OrderQuery) (*excelize.File, error) {
	q.Page = 1
	q.PageSize = 10000
	orders, _, err := s.orderDAO.Query(q)
	if err != nil {
		return nil, err
	}
	storeMap := make(map[uint64]string)
	if len(q.StoreIDs) > 0 {
		stores, _ := s.storeDAO.ListByIDs(q.StoreIDs)
		for _, st := range stores {
			storeMap[st.ID] = st.StoreName
		}
	}
	platformMap := make(map[uint64]string)
	if len(q.PlatformIDs) > 0 {
		platforms, _ := s.platformDAO.ListByIDs(q.PlatformIDs)
		for _, p := range platforms {
			platformMap[p.ID] = p.PlatformName
		}
	} else {
		var ids []uint64
		seen := make(map[uint64]bool)
		for _, o := range orders {
			if o.PlatformID > 0 && !seen[o.PlatformID] {
				ids = append(ids, o.PlatformID)
				seen[o.PlatformID] = true
			}
		}
		if len(ids) > 0 {
			platforms, _ := s.platformDAO.ListByIDs(ids)
			for _, p := range platforms {
				platformMap[p.ID] = p.PlatformName
			}
		}
	}
	f := excelize.NewFile()
	sheet := "对账明细"
	f.SetSheetName("Sheet1", sheet)
	headers := []string{"订单号", "代收平台", "门店", "支付渠道", "订单金额(元)", "实付金额(元)", "状态", "设备流水号", "备注", "创建时间", "支付时间"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}
	statusText := map[model.OrderStatus]string{
		model.OrderStatusPending: "待支付",
		model.OrderStatusPaid:    "已支付",
		model.OrderStatusClosed:  "已关闭",
		model.OrderStatusRefund:  "退款",
	}
	payTypeText := map[model.PayType]string{model.PayTypeWechat: "微信", model.PayTypeAlipay: "支付宝"}
	for row, o := range orders {
		storeName := storeMap[o.StoreID]
		if storeName == "" {
			storeName = fmt.Sprintf("门店#%d", o.StoreID)
		}
		platformName := platformMap[o.PlatformID]
		if platformName == "" && o.PlatformID > 0 {
			platformName = fmt.Sprintf("平台#%d", o.PlatformID)
		}
		payTime := ""
		if o.PayTime != nil {
			payTime = o.PayTime.Format("2006-01-02 15:04:05")
		}
		vals := []interface{}{
			o.OrderNo, platformName, storeName, payTypeText[o.PayType],
			fmt.Sprintf("%.2f", float64(o.TotalAmount)/100),
			fmt.Sprintf("%.2f", float64(o.PayAmount)/100),
			statusText[o.OrderStatus], o.DeviceSN, o.Subject,
			o.CreatedAt.Format("2006-01-02 15:04:05"), payTime,
		}
		for col, v := range vals {
			cell, _ := excelize.CoordinatesToCellName(col+1, row+2)
			f.SetCellValue(sheet, cell, v)
		}
	}
	return f, nil
}

func (s *ReconcileService) ExportStatExcel(q ReconcileQuery) (*excelize.File, error) {
	statResp, err := s.Stat(q)
	if err != nil {
		return nil, err
	}
	storeMap := make(map[uint64]string)
	platformMap := make(map[uint64]string)
	for _, st := range statResp.Stores {
		storeMap[st.ID] = st.StoreName
	}
	for _, p := range statResp.Platforms {
		platformMap[p.ID] = p.PlatformName
	}
	f := excelize.NewFile()
	sheet := "汇总统计"
	f.SetSheetName("Sheet1", sheet)
	byPlatform := q.Dimension == "platform"
	headers := []string{"日期", "门店", "订单总数", "订单总金额(元)", "已支付笔数", "已支付金额(元)"}
	if byPlatform {
		headers = []string{"日期", "代收平台", "订单总数", "订单总金额(元)", "已支付笔数", "已支付金额(元)"}
	}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}
	for row, st := range statResp.Stats {
		nameCol := storeMap[st.StoreID]
		if nameCol == "" {
			nameCol = fmt.Sprintf("门店#%d", st.StoreID)
		}
		if byPlatform {
			nameCol = platformMap[st.PlatformID]
			if nameCol == "" {
				nameCol = fmt.Sprintf("平台#%d", st.PlatformID)
			}
		}
		vals := []interface{}{
			st.StatDate, nameCol, st.TotalCount,
			fmt.Sprintf("%.2f", float64(st.TotalAmount)/100),
			st.PaidCount, fmt.Sprintf("%.2f", float64(st.PaidAmount)/100),
		}
		for col, v := range vals {
			cell, _ := excelize.CoordinatesToCellName(col+1, row+2)
			f.SetCellValue(sheet, cell, v)
		}
	}
	return f, nil
}
