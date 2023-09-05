package pushlist

import (
	"demo-tiktok/userlogin"
	"demo-tiktok/viedo"
	"encoding/json"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"log"
	"net/http"
	"strconv"
)

type douyinPublishListRequest struct {
	UserID string `json:"user_id" `
	Token  string `json:"token" `
}

type Res struct {
	StatusCode int64   `json:"status_code"` // 状态码，0-成功，其他值-失败
	StatusMsg  *string `json:"status_msg"`  // 返回状态描述
	VideoList  []Video `json:"video_list"`  // 用户发布的视频列表
}

type Video struct {
	ID            int64              `json:"id"`
	Author        userlogin.Usertype `json:"author"`
	PlayURL       string             `json:"play_url"`
	CoverURL      string             `json:"cover_url"`
	FavoriteCount int64              `json:"favorite_count"`
	CommentCount  int64              `json:"comment_count"`
	IsFavorite    bool               `json:"is_favorite"`
	Title         string             `json:"title"`
}

type FavoriteModel struct {
	gorm.Model
	UserID  int64
	VideoID int64
}

func Pushlist(w http.ResponseWriter, r *http.Request) {

	var request douyinPublishListRequest
	queryValues := r.URL.Query()
	request.Token = queryValues.Get("token")
	request.UserID = queryValues.Get("user_id")

	w.Header().Set("Content-Type", "application/json") //设置响应头

	// 根据用户ID查询用户的视频发布列表
	userID, err := strconv.Atoi(request.UserID)
	if err != nil {
		log.Println("字符串userID转化数字失败")
		return
	}
	num := int64(userID)
	videoList, err := queryVideoListByUserID(num)
	if err != nil {
		http.Error(w, "request error", http.StatusBadRequest)
		return
	}
	response := Res{
		StatusCode: 0,
		VideoList:  videoList,
	}

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "response error", http.StatusBadRequest)
		return
	}
}

func queryVideoListByUserID(userID int64) ([]Video, error) {

	dsn := "root:123456@tcp(127.0.0.1:3306)/videolist?charset=utf8mb4&parseTime=True&loc=Local" //连接数据库
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			//解决查表的时候自动添加复数的问题，例如user变成了users
			SingularTable: true,
		},
	})

	err = db.AutoMigrate(&viedo.Video{})
	if err != nil {
		return nil, err
	}

	var videoModels []viedo.Video
	result := db.Where("author = ?", userID).Find(&videoModels)

	//查询数据库
	if result.Error != nil {
		return nil, result.Error
	}

	var videoList []Video
	for _, videoModel := range videoModels {
		var user userlogin.Usertype
		user = getUser(videoModel.Author)
		u, _ := viedo.GetPlayUrl(videoModel.Title)
		log.Println("在list中拿到的预前面URL为：", u)
		video := Video{
			ID: videoModel.ID,

			PlayURL:       u,
			CoverURL:      videoModel.CoverURL,
			FavoriteCount: 0,     //int64(len(videoModel.FavoriteList)),
			CommentCount:  0,     // TODO: 查询评论数量
			IsFavorite:    false, // TODO: 判断当前用户是否点赞
			Title:         videoModel.Title,
		}
		video.Author = user
		log.Println(user.Username, " 发了：", video.Title)
		videoList = append(videoList, video)
	}

	return videoList, nil
}
func getUser(userID int64) userlogin.Usertype {
	dsn := "root:123456@tcp(127.0.0.1:3306)/userlist?charset=utf8mb4&parseTime=True&loc=Local" //连接数据库
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Println("usermessage error in db")

	}
	db.AutoMigrate(&userlogin.Usertype{})
	var user userlogin.Usertype
	result := db.Where("id = ?", userID).Find(&user)
	if result.Error != nil {
		log.Println("find author error in listpublish error is :", result.Error)
	}
	return user
}
