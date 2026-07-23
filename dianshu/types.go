package dianshu

// UserInfo 用户信息（映射 /login/getUserInfo 响应）
type UserInfo struct {
	ID                   int64   `json:"id"`
	Nickname             string  `json:"userName"`
	CompanyID            int64   `json:"companyId"`
	CompanyCode          string  `json:"companyCode"`
	CompanyName          string  `json:"companyName"`
	Logo                 string  `json:"logo"`
	Phone                string  `json:"mobile"`
	UserEmail            string  `json:"userEmail"`
	Role                 int     `json:"role"`
	IsCertificated       int     `json:"isCertificated"`
	TaskChargeRatio      float64 `json:"taskChargeRatio"`
	CompanyLockTime      string  `json:"companyLockTime"`
	DataLockTime         string  `json:"dataLockTime"`
	LastLoginTime        string  `json:"lastLoginTime"`
	PayType              string  `json:"payType"`
	Avatar               string  `json:"userImage"`
	ChainAddress         string  `json:"chainAddress"`
	Credential           string  `json:"credential"`
	PrivateKey           string  `json:"privateKey"`
	DatasetChargeRatio   float64 `json:"datasetChargeRatio"`
	AlgorithmChargeRatio float64 `json:"algorithmChargeRatio"`
	ComputingChargeRatio float64 `json:"computingChargeRatio"`
	OpenID               string  `json:"openId"`
	IsRegister           int     `json:"isRegister"`
	Activity             string  `json:"activity"`
	AppCode              string  `json:"appCode"`
	BindStatus           int     `json:"bindStatus"`
	ChatStatus           int     `json:"chatStatus"`
	UserID               string  `json:"userNo"`
	IsNotify             int     `json:"isNotify"`
	FirmVerify           int     `json:"firmVerify"`
	FirmName             string  `json:"firmName"`
	FirmTaxID            string  `json:"firmTaxId"`
	FirmWebsite          string  `json:"firmWebsite"`
	FirmDescription      string  `json:"firmDescription"`
	Description          string  `json:"description"`
	UniqueUserID         string  `json:"uniqueUserId"`
	DSUserNo             string  `json:"dsUserNo"`
	AuthPasswordHosted   int     `json:"authPasswordHosted"`
	ShowEnterpriseGuide  bool    `json:"showEnterpriseGuide"`
}
