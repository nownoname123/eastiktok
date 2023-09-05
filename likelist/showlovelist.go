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
	"time"
)

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
func Favouritelist(w http.ResponseWriter, r *http.Request) {

	dsn := "root:123456@tcp(127.0.0.1:3306)/likelist?charset=utf8mb4&parseTime=True&loc=Local" //连接数据库
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

	err = db.AutoMigrate(&getlike.Pointlike{})
	if err != nil {
		return
	}

	//解析json

	var res Response
	//err = json.NewDecoder(r.Body).Decode(&reqdata)
	queryValues := r.URL.Query()
	token := queryValues.Get("token")
	//Id := queryValues.Get("user_id")
	w.Header().Set("Content-Type", "application/json")

	token1, _ := userlogin.CheckToken(token)
	var username string
	if token1.Valid {
		claims := token1.Claims.(jwt.MapClaims)
		username = claims["username"].(string)
		//expiration := claims["exp"].(float64)

	} else {
		res.StatusMsg = "未登录不可查看喜欢列表"
		res.StatusCode = "0"
		log.Println("Token is not valid")
		err = json.NewEncoder(w).Encode(res)
		if err != nil {
			http.Error(w, "response error", http.StatusBadRequest)
			return
		}
		return
	}
	var likedata []DataLike
	db.Where("username = ?", username).Find(&likedata)

	//根据likedata中的videoid获取视频信息
	dsnVideo := "root:123456@tcp(127.0.0.1:3306)/videolist?charset=utf8mb4&parseTime=True&loc=Local"
	dbVideo, err := gorm.Open(mysql.Open(dsnVideo), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			//解决查表的时候自动添加复数的问题，例如user变成了users
			SingularTable: true,
		},
	})
	err = dbVideo.AutoMigrate(&viedo.Video{})
	var v []Video
	for _, i := range likedata {
		var vv viedo.Video
		dbVideo.Where("id = ?", i.VideoId).First(&vv)
		var u userlogin.Usertype
		u = getUser(vv.Author)
		u1, _ := viedo.GetPlayUrl(vv.Title + ".jpg")
		u2, _ := viedo.GetPlayUrl(vv.Title + ".mp4")
		a := Video{
			Author:        u,
			CommentCount:  vv.CommentCount,
			CoverURL:      u1,
			FavoriteCount: vv.FavoriteCount,
			ID:            vv.ID,
			IsFavorite:    vv.IsFavorite,
			PlayURL:       u2,
			Title:         vv.Title,
		}
		v = append(v, a)
	}
	res.VideoList = v
	res.StatusCode = "0"
	err = json.NewEncoder(w).Encode(res)
	if err != nil {
		http.Error(w, "response error", http.StatusBadRequest)
		return
	}
}
