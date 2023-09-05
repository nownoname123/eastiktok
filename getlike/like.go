package getlike

type Pointlike struct {
	ActionType string `json:"action_type"` // 1-点赞，2-取消点赞
	Token      string `json:"token"`       // 用户鉴权token
	VideoID    string `json:"video_id"`    // 视频id

}
type Respon struct {
	Status_code int    `json:"status_code"`
	Status_msg  string `json:"status_msg"`
}
