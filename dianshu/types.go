package dianshu

// API 响应相关的数据结构

// UserInfo 用户基本信息（映射 /login/getUserInfo 响应）
type UserInfo struct {
	UserID   string `json:"userNo"`    // 实际字段名 userNo
	Nickname string `json:"userName"`  // 实际字段名 userName
	Avatar   string `json:"userImage"` // 实际字段名 userImage
	Phone    string `json:"mobile"`    // 实际字段名 mobile
}

// OrderQueryData 订单查询数据
type OrderQueryData struct {
	Total     int         `json:"total"`
	PageNo    int         `json:"pageNo"`
	PageSize  int         `json:"pageSize"`
	OrderList []OrderItem `json:"list"`
}

// TaskListResponse 任务列表响应
type TaskListResponse struct {
	ResultCode int         `json:"resultCode"`
	ResultDesc string      `json:"resultDesc"`
	Data       []TaskItem  `json:"data"`
}

// TaskItem 任务项（来自 /system/task/taskList）
type TaskItem struct {
	ID             int          `json:"id"`
	DatasetID      int          `json:"datasetId"`
	DatasetName    string       `json:"datasetName"`
	Price          float64      `json:"price"`
	TaskCode       string       `json:"taskCode"`
	CreateTime     int64        `json:"createTime"`
	CreateTimeSql  string       `json:"createTimeSql"`
	Status         int          `json:"status"`
	PayStatus      int          `json:"payStatus"`
	TaskStatus     int          `json:"taskStatus"`
	DatasetUser    int          `json:"datasetUser"`
	DatasetUserName string      `json:"datasetUserName"`
	FileURL        string       `json:"fileUrl"`
}

// OrderItem 订单项
type OrderItem struct {
	OrderCode         string `json:"orderCode"`
	OrderType         int    `json:"orderType"`
	OrderName         string `json:"orderName"`
	OrderAmount       string `json:"orderAmount"`
	OrderPrice        string `json:"orderPrice"`
	OrderStatus       string `json:"orderStatus"`
	OrderPayStatus    string `json:"orderPayStatus"`
	OrderPayWay       string `json:"orderPayWay"`
	OrderCreateTime   string `json:"orderCreateTime"`
	OrderRefundButton string `json:"orderRefundButtonText"`
	OrderSeaButton    string `json:"orderSeaButtonText"`
}

// 以下为 JSON 序列化辅助类型，与 types.go 中的类型对应
// 这里重新定义以避免循环依赖
