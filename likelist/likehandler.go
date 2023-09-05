package likelist

import (
	"demo-tiktok/getlike"
	"demo-tiktok/userlogin"
	"demo-tiktok/viedo"
	"encoding/json"
	"github.com/dgrijalva/jwt-go"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"log"
	"net/http"
	"strconv"
	"time"
)

type DataLike struct {
	UserName string `gorm:"column:username"`
	VideoId  int64  `gorm:"column:videoid"`
}

// AddFavorite 给用户标记喜欢或者不喜欢
func AddFavorite(username string, t bool) {
	dsn := "root:123456@tcp(127.0.0.1:3306)/userlist?charset=utf8mb4&parseTime=True&loc=Local" //连接数据库
	db, _ := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), //禁用日志记录

	})

	err := db.AutoMigrate(&userlogin.Usertype{})
	if err != nil {
		log.Println("db error in likehandler")

	}
	var user userlogin.Usertype
	result := db.Where("username = ?", username).First(&user)
	if result.Error != nil {
		log.Println("get user error in likehandler :", result.Error)
		return
	}
	//对用户喜欢的作品数进行修改
	if t {
		user.FavoriteCount++

	} else {
		if user.FavoriteCount != 0 {
			user.FavoriteCount--
		}
	}
	if err := db.Save(&user).Error; err != nil {
		log.Println("save user error in likehandler :", err)
	}
	return

}
func Addlike(username string, videoId int64) {
	dsn := "root:123456@tcp(127.0.0.1:3306)/likelist?charset=utf8mb4&parseTime=True&loc=Local" //连接数据库
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			//解决查表的时候自动添加复数的问题，例如user变成了users
			SingularTable: true,
		},
	})
	err = db.AutoMigrate(&DataLike{})
	if err != nil {
		return
	}
	a := DataLike{
		UserName: username,
		VideoId:  videoId,
	}
	db.Create(a)
	return
}
func Dellike(username string, videoId int64) {
	dsn := "root:123456@tcp(127.0.0.1:3306)/likelist?charset=utf8mb4&parseTime=True&loc=Local" //连接数据库
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			//解决查表的时候自动添加复数的问题，例如user变成了users
			SingularTable: true,
		},
	})
	err = db.AutoMigrate(&DataLike{})
	if err != nil {
		return
	}
	var a DataLike
	db.Where("username = ? AND videoid = ?", username, videoId).Delete(&a)
	log.Println("已删除喜欢列表中的：", a.UserName, "  喜欢：", a.VideoId)
	return
}

// LikeHandler 处理点赞请求的处理函数
func LikeHandler(w http.ResponseWriter, r *http.Request) {
	dsn := "root:123456@tcp(127.0.0.1:3306)/videolist?charset=utf8mb4&parseTime=True&loc=Local" //连接数据库
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			//解决查表的时候自动添加复数的问题，例如user变成了users
			SingularTable: true,
		},
	})

	sqlDB, err := db.DB()

	sqlDB.SetMaxIdleConns(10)

	sqlDB.SetMaxOpenConns(100)

	sqlDB.SetConnMaxLifetime(10 * time.Second)

	err = db.AutoMigrate(&viedo.Video{})
	if err != nil {
		return
	}

	// 解析 JSON 数据
	//var reqData getlike.Pointlike
	//err = json.NewDecoder(r.Body).Decode(&reqData)
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

	// 判断点赞类型
	if ActionType == "1" {

		AddFavorite(username, true) //添加用户喜欢总数
		Addlike(username, videoID)  //在数据库中记录喜欢关系
		video.FavoriteCount++
		video.IsFavorite = true
		if err := db.Save(&video).Error; err != nil {
			log.Println("save video error in likehandler :", err)
			return
		}
		response := getlike.Respon{
			Status_code: 0,
			Status_msg:  "Liked successfully",
		}
		err := json.NewEncoder(w).Encode(response)
		if err != nil {
			return
		}
	} else if ActionType == "2" {
		AddFavorite(username, false) //删除用户在点赞数据库中的数据
		Dellike(username, videoID)   //在likelist中删除喜欢关系
		video.FavoriteCount--
		video.IsFavorite = false
		if err := db.Save(&video).Error; err != nil {
			log.Println("save video error in likehandler second :", err)
			return
		}
		response := getlike.Respon{
			Status_code: 0,
			Status_msg:  "unLiked successfully",
		}

		err := json.NewEncoder(w).Encode(response)
		//json 包来编码 response 变量中的数据为 JSON 格式，
		//并将编码后的 JSON 数据写入到 w（http.ResponseWriter）中，即将 JSON 数据作为 HTTP 响应体发送回客户端。

		if err != nil {
			return

		}
	} else {
		// 如果传入的点赞类型不正确，返回错误响应
		response := getlike.Respon{
			Status_code: 1,
			Status_msg:  "reading error",
		}

		err := json.NewEncoder(w).Encode(response)
		if err != nil {
			return

		}
		http.Error(w, "Invalid action_type", http.StatusBadRequest)
	}

}
