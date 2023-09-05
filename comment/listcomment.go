package comment

import (
	"demo-tiktok/writecomment"
	"encoding/json"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"net/http"
)

type Request struct {
	Token   string `json:"token"`    // 用户鉴权token
	VideoID string `json:"video_id"` // 视频id
}
type Response struct {
	CommentList []writecomment.CommentRes `json:"comment_list"` // 评论列表
	StatusCode  int64                     `json:"status_code"`  // 状态码，0-成功，其他值-失败
	StatusMsg   string                    `json:"status_msg"`   // 返回状态描述
}

var (
	db *gorm.DB
)

func linksql() {
	// 初始化数据库连接
	dsncomment := "root:123456@tcp(127.0.0.1:3306)/commentlist?charset=utf8mb4&parseTime=True&loc=Local"
	db, _ = gorm.Open(mysql.Open(dsncomment), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			//解决查表的时候自动添加复数的问题，例如user变成了users
			SingularTable: true,
		},
	})
	err := db.AutoMigrate(&writecomment.Comment{})
	if err != nil {
		return
	}

}
func Showcommentlist(w http.ResponseWriter, r *http.Request) {
	linksql()
	var req Request
	queryValues := r.URL.Query()
	req.Token = queryValues.Get("token")
	req.VideoID = queryValues.Get("video_id")
	w.Header().Set("Content-Type", "application/json")
	var comments []writecomment.Comment
	if err := db.Where("video_id = ?", req.VideoID).Find(&comments).Error; err != nil {
		response := Response{

			StatusCode: 1, // Successful status code
			StatusMsg:  "error",
		}
		err = json.NewEncoder(w).Encode(response)
		http.Error(w, "Failed to fetch comments", http.StatusInternalServerError)
		return
	}
	var com []writecomment.CommentRes
	for _, comment := range comments {
		c := writecomment.CommentRes{
			User:       comment.Usertype,
			Id:         comment.Id,
			Content:    comment.Content,
			CreateDate: comment.CreateDate,
		}
		com = append(com, c)
	}
	response := Response{
		CommentList: com,
		StatusCode:  0, // Successful status code
		//StatusMsg:   nil,
	}

	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

}
