package message

import (
	"demo-tiktok/userlogin"
	"encoding/json"
	"github.com/dgrijalva/jwt-go"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"log"
	"net/http"
	"time"
)

type Request struct {
	ToUserID string `json:"to_user_id"` // 对方用户id
	Token    string `json:"token"`      // 用户鉴权token
}
type Res struct {
	MessageList []Message `json:"message_list"` // 用户列表
	StatusCode  string    `json:"status_code"`  // 状态码，0-成功，其他值-失败
	StatusMsg   string    `json:"status_msg"`   // 返回状态描述
}

var lasttime int64

func Messagelist(w http.ResponseWriter, r *http.Request) {
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

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		return
	}

	var res Res
	queryValues := r.URL.Query()
	token := queryValues.Get("token")
	ToUserID := queryValues.Get("to_user_id")
	token1, err := userlogin.CheckToken(token)
	if err != nil {
		log.Println(err)
		log.Println("check token error in viedopublish")
		return
	}
	var userid int64
	if token1.Valid {
		claims := token1.Claims.(jwt.MapClaims)
		k := claims["userid"].(float64)
		userid = int64(k)
		//log.Println("userid is:", userid)

	} else {
		log.Println("Token is not valid")
		res.StatusCode = "1"
		res.StatusMsg = "get userid error"
		err = json.NewEncoder(w).Encode(res)
		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}
	var messages []Message
	var lastmsg []Message //找对方的聊天记录

	// 构建查询条件，按照 CreateTime 降序排列
	log.Println(userid, " and ", ToUserID)
	result := db.Where("to_user_id = ? AND from_user_id = ? And create_time > ?", userid, ToUserID, lasttime).Order("create_time DESC").Find(&lastmsg)
	result = db.Where("from_user_id = ? AND to_user_id = ? And create_time > ?", userid, ToUserID, lasttime).Order("create_time DESC").Find(&messages)
	length := len(messages)
	if length != 0 {
		lasttime = messages[length-1].CreateTime //更新时间节点，作为下一次提取消息的开始
	}

	if result.Error != nil {
		res.StatusCode = "1"
		st := "get list error"
		res.StatusMsg = st
		err = json.NewEncoder(w).Encode(res)
		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}
	res.MessageList = lastmsg
	for _, i := range messages {
		res.MessageList = append(res.MessageList, i)
	}
	res.StatusCode = "0"

	err = json.NewEncoder(w).Encode(res)
	if result.Error != nil {

		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}
	return
}
