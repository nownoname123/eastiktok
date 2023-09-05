// Package viedo 视频流接口
package viedo

import (
	"demo-tiktok/userlogin"
	"encoding/json"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"log"
	"net/http"
	"time"
)

type VideoType struct {
	Author        User1  `json:"author"` // 视频作者信息
	PlayUrl       string `json:"play_url" json:"play_url,omitempty"`
	CoverUrl      string `json:"cover_url,omitempty"`
	FavoriteCount int64  `json:"favorite_count,omitempty"`
	CommentCount  int64  `json:"comment_count,omitempty"`
	IsFavorite    bool   `json:"is_favorite,omitempty"`
	Id            int64  `json:"id,omitempty"` // 视频唯一标识

	//	Title         string             `json:"title"`          // 视频标题
}
type User1 struct {
	Id            int64  `json:"id,omitempty"`
	Name          string `json:"name,omitempty"`
	FollowCount   int64  `json:"follow_count,omitempty"`
	FollowerCount int64  `json:"follower_count,omitempty"`
	IsFollow      bool   `json:"is_follow,omitempty"`
}
type Request struct {
	LatestTime string `json:"latest_time,omitempty"` // 可选参数，限制返回视频的最新投稿时间戳，精确到秒，不填表示当前时间
	Token      string `json:"token,omitempty"`       // 用户登录状态下设置
}

func getUser(userID int64) userlogin.Usertype {
	dsn := "root:123456@tcp(127.0.0.1:3306)/userlist?charset=utf8mb4&parseTime=True&loc=Local" //连接数据库
	db, _ := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), //禁用日志记录

	})

	err := db.AutoMigrate(&userlogin.Usertype{})
	if err != nil {
		log.Println("db error in feed")

	}
	var user userlogin.Usertype
	log.Println("we are here")
	result := db.Where("id = ?", userID).First(&user) //查找作者信息
	log.Println("user msg is:", user.ID)
	if result.Error != nil {
		log.Println("dbuser error in feed:", result.Error)
		return user
	}
	return user
}
func Feed(w http.ResponseWriter, r *http.Request) {
	dsn := "root:123456@tcp(127.0.0.1:3306)/videolist?charset=utf8mb4&parseTime=True&loc=Local" //连接数据库
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), //禁用日志记录
		NamingStrategy: schema.NamingStrategy{
			//解决查表的时候自动添加复数的问题，例如user变成了users
			SingularTable: true,
		},
	})

	err = db.AutoMigrate(&Video{})
	if err != nil {
		log.Println("db error")
		return
	}

	var reqdata Request
	queryValues := r.URL.Query()
	reqdata.LatestTime = queryValues.Get("latest_time")
	reqdata.Token = queryValues.Get("token")

	//获取视频
	var videos []Video
	var ans []VideoType

	err = db.Order("creat_time desc").Limit(30).Find(&videos).Error

	if err != nil {
		log.Println("err in feed ", err)
		return
		//panic("failed to retrieve videos, err: " + err.Error())
	}
	log.Println("this :", videos[0].Author)
	for _, video := range videos {

		user := getUser(video.Author)

		var it VideoType
		it.Author.Id = user.ID
		it.Author.Name = user.Username
		it.Author.IsFollow = user.IsFollow
		it.Author.FollowerCount = user.FollowerCount
		it.Author.FollowCount = user.FollowCount
		it.CommentCount = video.CommentCount
		it.CoverUrl, _ = GetPlayUrl(video.Title + ".jpg")
		it.FavoriteCount = video.FavoriteCount
		it.Id = video.ID
		it.IsFavorite = video.IsFavorite
		it.PlayUrl, _ = GetPlayUrl(video.Title + ".mp4")
		//it.Title = video.Title

		ans = append(ans, it)
		log.Println(user.Username, " 发了：", video.Title, "url:", it.PlayUrl)
	}

	// 设置返回头
	w.Header().Set("Content-Type", "application/json")
	//statumsg := "success"
	response := Response{
		NextTime:   time.Now().Unix(),
		StatusCode: 0,
		//StatusMsg:  statumsg,
		VideoList: ans,
	}
	log.Println("feed返回成功")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Println("with error in feed's response:", err)
		http.Error(w, "Error writing video file", http.StatusInternalServerError)
		return
	}

}
