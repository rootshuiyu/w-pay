package model

// PayScene 收银唤起方式
type PayScene string

const (
	PaySceneH5     PayScene = "h5"     // 手机浏览器：微信 H5 / 支付宝 WAP
	PaySceneNative PayScene = "native" // 扫码：微信 Native / 支付宝 precreate
)

func (p PayScene) Valid() bool {
	return p == PaySceneH5 || p == PaySceneNative
}

func ParsePayScene(s string) PayScene {
	switch PayScene(s) {
	case PaySceneNative:
		return PaySceneNative
	default:
		return PaySceneH5
	}
}
