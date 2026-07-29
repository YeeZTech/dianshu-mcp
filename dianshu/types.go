// Package dianshu - see README for details.
//
// Author: zhyyao

package dianshu

// API 响应相关的数据结构

// UserInfo 用户信息（映射 /login/getUserInfo 响应）
type UserInfo struct {
	ID                   int64   `json:"id"`                   // 用户 ID
	Nickname             string  `json:"userName"`             // 昵称
	CompanyID            int64   `json:"companyId"`            // 企业 ID
	CompanyCode          string  `json:"companyCode"`          // 企业编码
	CompanyName          string  `json:"companyName"`          // 企业名称
	Logo                 string  `json:"logo"`                 // 企业 Logo
	Phone                string  `json:"mobile"`               // 手机号
	UserEmail            string  `json:"userEmail"`            // 邮箱
	Role                 int     `json:"role"`                 // 角色
	IsCertificated       int     `json:"isCertificated"`       // 是否实名/认证
	TaskChargeRatio      float64 `json:"taskChargeRatio"`      // 任务分成比例
	CompanyLockTime      string  `json:"companyLockTime"`      // 企业锁定时间
	DataLockTime         string  `json:"dataLockTime"`         // 数据锁定时间
	LastLoginTime        string  `json:"lastLoginTime"`        // 最后登录时间
	PayType              string  `json:"payType"`              // 支付类型
	Avatar               string  `json:"userImage"`            // 头像
	ChainAddress         string  `json:"chainAddress"`         // 链上地址
	Credential           string  `json:"credential"`           // 凭证
	PrivateKey           string  `json:"privateKey"`           // 私钥 JSON
	DatasetChargeRatio   float64 `json:"datasetChargeRatio"`   // 数据集分成比例
	AlgorithmChargeRatio float64 `json:"algorithmChargeRatio"` // 算法分成比例
	ComputingChargeRatio float64 `json:"computingChargeRatio"` // 算力分成比例
	OpenID               string  `json:"openId"`               // OpenID
	IsRegister           int     `json:"isRegister"`           // 注册状态
	Activity             string  `json:"activity"`             // 活动信息
	AppCode              string  `json:"appCode"`              // AppCode
	BindStatus           int     `json:"bindStatus"`           // 绑定状态
	ChatStatus           int     `json:"chatStatus"`           // 聊天状态
	UserID               string  `json:"userNo"`               // 用户编号
	IsNotify             int     `json:"isNotify"`             // 通知状态
	FirmVerify           int     `json:"firmVerify"`           // 企业认证状态
	FirmName             string  `json:"firmName"`             // 企业认证名称
	FirmTaxID            string  `json:"firmTaxId"`            // 企业税号
	FirmWebsite          string  `json:"firmWebsite"`          // 企业官网
	FirmDescription      string  `json:"firmDescription"`      // 企业介绍
	Description          string  `json:"description"`          // 卖家介绍
	UniqueUserID         string  `json:"uniqueUserId"`         // 唯一用户 ID
	DSUserNo             string  `json:"dsUserNo"`             // 典枢号
	AuthPasswordHosted   int     `json:"authPasswordHosted"`   // 托管密码状态
	ShowEnterpriseGuide  bool    `json:"showEnterpriseGuide"`  // 是否展示企业引导
}

// PageRequest 通用分页请求
type PageRequest struct {
	PageNo   int `json:"pageNo"`
	PageSize int `json:"pageSize"`
}

// PageInfo 通用分页信息
type PageInfo struct {
	PageNo    int    `json:"pageNo"`
	PageSize  int    `json:"pageSize"`
	OrderBy   string `json:"orderBy"`
	Order     string `json:"order"`
	Count     int    `json:"count"`
	TotalPage int    `json:"totalPage"`
}

// WalletBalance 钱包余额
type WalletBalance struct {
	Available    float64  `json:"available"`
	Frozen       float64  `json:"frozen"`
	Profit       *float64 `json:"profit"`
	WithDrawable float64  `json:"withDrawable"`
}

// WalletTransaction 钱包交易明细
type WalletTransaction struct {
	Code                     string   `json:"code"`
	ChangeType               string   `json:"changeType"`
	OrderStatus              string   `json:"orderStatus"`
	Type                     string   `json:"type"`
	Status                   string   `json:"status"`
	Amount                   float64  `json:"amount"`
	Amount1                  float64  `json:"amount1"`
	Amount2                  float64  `json:"amount2"`
	Time                     int64    `json:"time"`
	LoginNo                  string   `json:"loginNo"`
	CreateCompanyID          string   `json:"createCompanyId"`
	PriceInfo                any      `json:"priceInfo"`
	EndTime                  int64    `json:"endTime"`
	DatasetID                string   `json:"datasetId"`
	DatasetName              string   `json:"datasetName"`
	ServiceCharge            float64  `json:"serviceCharge"`
	ServiceChargeRatio       float64  `json:"serviceChargeRatio"`
	CreateUser               string   `json:"createUser"`
	ChargeRatio              float64  `json:"chargeRatio"`
	TransactionHash          *string  `json:"transactionHash"`
	TotalCommissionCharge    *float64 `json:"totalCommissionCharge"`
	SellerCommissionCharge   float64  `json:"sellerCommissionCharge"`
	SellerCommissionRatio    *float64 `json:"sellerCommissionRatio"`
	PlatformCommissionCharge float64  `json:"platformCommissionCharge"`
	PlatformCommissionRatio  *float64 `json:"platformCommissionRatio"`
	Items                    any      `json:"items"`
	TradeID                  *string  `json:"tradeId"`
	Pattern                  *string  `json:"pattern"`
}

// WalletTransactionListResponse 钱包交易明细响应
type WalletTransactionListResponse struct {
	ResultCode int                 `json:"resultCode"`
	ResultDesc string              `json:"resultDesc"`
	Data       []WalletTransaction `json:"data"`
	Page       PageInfo            `json:"page"`
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

// TaskItem 任务项（来自 /system/task/taskList 和 /system/trade/tradeList）
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
	FileURL            string                   `json:"fileUrl"`
	Pattern            string                   `json:"pattern"`
	DownloadList       map[string][]DownloadURL `json:"downloadList"`
	ClientDownloadType string                   `json:"clientDownloadType"`
	ClientDownloadUrl  string                   `json:"clientDownloadUrl"`
	ChecksumUrl        string                   `json:"checksumUrl"`
	APIType            int                      `json:"apiType"`
	PrivateKey         string                   `json:"privateKey"`
	PublishStatus      int                      `json:"publishStatus"`
	EncryptFileHash    string                   `json:"encryptFileHash"`
	ClientDownloadURL  string                   `json:"clientDownloadUrl"`
}

// TaskPrivateKeyResult 任务密封文件私钥
type TaskPrivateKeyResult struct {
	ID            int    `json:"id"`
	PrivateKey    string `json:"privateKey"`
	PublishStatus int    `json:"publishStatus"`
}

// API 详情

// APIDetail API 详情（来自 /api/detail）
type APIDetail struct {
	APIID         int          `json:"apiId"`
	UniqueAPIID   string       `json:"uniqueApiId"` // SDK 使用的 apiCode
	APIName       string       `json:"apiName"`
	APICode       string       `json:"apiCode"`
	APIURL        string       `json:"apiUrl"`
	MappingURL    string       `json:"mappingUrl"`
	RequestMethod int          `json:"requestMethod"` // 0=POST, 1=GET
	ReqHeaders    []ParamField `json:"requestHeaders"`
	QueryParams   []ParamField `json:"queryParams"`
	BodyParams    []ParamField `json:"bodyParams"`
	APIType       string       `json:"apiType"` // sync/async
	Description   string       `json:"descriptionTxt"`
}

// ParamField API 参数/请求头字段
type ParamField struct {
	ParamName    string `json:"paramName"`
	TypeName     string `json:"typeName"`
	Required     int    `json:"required"`
	ExampleValue string `json:"exampleValue"`
	Description  string `json:"description"`
}

// APIInitResponse /api/privateKey 返回
type APIInitResponse struct {
	EnclaveHash string `json:"enclaveHash"`
	DianPkey    string `json:"dianPkey"`
	APIHash     string `json:"apiHash"`
	PrivateKey  string `json:"privateKey"`
	PublicKey   string `json:"publicKey"`
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

// OrderItem 订单项

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

// PurchasedAPIItem 已购买 API 列表项（来自 data-api getBuyerApi）
type PurchasedAPIItem struct {
	APIID          int    `json:"apiId"`
	APICode        string `json:"apiCode"`
	APIName        string `json:"apiName"`
	CreateUser     string `json:"createUser"`
	CreateTime     string `json:"createTime"`
	Frequency      int    `json:"frequency"`
	RemainingTimes int    `json:"remainingTimes"`
	Usage          string `json:"usage"`
}

// PurchasedAPIListResponse 已购买 API 列表响应
type PurchasedAPIListResponse struct {
	ResultCode int                `json:"resultCode"`
	ResultDesc string             `json:"resultDesc"`
	Data       []PurchasedAPIItem `json:"data"`
	Page       PageInfo           `json:"page"`
}
