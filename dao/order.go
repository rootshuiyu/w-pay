package dao

import (
	"time"

	"wpay/model"

	"gorm.io/gorm"
)

type OrderQuery struct {
	StoreIDs    []uint64
	PlatformIDs []uint64
	PayType   *model.PayType
	Status    *model.OrderStatus
	StartTime *time.Time
	EndTime   *time.Time
	OrderNo   string
	Page      int
	PageSize  int
}

type OrderDAO struct{}

func NewOrderDAO() *OrderDAO { return &OrderDAO{} }

func (d *OrderDAO) Create(order *model.Order) error {
	return DB.Create(order).Error
}

func (d *OrderDAO) FindByOrderNo(orderNo string) (*model.Order, error) {
	var order model.Order
	err := DB.Where("order_no = ?", orderNo).First(&order).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &order, err
}

// UpdateStatusPaid 事务原子更新订单为已支付，并写入脱敏回调摘要
func (d *OrderDAO) UpdateStatusPaid(orderNo string, payAmount int64, transactionID string, payTime time.Time, notifyData string) error {
	return DB.Model(&model.Order{}).
		Where("order_no = ? AND order_status = ?", orderNo, model.OrderStatusPending).
		Updates(map[string]interface{}{
			"order_status":   model.OrderStatusPaid,
			"pay_amount":     payAmount,
			"transaction_id": transactionID,
			"pay_time":       payTime,
			"notify_data":    notifyData,
		}).Error
}

func (d *OrderDAO) CloseTimeoutOrders(before time.Time) (int64, error) {
	result := DB.Model(&model.Order{}).
		Where("order_status = ? AND created_at < ?", model.OrderStatusPending, before).
		Update("order_status", model.OrderStatusClosed)
	return result.RowsAffected, result.Error
}

func (d *OrderDAO) Query(q OrderQuery) ([]model.Order, int64, error) {
	var list []model.Order
	var total int64
	query := DB.Model(&model.Order{})
	if len(q.StoreIDs) > 0 {
		query = query.Where("store_id IN ?", q.StoreIDs)
	}
	if len(q.PlatformIDs) > 0 {
		query = query.Where("platform_id IN ?", q.PlatformIDs)
	}
	if q.PayType != nil {
		query = query.Where("pay_type = ?", *q.PayType)
	}
	if q.Status != nil {
		query = query.Where("order_status = ?", *q.Status)
	}
	if q.StartTime != nil {
		query = query.Where("created_at >= ?", *q.StartTime)
	}
	if q.EndTime != nil {
		query = query.Where("created_at <= ?", *q.EndTime)
	}
	if q.OrderNo != "" {
		query = query.Where("order_no = ?", q.OrderNo)
	}
	query.Count(&total)
	err := query.Offset((q.Page - 1) * q.PageSize).Limit(q.PageSize).Order("created_at DESC").Find(&list).Error
	return list, total, err
}

type OrderStat struct {
	StoreID     uint64 `json:"store_id,string,omitempty"`
	PlatformID  uint64 `json:"platform_id,string,omitempty"`
	TotalCount  int64  `json:"total_count"`
	TotalAmount int64  `json:"total_amount"`
	PaidCount   int64  `json:"paid_count"`
	PaidAmount  int64  `json:"paid_amount"`
	StatDate    string `json:"stat_date"`
}

func (d *OrderDAO) StatByStores(storeIDs []uint64, start, end time.Time, groupByDay bool) ([]OrderStat, error) {
	var stats []OrderStat
	dateFmt := "YYYY-MM-DD"
	if !groupByDay {
		dateFmt = "YYYY-MM"
	}
	query := DB.Model(&model.Order{}).
		Select("store_id, to_char(created_at, ?) as stat_date, COUNT(*) as total_count, SUM(total_amount) as total_amount, "+
			"SUM(CASE WHEN order_status = ? THEN 1 ELSE 0 END) as paid_count, "+
			"SUM(CASE WHEN order_status = ? THEN pay_amount ELSE 0 END) as paid_amount",
			dateFmt, model.OrderStatusPaid, model.OrderStatusPaid).
		Where("created_at >= ? AND created_at <= ?", start, end)
	if len(storeIDs) > 0 {
		query = query.Where("store_id IN ?", storeIDs)
	}
	err := query.Group("store_id, stat_date").Order("stat_date DESC").Scan(&stats).Error
	return stats, err
}

func (d *OrderDAO) StatByPlatforms(platformIDs []uint64, start, end time.Time, groupByDay bool) ([]OrderStat, error) {
	var stats []OrderStat
	dateFmt := "YYYY-MM-DD"
	if !groupByDay {
		dateFmt = "YYYY-MM"
	}
	query := DB.Model(&model.Order{}).
		Select("platform_id, to_char(created_at, ?) as stat_date, COUNT(*) as total_count, SUM(total_amount) as total_amount, "+
			"SUM(CASE WHEN order_status = ? THEN 1 ELSE 0 END) as paid_count, "+
			"SUM(CASE WHEN order_status = ? THEN pay_amount ELSE 0 END) as paid_amount",
			dateFmt, model.OrderStatusPaid, model.OrderStatusPaid).
		Where("created_at >= ? AND created_at <= ? AND platform_id > 0", start, end)
	if len(platformIDs) > 0 {
		query = query.Where("platform_id IN ?", platformIDs)
	}
	err := query.Group("platform_id, stat_date").Order("stat_date DESC").Scan(&stats).Error
	return stats, err
}

func (d *OrderDAO) CreateSensitiveLog(log *model.SensitiveLog) error {
	return DB.Create(log).Error
}
