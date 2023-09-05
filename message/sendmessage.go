package message

import (
	"demo-tiktok/userlogin"
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"log"
	"net/http"
	"strconv"
	"time"
)

type Req struct {
	ActionType string `json:"action_type"` // 1-发送消息
	Content    string `json:"content"`     // 消息内容
	ToUserID   string `json:"to_user_id"`  // 对方用户id
	Token      string `json:"token"`       // 用户鉴权token
}
type Response struct {
	StatusCode int64  `json:"status_code"` // 状态码，0-成功，其他值-失败
	StatusMsg  string `json:"status_msg"`  // 返回状态描述
}
type Message struct {
	Content    string `json:"content"`      // 消息内容
	CreateTime int64  `json:"create_time"`  // 消息发送时间 yyyy-MM-dd HH:MM:ss
	FromUserID int64  `json:"from_user_id"` // 消息发送者id
	ID         int64  `json:"id"`           // 消息id
	ToUserID   int64  `json:"to_user_id"`   // 消息接收者id
}

func Sendmessage(w http.ResponseWriter, r *http.Request) {

	dsn := "root:123456@tcp(127.0.0.1:3306)/message?charset=utf8mb4&parseTime=True&loc=Local" //连接数据库
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

	err = db.AutoMigrate(&Message{})

	var reqdata Req
	var res Response

	queryValues := r.URL.Query()
	reqdata.Content = queryValues.Get("content")
	reqdata.ToUserID = queryValues.Get("to_user_id")
	log.Println("to user id :", reqdata.ToUserID)
	token := queryValues.Get("token")
	reqdata.ActionType = queryValues.Get("action_type")

	token1, err := userlogin.CheckToken(token)
	if err != nil {
		log.Println(err)
		log.Println("check token error in viedopublish")
		return
	}
	var userid int64
	if token1.Valid {
		claims := token1.Claims.(jwt.MapClaims)
		log.Println("11111")
		k := claims["userid"].(float64)
		userid = int64(k)
		//expiration := claims["exp"].(float64)
		log.Println("222222")

	} else {
		log.Println("Token is not valid")
		res.StatusCode = 1
		res.StatusMsg = "get userid error"
		err = json.NewEncoder(w).Encode(res)
		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	var msg Message
	msg.FromUserID = userid
	msg.Content = reqdata.Content
	val, err := strconv.Atoi(reqdata.ToUserID)
	if err != nil {
		res.StatusCode = 1
		res.StatusMsg = "toUser ID error"
		err = json.NewEncoder(w).Encode(res)
		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}
	msg.ToUserID = int64(val)

	currentTime := time.Now()
	// 格式化时间为 "yyyy-MM-dd HH:MM:ss" 形式的字符串
	createTimeStr := currentTime.Format("2006-01-02 15:04:05")

	// 将时间字符串解析为对应的时间对象，以便后续操作
	parsedTime, err := time.Parse("2006-01-02 15:04:05", createTimeStr)
	if err != nil {
		fmt.Println("Error parsing time:", err)
		return
	}

	// 获取 Unix 时间戳（秒数）表示
	unixTimestamp := parsedTime.Unix()

	msg.CreateTime = unixTimestamp

	lasttime = unixTimestamp
	db.Create(&msg)
	res.StatusCode = 0
	res.StatusMsg = "消息发送成功"
	err = json.NewEncoder(w).Encode(res)
	if err != nil {

		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}
}
