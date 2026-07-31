// 模拟支付成功：更新订单 + 累计商户码日额度（与回调 finishPaid 一致）
//go:build ignore

package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"wpay/config"
	"wpay/dao"
)

func main() {
	orderNo := flag.String("order", "", "order_no")
	amount := flag.Int64("amount", 0, "pay amount fen")
	flag.Parse()
	if *orderNo == "" {
		fmt.Fprintln(os.Stderr, "usage: go run sim_mark_paid.go -order=xxx -amount=10000")
		os.Exit(1)
	}

	os.Setenv("APP_ENV", "dev")
	cfg, err := config.Load("dev")
	if err != nil {
		panic(err)
	}
	if err := dao.InitDB(&cfg.Database); err != nil {
		panic(err)
	}

	orderDAO := dao.NewOrderDAO()
	order, err := orderDAO.FindByOrderNo(*orderNo)
	if err != nil || order == nil {
		panic("order not found")
	}
	payAmt := *amount
	if payAmt <= 0 {
		payAmt = order.TotalAmount
	}
	if err := orderDAO.UpdateStatusPaid(*orderNo, payAmt, "SIM_TXN_"+*orderNo, time.Now(), "sim_notify"); err != nil {
		panic(err)
	}
	if order.ChannelID > 0 && payAmt > 0 {
		_ = dao.NewPayChannelDAO().AddDailyUsed(order.ChannelID, payAmt)
	}
	fmt.Printf("marked paid: %s amount=%d channel=%d\n", *orderNo, payAmt, order.ChannelID)
}
