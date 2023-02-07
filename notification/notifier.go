package notification

type Notifier interface {
	Login()
	Event()
	Risk()
	SendDong(uid int64, title string, content string) error
}

// TemplateRender 模板渲染接口
type TemplateRender interface {
	RendEventDong() (title string, content string, err error)
	RendEventEmail() (title string, content string, err error)
	RendEventSMS()
	RendEventPhone()

	RendRiskDong() (title string, content string, err error)
	RendRiskEmail() (title string, content string, err error)
	RendRiskSMS() (content string, err error)
	RendRiskPhone() (content string, err error)
}

// 产生 risk/event -> 保存 -> 通知 -> 过滤器 -> 渲染 -> 发送
