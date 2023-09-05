package user

import (
	"demo-tiktok/userlogin"
	"encoding/json"

	"github.com/dgrijalva/jwt-go"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"net/http"
)

type DouyinUserRequest struct {
	UserID string `json:"user_id"`
	Token  string `json:"token"`
}

type DouyinUserResponse struct {
	StatusCode int32   `json:"status_code"` // 状态码，0-成功，其他值-失败
	StatusMsg  *string `json:"status_msg"`  // 返回状态描述
	User       User    `json:"user"`        // 用户信息
}

type User struct {
	Avatar          string `json:"avatar"`           // 用户头像
	BackgroundImage string `json:"background_image"` // 用户个人页顶部大图
	FavoriteCount   int64  `json:"favorite_count"`   // 喜欢数
	FollowCount     int64  `json:"follow_count"`     // 关注总数
	FollowerCount   int64  `json:"follower_count"`   // 粉丝总数
	ID              int64  `json:"id"`               // 用户id
	IsFollow        bool   `json:"is_follow"`        // true-已关注，false-未关注
	Name            string `json:"name"`             // 用户名称
	Signature       string `json:"signature"`        // 个人简介
	TotalFavorited  string `json:"total_favorited"`  // 获赞数量
	WorkCount       int64  `json:"work_count"`       // 作品数
}

func Usermessage(w http.ResponseWriter, r *http.Request) {
	// 连接数据库
	dsn := "root:123456@tcp(127.0.0.1:3306)/userlist?charset=utf8mb4&parseTime=True&loc=Local" //连接数据库
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Println("usermessage error in db")
		return
	}
	var reqdata DouyinUserRequest
	/*err = json.NewDecoder(r.Body).Decode(&reqdata) //将接受到的请求解析到reqData中
	if err != nil {

		return
	}*/
	queryValues := r.URL.Query()
	reqdata.UserID = queryValues.Get("user_id")
	reqdata.Token = queryValues.Get("token")
	token := reqdata.Token
	log.Println("token is:", token, "id is", reqdata.UserID)
	var user userlogin.Usertype
	db.AutoMigrate(&userlogin.Usertype{})

	token1, err := userlogin.CheckToken(token)
	if err != nil {
		log.Println("Token validation failed:", err)
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

	result := db.Where("username = ?", username).Find(&user)
	//	w.Header().Set("Content-Type", "application/json") //设置响应头
	if result.Error != nil {
		st := "找不到用户信息"
		response := DouyinUserResponse{
			StatusCode: 1,
			StatusMsg:  &st,
			//User:       nil,
		}
		err = json.NewEncoder(w).Encode(response)

		return
	}

	usermsg := User{
		Avatar:          user.Avatar,
		BackgroundImage: user.BackgroundImage,
		FavoriteCount:   user.FavoriteCount,
		FollowCount:     user.FollowCount,
		FollowerCount:   user.FollowerCount,
		ID:              user.ID,
		IsFollow:        user.IsFollow,
		Name:            user.Username,
		Signature:       user.Signature,
		TotalFavorited:  user.TotalFavorited,
		WorkCount:       user.WorkCount,
	}

	log.Print(user.ID, user.Username, user.IsFollow)
	response := DouyinUserResponse{
		StatusCode: 0,
		StatusMsg:  nil,
		User:       usermsg,
	}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(&response)
	/*data, _ := json.Marshal(response)
	// 发送JSON格式的数据给前端
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)*/
	if err != nil {
		log.Println("user message response error :")
		log.Println(err)
		http.Error(w, "response error", http.StatusInternalServerError)
		return
	}
	log.Println("usermessage 运行完毕")
	log.Println("statusscode: ", response.StatusCode)
	log.Println("statusMsg ", response.StatusMsg)
}
