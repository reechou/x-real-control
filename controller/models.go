package controller

const (
	CONTENT_TYPE_VIDEO = iota
)

const (
	CONTENT_T_NORMAL = iota
	CONTENT_T_ADS
)

const (
	DOMAIN_GROUP_TYPE_SHOW = iota
	DOMAIN_GROUP_TYPE_JUMP
)

type DomainGroupInfo struct {
	ID            int64   `json:"id"`
	Name          string  `json:"name"`
	Status        int64   `json:"status"`
	ShareStatus   int64   `json:"shareStatus"`
	AdsStatus     int64   `json:"adsStatus"`
	Type          int64   `json:"type"`
	ShowGroupList []int64 `json:"showGroupList"`
	ShowListStr   string  `json:"showGroupListStr"`
	Time          string  `json:"time"`
}

type DomainInfo struct {
	ID          int64  `json:"id"`
	GroupID     int64  `json:"groupID"`
	Domain      string `json:"domain"`
	Status      int64  `json:"status"`
	ShowGroupID int64  `json:"showGroupID"`
	Time        string `json:"time"`
}

type DomainList struct {
	GroupID    int64         `json:"groupID"`
	DomainList []*DomainInfo `json:"domainList"`
	UpdateTime int64
}

type ContentGroupInfo struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	JsonUrl     string  `json:"jsonUrl"`
	Type        int64   `json:"type"`
	MainContent []int64 `json:"mainContent"`
	Time        string  `json:"time"`
	UpdateTime  int64
}

type Video struct {
	Content  string `json:"content"`
	ID       int64  `json:"id"`
	Title    string `json:"title"`
	VideoSrc string `json:"videoSrc"`
	ImageUrl string `json:"imageUrl"`
	TitleImg string `json:"titleImg"`
	Type     int64  `json:"type"`
}

type ContentInfo struct {
	ID      int64  `json:"id"`
	GroupID int64  `json:"groupID"`
	Value   string `json:"value"`
	Type    int64  `json:"type"`
	Time    string `json:"time"`
}

type ContentList struct {
	GroupID     int64          `json:"groupID"`
	ContentList []*ContentInfo `json:"contentList"`
	UpdateTime  int64
}

type RealContentInfo struct {
	ContentGroupID int64  `json:"contentGroupID"`
	ContentUrl     string `json:"contentUrl"`
	IfOffLine      bool   `json:"ifOffLine"`
	IfForceShare   bool   `json:"ifForceShare"`
	IfShowAds      bool   `json:"ifShowAds"`
}

const (
	RES_OK = iota
	RES_ERR
)

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}
