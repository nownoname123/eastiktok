package userlogin

import (
	"encoding/json"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"math/rand"
	"net/http"
	"time"
)

var _ int64

type Usertype struct {
	//gorm.Model

	Username string `gorm:"varchar(32);not null;unique" json:"name"` //用户名称
	Password string `gorm:"not null" json:"password"`
	//Token           string `gorm:"not null" json:"token"`
	ID              int64  `gorm:"primaryKey;autoIncrement" json:"user_id"` //用户id  用自增主键
	Avatar          string `json:"avatar"`                                  // 用户头像
	BackgroundImage string `json:"background_image"`                        // 用户个人页顶部大图
	FavoriteCount   int64  `json:"favorite_count"`                          // 喜欢数
	FollowCount     int64  `json:"follow_count"`                            // 关注总数
	FollowerCount   int64  `json:"follower_count"`                          // 粉丝总数

	IsFollow bool `json:"is_follow"` // true-已关注，false-未关注

	Signature      string `json:"signature"`       // 个人简介
	TotalFavorited string `json:"total_favorited"` // 获赞数量
	WorkCount      int64  `json:"work_count"`      // 作品数
}

type Requestdata struct {
	Password string `json:"password"` // 密码，最长32个字符
	Username string `json:"username"` // 注册用户名，最长32个字符
}
type Response struct {
	StatusCode int64  `json:"status_code"` // 状态码，0-成功，其他值-失败
	StatusMsg  string `json:"status_msg"`  // 返回状态描述
	Token      string `json:"token"`       // 用户鉴权token
	UserID     int64  `json:"user_id"`     // 用户id
}

func InitDB() *gorm.DB {
	dsn := "root:123456@tcp(127.0.0.1:3306)/userlist?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		//Logger: logger.Default.LogMode(logger.Info),记录日志
		Logger: logger.Default.LogMode(logger.Silent), //禁用日志记录
	})
	if err != nil {
		log.Println("line:54 :: ", err)
		return nil
	}

	//迁移

	err = db.AutoMigrate(&Usertype{})
	if err != nil {

		log.Println("数据库迁移错误")
		return nil
	}

	return db

}
func Register(w http.ResponseWriter, r *http.Request) {

	db := InitDB()
	sqlDB, err := db.DB()
	if err != nil {
		log.Println("failed to connect database, err:", err)
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}

	sqlDB.SetMaxIdleConns(10)

	sqlDB.SetMaxOpenConns(100)

	sqlDB.SetConnMaxLifetime(10 * time.Second)

	//var reqData Requestdata
	queryValues := r.URL.Query()
	name := queryValues.Get("username")
	password := queryValues.Get("password")
	/*	err = json.NewDecoder(r.Body).Decode(&reqData)
		if err != nil && err != io.EOF {
			log.Println(err)
			http.Error(w, "Failed to decode request body", http.StatusBadRequest)
			return
		}
		log.Println(reqData)
		defer r.Body.Close()*/

	w.Header().Set("Content-Type", "application/json") //设置响应头

	//数据验证
	if len(name) == 0 {
		response := Response{
			StatusCode: 1,
			StatusMsg:  "用户名为空",
		}

		err = json.NewEncoder(w).Encode(response)

		return
	}

	if len(password) < 8 {
		response := Response{
			StatusCode: 1,
			StatusMsg:  "密码不能少于8位",
		}
		err = json.NewEncoder(w).Encode(response)
		return
	}

	hasedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		response := Response{
			StatusCode: 1,
			StatusMsg:  "密码加密错误",
		}
		err = json.NewEncoder(w).Encode(response)
		return
	} //密码加密，用哈希算法进行密码加密

	ntr := rand.Intn(5)
	var head [6]string //头像地址

	head[0] = "https://cn.bing.com/images/search?view=detailV2&ccid=8Ka%2fQr0U&id=CD9C6AB1F9A394EB6BBB1E56267CC8A380700988&thid=OIP.8Ka_Qr0UZpBWSu1UF5RcQAHaHa&mediaurl=https%3a%2f%2fpic2.zhimg.com%2fv2-2d44b34343fadb3f01872fa244580bc1_r.jpg&exph=700&expw=700&q=%e4%ba%8c%e6%ac%a1%e5%85%83%e5%a4%b4%e5%83%8f%e5%a4%a7%e5%85%a8&simid=608008520324423137&FORM=IRPRST&ck=7C4805037328343D3156714186512B3E&selectedIndex=9"
	head[1] = "https://cn.bing.com/images/search?view=detailV2&ccid=ZeHJt4kY&id=88B1F3351D0F698B3806D2DE980A910EBCB7B01D&thid=OIP.ZeHJt4kYjExfeHk0cSEybwHaHa&mediaurl=https%3a%2f%2fimg.zcool.cn%2fcommunity%2f01e8ce5d89c4eaa8012060bed28708.png%402o.png&exph=2000&expw=2000&q=%e4%ba%8c%e6%ac%a1%e5%85%83%e5%a4%b4%e5%83%8f%e5%a4%a7%e5%85%a8&simid=607995360541499991&FORM=IRPRST&ck=F788604851AF30E61456499FB5A0A59F&selectedIndex=16"
	head[2] = "https://cn.bing.com/images/search?view=detailV2&ccid=kwBQODCZ&id=39EA1E612A979AC08F88B3DA09A1AF069DDE0D14&thid=OIP.kwBQODCZLDq2v2XpXcGxsQHaHa&mediaurl=https%3a%2f%2fpic1.zhimg.com%2fv2-7dc0fb227cc67e12f186b19858356567_r.jpg%3fsource%3d1940ef5c&exph=1080&expw=1080&q=%e4%ba%8c%e6%ac%a1%e5%85%83%e5%a4%b4%e5%83%8f%e5%a4%a7%e5%85%a8&simid=608046973159152076&FORM=IRPRST&ck=0A78B7F2274110AE0B09C655075037FA&selectedIndex=32"
	head[3] = "https://ts1.cn.mm.bing.net/th/id/R-C.abd68212f721eb2546840ba4735aa562?rik=WRkygXN7Qfgfeg&riu=http%3a%2f%2fimg4.a0bi.com%2fupload%2fttq%2f20200724%2f1595558809098.jpg%3fimageView2%2f0%2fw%2f600%2fh%2f800&ehk=OVWrnvEj%2b3opKgE9YBHy7BpWkSmaE0P0Lz2HTeijO7o%3d&risl=&pid=ImgRaw&r=0"
	head[4] = "https://img.zcool.cn/community/01e8ce5d89c4eaa8012060bed28708.png@2o.png"
	head[5] = "https://pic4.zhimg.com/v2-77fa45645098cec0aa17f57daffd7881_r.jpg"

	newUser := Usertype{
		Username:        name,
		Password:        string(hasedPassword),
		FollowCount:     0,
		FollowerCount:   0,
		IsFollow:        false,
		Signature:       "这个人很懒，还没有自我介绍",
		WorkCount:       0,
		Avatar:          head[ntr],
		BackgroundImage: "no",
		TotalFavorited:  "0",
		FavoriteCount:   0,
	}
	db.Create(&newUser)
	token, err := GetToken(name, newUser.ID)
	if err != nil {
		log.Println("get-token error")
		return
	}

	//返回结果

	response := Response{
		StatusCode: 0,
		StatusMsg:  "用户注册成功",
		Token:      token,
		UserID:     newUser.ID,
	}
	err = json.NewEncoder(w).Encode(response)

}
