package userlogin

import (
	"encoding/json"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"time"
)

type Reqdata struct {
	Password string `json:"password"` // 登录密码
	Username string `json:"username"` // 登录用户名
}
type Res struct {
	StatusCode int32  `json:"status_code"` // 状态码，0-成功，其他值-失败
	StatusMsg  string `json:"status_msg"`  // 返回状态描述
	Token      string `json:"token"`       // 用户鉴权token
	UserID     int64  `json:"user_id"`     // 用户id
}

func Userlogin(w http.ResponseWriter, r *http.Request) {
	db := InitDB()
	sqlDB, err := db.DB()
	if err != nil {
		panic("failed to connect database, err:" + err.Error())
	}

	sqlDB.SetMaxIdleConns(10)

	sqlDB.SetMaxOpenConns(100)

	sqlDB.SetConnMaxLifetime(10 * time.Second)

	queryValues := r.URL.Query()
	username := queryValues.Get("username")
	password := queryValues.Get("password")

	var user Usertype
	db.Where("username = ?", username).First(&user)

	w.Header().Set("Content-Type", "application/json") //设置响应头
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		st := "密码错误"
		response := Res{
			StatusCode: 2,
			StatusMsg:  st,
			//Token:      nil,
			//UserID:     nil,
		}
		err = json.NewEncoder(w).Encode(response)
		return
	}
	st := "密码正确"
	idx := user.ID
	log.Println(st)
	token, err := GetToken(username, user.ID)
	if err != nil {
		log.Println("login Gettoken error")
		return
	}
	response := Res{
		StatusCode: 0,
		StatusMsg:  st,
		Token:      token,
		UserID:     idx,
	}

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("登陆接口执行完毕")

}
