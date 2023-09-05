package writecomment

import (
	"demo-tiktok/getlike"
	"demo-tiktok/userlogin"
	"demo-tiktok/viedo"
	"encoding/json"
	"github.com/dgrijalva/jwt-go"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"log"
	"net/http"
	"strconv"
	"time"
)

type Request struct {
	ActionType  string  `json:"action_type"`                                          // 1-发布评论，2-删除评论
	CommentID   *string `gorm:"primaryKey;autoIncrement" json:"comment_id,omitempty"` // 要删除的评论id，在action_type=2的时候使用
	CommentText *string `json:"comment_text,omitempty"`                               // 用户填写的评论内容，在action_type=1的时候使用
	Token       string  `json:"token"`                                                // 用户鉴权token
	VideoID     string  `json:"video_id"`                                             // 视频id
}
type Comment struct {
	VideoID int64 `json:"video_id"`
	Id      int64 `json:"id,omitempty"` // 评论id
	//User       userlogin.Usertype `json:"user"`                  // 评论用户信息
	userlogin.Usertype        // 评论用户信息
	Content            string `json:"content,omitempty"`     //评论内容
	CreateDate         string `json:"create_date,omitempty"` // 评论发布日期，格式 mm-dd`
}
type CommentRes struct {
	//	VideoID int64              `json:"video_id"`
	Id   int64              `json:"id,omitempty"` // 评论id
	User userlogin.Usertype `json:"user"`         // 评论用户信息

	Content    string `json:"content,omitempty"`     //评论内容
	CreateDate string `json:"create_date,omitempty"` // 评论发布日期，格式 mm-dd`
}

// 评论用户信息

type Res struct {
	Comment    CommentRes `json:"comment"` // 评论成功返回评论内容，不需要重新拉取整个列表
	StatusCode int32      `json:"status_code"`
	StatusMsg  string     `json:"status_msg,omitempty"`
}

var (
	db        *gorm.DB
	dbUser    *gorm.DB
	dbcomment *gorm.DB
)

func linksql() {
	// 初始化数据库连接
	dsn := "root:123456@tcp(127.0.0.1:3306)/videolist?charset=utf8mb4&parseTime=True&loc=Local" //连接数据库
	db, _ = gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			//解决查表的时候自动添加复数的问题，例如user变成了users
			SingularTable: true,
		},
	})
	dsnUser := "root:123456@tcp(127.0.0.1:3306)/userlist?charset=utf8mb4&parseTime=True&loc=Local"

	dbUser, _ = gorm.Open(mysql.Open(dsnUser), &gorm.Config{})
	sqlDB, _ := db.DB()

	sqlDB.SetMaxIdleConns(10)

	sqlDB.SetMaxOpenConns(100)

	sqlDB.SetConnMaxLifetime(10 * time.Second)
	err := db.AutoMigrate(&viedo.Video{})
	if err != nil {
		return
	}
	sqlDB, _ = dbUser.DB()

	sqlDB.SetMaxIdleConns(10)

	sqlDB.SetMaxOpenConns(100)

	sqlDB.SetConnMaxLifetime(10 * time.Second)
	err = dbUser.AutoMigrate(&userlogin.Usertype{})
	if err != nil {
		return
	}

	dsncomment := "root:123456@tcp(127.0.0.1:3306)/commentlist?charset=utf8mb4&parseTime=True&loc=Local"
	dbcomment, _ = gorm.Open(mysql.Open(dsncomment), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			//解决查表的时候自动添加复数的问题，例如user变成了users
			SingularTable: true,
		},
	})
	err = dbcomment.AutoMigrate(&Comment{})
	if err != nil {
		return
	}

}
func Write(w http.ResponseWriter, r *http.Request) {
	linksql()
	var req Request
	//err := json.NewDecoder(r.Body).Decode(&req)
	queryValues := r.URL.Query()
	token := queryValues.Get("token")
	Id := queryValues.Get("video_id")
	ActionType := queryValues.Get("action_type")

	w.Header().Set("Content-Type", "application/json") //设置Http响应头，它告诉客户端接收到的数据是以 JSON 格式表示的。

	//将数据库中的视频数据提取出来
	var videoID int64
	i, _ := strconv.Atoi(Id)
	videoID = int64(i)
	var video viedo.Video
	result := db.Where("id = ?", videoID).First(&video)
	if result.Error != nil {
		log.Println("get video error in likehandler")
		response := getlike.Respon{
			Status_code: 1,
			Status_msg:  "error",
		}

		_ = json.NewEncoder(w).Encode(response)
		return
	}

	//验证token
	token1, err := userlogin.CheckToken(token)
	if err != nil {
		log.Println(err)
		log.Println("check token error in viedopublish")
		return
	}
	var username string
	if token1.Valid {
		claims := token1.Claims.(jwt.MapClaims)
		username = claims["username"].(string)
		//expiration := claims["exp"].(float64)

	} else {
		log.Println("Token is not valid")
	}
	w.Header().Set("Content-Type", "application/json") //设置Http响应头，它告诉客户端接收到的数据是以 JSON 格式表示的。

	if ActionType == "1" {
		text := queryValues.Get("comment_text")
		var user userlogin.Usertype
		dbUser.Where("username = ?", username).Find(&user)
		var com Comment
		com = Comment{
			//VideoID:    123,
			//Content:    "This is a comment.",
			//CreateDate: "08-29",
			Usertype: user, // 直接赋值给匿名字段
		}
		com.Content = text

		currentTime := time.Now()
		formattedDate := currentTime.Format("01-02")
		com.CreateDate = formattedDate
		com.VideoID = video.ID
		// 将评论放入数据库中
		dbcomment.Create(&com)

		msg := "success"
		cc := CommentRes{
			User:       user,
			Id:         com.Id,
			Content:    com.Content,
			CreateDate: com.CreateDate,
		}
		res := Res{
			Comment:    cc,
			StatusCode: 0,
			StatusMsg:  msg,
		}
		err = json.NewEncoder(w).Encode(res)
		if err != nil {

			return
		}
		return
	} else if req.ActionType == "2" {
		var com Comment
		CommentID := queryValues.Get("comment_id")
		dbcomment.Where("id = ? ", CommentID).Delete(&com)
		msg := "success"
		res := Res{
			//Comment:    nil,
			StatusCode: 0,
			StatusMsg:  msg,
		}
		err = json.NewEncoder(w).Encode(res)
		if err != nil {

			return
		}
		return
	}
}
