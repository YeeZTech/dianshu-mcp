package dianshu

// DatasetItem 典枢平台数据集列表项
type DatasetItem struct {
	ID                 int            `json:"id"`
	DatasetName        string         `json:"datasetName"`
	Tag                string         `json:"tag"`
	DatasetSize        int64          `json:"datasetSize"`
	Price              float64        `json:"price"`
	OriginalPrice      float64        `json:"originalPrice"`
	Pattern            string         `json:"pattern"`
	CreateCompanyName  string         `json:"createCompanyName"`
	CreateTime         int64          `json:"createTime"`
	SalesVolume        string         `json:"salesVolume"`
	PlatformTag        string         `json:"platformTag"`
	DatasetCode        string         `json:"datasetCode"`
	SourceType         int            `json:"sourceType"`
	IsAssessmentReport int            `json:"isAssessmentReport"`
	Description        *string        `json:"description"`
	ImageList          []DatasetImage `json:"imageList"`
	PatternFormat      string         `json:"patternFormat"`
	UniqueUserID       string         `json:"uniqueUserId"`
	UserID             int            `json:"userId"`
	DatasetActualSize  *int64         `json:"datasetActualSize"`
	DescriptionTxt     *string        `json:"descriptionTxt"`
	CreateTimeSQL      *string        `json:"createTimeSql"`
}

// DatasetImage 数据集图片
type DatasetImage struct {
	ImageURL   string `json:"imageUrl"`
	ImageOrder int    `json:"imageOrder"`
}

// DatasetSearchResponse 典枢平台数据集搜索响应
type DatasetSearchResponse struct {
	ResultCode int           `json:"resultCode"`
	ResultDesc string        `json:"resultDesc"`
	Data       []DatasetItem `json:"data"`
	Page       PageInfo      `json:"page"`
}

// RecommendDetail 首页推荐区块中的数据集条目
type RecommendDetail struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	Description   string  `json:"description"`
	ImageBgURL    string  `json:"imageBgUrl"`
	ImageFrontURL *string `json:"imageFrontUrl"`
	HrefURL       *string `json:"hrefUrl"`
	DatasetID     string  `json:"datasetId"`
	DatasetName   *string `json:"datasetName"`
	Price         *string `json:"price"`
	Pattern       *string `json:"pattern"`
	SecurityLevel *string `json:"securityLevel"`
}

// RecommendBlock 首页推荐区块
type RecommendBlock struct {
	ID      string            `json:"id"`
	Type    int               `json:"type"`
	Name    string            `json:"name"`
	Details []RecommendDetail `json:"details"`
}

// HomepageRecommendResponse 首页推荐 GraphQL 响应
type HomepageRecommendResponse struct {
	Data struct {
		QueryRecommend struct {
			ResultCode int              `json:"resultCode"`
			ResultDesc string           `json:"resultDesc"`
			Data       []RecommendBlock `json:"data"`
		} `json:"queryRecommend"`
	} `json:"data"`
}
