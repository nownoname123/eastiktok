// Package friendlist 好友列表接口
package friendlist

import (
	"demo-tiktok/relation"
	"demo-tiktok/userlogin"
	"encoding/json"
	"github.com/dgrijalva/jwt-go"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"log"
	"net/http"
)

type Response struct {
	StatusCode string `json:"status_code"` // 状态码，0-成功，其他值-失败
	StatusMsg  string `json:"status_msg"`  // 返回状态描述
	UserList   []User `json:"user_list"`   // 用户列表
}

type Req struct {
	Token  string `json:"token"`   // 用户鉴权token
	UserID string `json:"user_id"` // 用户id
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

func FindFriends(username string) ([]User, error) {
	var res []User
	dsn := "root:123456@tcp(127.0.0.1:3306)/userlist?charset=utf8mb4&parseTime=True&loc=Local" //连接数据库
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return res, err
	}
	db.AutoMigrate(&userlogin.Usertype{})

	dsn = "root:123456@tcp(127.0.0.1:3306)/relation?charset=utf8mb4&parseTime=True&loc=Local"

	dbre, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			//解决查表的时候自动添加复数的问题，例如user变成了users
			SingularTable: true,
		},
	})

	err = dbre.AutoMigrate(&relation.Lover{})
	var love []relation.Lover
	result := dbre.Where("username = ?", username).Find(&love)
	log.Println("找", username, "关注的人")
	if result.Error != nil {
		log.Println("dbre error in friend : ", result.Error)
		return res, err
	}

	for _, i := range love {

		var ff []relation.Lover
		log.Println(username, "关注了", i.Lovername)
		result := dbre.Where("username = ?", i.Lovername).Find(&ff)
		if result.Error != nil {
			log.Println("dbre error in friend 2 : ", result.Error)

		}
		var f bool
		f = false // 遍历user关注的所有对象，看看是不是也关注了自己，是的就是互相关注了，作为好友
		for _, j := range ff {
			log.Println(i.Lovername, "关注了", j.Lovername)
			if j.Lovername == username {
				f = true
				break
			}

		}
		if f {
			var u userlogin.Usertype

			log.Println(username, "和", i.Lovername, "互相关注了")
			db.Where("username = ?", i.Lovername).First(&u)
			var v User
			v = User{
				Avatar:          u.Avatar,
				BackgroundImage: u.BackgroundImage,
				FavoriteCount:   u.FavoriteCount,
				FollowCount:     u.FollowCount,
				FollowerCount:   u.FollowerCount,
				ID:              u.ID,
				IsFollow:        u.IsFollow,
				Name:            u.Username,
				Signature:       u.Signature,
				TotalFavorited:  u.TotalFavorited,
				WorkCount:       u.WorkCount,
			}
			res = append(res, v)
		}
	}
	return res, nil
}
func Getfriendlist(w http.ResponseWriter, r *http.Request) {

	queryValues := r.URL.Query()
	token := queryValues.Get("token")
	//userId := queryValues.Get("user_id")
	w.Header().Set("Content-Type", "application/json")
	token1, _ := userlogin.CheckToken(token)
	var username string
	if token1.Valid {
		claims := token1.Claims.(jwt.MapClaims)
		username = claims["username"].(string)
		//expiration := claims["exp"].(float64)

	} else {
		return
	}
	var response Response
	var err error
	response.UserList, err = FindFriends(username)
	if err != nil {
		response.StatusCode = "1"
		response.StatusMsg = "好友列表加载失败请重试"
		err = json.NewEncoder(w).Encode(response)
		return
	}
	response.StatusCode = "0"
	err = json.NewEncoder(w).Encode(response)
	log.Println("好友列表执行完成")
}
