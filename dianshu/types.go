package dianshu

// API 响应相关的数据结构

// UserInfo 用户基本信息（映射 /login/getUserInfo 响应）
type UserInfo struct {
	UserID   string `json:"userNo"`    // 用户编号
	Nickname string `json:"userName"`  // 用户名
	Avatar   string `json:"userImage"` // 头像
	Phone    string `json:"mobile"`    // 手机号
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
	ResultCode int        `json:"resultCode"`
	ResultDesc string     `json:"resultDesc"`
	Data       []TaskItem `json:"data"`
}

// DownloadURL 下载地址
type DownloadURL struct {
	URL string `json:"url"`
}

// TaskItem 任务项（来自 /system/task/taskList）
type TaskItem struct {
	ID                 int                      `json:"id"`
	DatasetID          int                      `json:"datasetId"`
	DatasetName        string                   `json:"datasetName"`
	Price              float64                  `json:"price"`
	TaskCode           string                   `json:"taskCode"`
	CreateTime         int64                    `json:"createTime"`
	CreateTimeSql      string                   `json:"createTimeSql"`
	Status             int                      `json:"status"`
	PayStatus          int                      `json:"payStatus"`
	TaskStatus         int                      `json:"taskStatus"`
	DatasetUser        int                      `json:"datasetUser"`
	DatasetUserName    string                   `json:"datasetUserName"`
	FileURL            string                   `json:"fileUrl"`            // 文件哈希标识（如 xxx.sealed）
	Pattern            string                   `json:"pattern"`            // 文件格式（如 zip）
	DownloadList       map[string][]DownloadURL `json:"downloadList"`       // 按类型分组的下载地址
	ClientDownloadType string                   `json:"clientDownloadType"` // 客户端下载类型
	ClientDownloadUrl  string                   `json:"clientDownloadUrl"`  // 客户端下载链接
	ChecksumUrl        string                   `json:"checksumUrl"`        // 校验值地址
	APIType            int                      `json:"apiType"`            // 0=数据产品, 1=API产品
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

// APIDetail API 产品详情（来自 data-api.dianshudata.com/api/detail）
type APIDetail struct {
	ApiID              int           `json:"apiId"`
	ApiName            string        `json:"apiName"`
	ApiType            string        `json:"apiType"`       // "sync" 同步API
	RequestMethod      int           `json:"requestMethod"` // 0=GET, 1=POST
	BodyParams         []APIParam    `json:"bodyParams"`
	JavaRequestExample string        `json:"javaRequestExample"`
	ExampleCodeList    []CodeExample `json:"exampleCodeList"`
	AppCode            string        `json:"appCode"`     // 用户标识
	ApiCode            string        `json:"apiCode"`     // API 产品标识
	ApiEndpoint        string        `json:"apiEndpoint"` // 调用地址
}

// APIParam API 请求参数
type APIParam struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Example     string `json:"example"`
	Description string `json:"description"`
}

// CodeExample 多语言代码示例
type CodeExample struct {
	Language string `json:"language"`
	Code     string `json:"code"`
}

// APIDetailResponse API 详情响应
type APIDetailResponse struct {
	ResultCode int        `json:"resultCode"`
	ResultDesc string     `json:"resultDesc"`
	Data       *APIDetail `json:"data"`
}

// 以下为 JSON 序列化辅助类型，与 types.go 中的类型对应
// 这里重新定义以避免循环依赖
