package relation

import (
	"demo-tiktok/userlogin"
	"encoding/json"
	"github.com/dgrijalva/jwt-go"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"log"
	"net/http"
)

// TableName 重命名表名
func (ActionTry) TableName() string {
	return "relation"
}

// FindFollowT 查询的对象
type FindFollowT struct {
	Token  string `json:"token"`   // 用户鉴权token
	UserID string `json:"user_id"` // 用户id
}

// FollowList 关注列表
type FollowList struct {
	StatusCode string               `json:"status_code"` // 状态码，0-成功，其他值-失败
	StatusMsg  string               `json:"status_msg"`  // 返回状态描述
	UserList   []userlogin.Usertype `json:"user_list"`   // 用户信息列表
}

// FollowerList 粉丝列表
type FollowerList struct {
	StatusCode string               `json:"status_code"` // 状态码，0-成功，其他值-失败
	StatusMsg  string               `json:"status_msg"`  // 返回状态描述
	UserList   []userlogin.Usertype `json:"user_list"`   // 用户列表
}

//var dbuser *gorm.DB
//var dbre *gorm.DB

func linkDB() error {

	return nil
}

//var db = InitDB()

func Getfollow(username string) []userlogin.Usertype {
	dsn := "root:123456@tcp(127.0.0.1:3306)/relation?charset=utf8mb4&parseTime=True&loc=Local"
	var err error
	dbre, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			//解决查表的时候自动添加复数的问题，例如user变成了users
			SingularTable: true,
		},
	})

	err = dbre.AutoMigrate(&Lover{})
	if err != nil {
		//panic("Action->InitDB 报错：" + err.Error())
		log.Println("error in linkDB:", err)

	}

	dsnUser := "root:123456@tcp(127.0.0.1:3306)/userlist?charset=utf8mb4&parseTime=True&loc=Local"
	dbuser, _ := gorm.Open(mysql.Open(dsnUser), &gorm.Config{
		//Logger: logger.Default.LogMode(logger.Info),记录日志
		Logger: logger.Default.LogMode(logger.Silent), //禁用日志记录
	})
	log.Println("linkDB here")
	err = dbuser.AutoMigrate(&userlogin.Usertype{})
	if err != nil {
		log.Println("dbuser error")

	}

	var lovername []Lover
	log.Println("找", username, "关注的人")
	log.Println(username)
	result := dbre.Where("username = ?", username).Find(&lovername)
	if result.Error != nil {
		log.Println("dbre error:", result.Error)
	}
	log.Println(len(lovername))

	var res []userlogin.Usertype
	for _, i := range lovername {
		var user userlogin.Usertype
		dbuser.Where("username = ?", i.Lovername).First(&user)
		log.Println("你关注了：", i)

		res = append(res, user)
	}
	return res
}

// FollowListGet 关注列表
func FollowListGet(w http.ResponseWriter, r *http.Request) {
	log.Println("开始执行关注列表")
	erro := linkDB()
	log.Println("关注列表执行到这里了")
	if erro != nil {
		log.Println("db error in FollowListGet")
		return
	}
	queryValues := r.URL.Query()
	token := queryValues.Get("token")
	//userID := queryValues.Get("user_id")
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
	var users []userlogin.Usertype

	users = Getfollow(username)
	var fl FollowList

	fl.UserList = users
	fl.StatusMsg = "返回关注列表成功"
	fl.StatusCode = "0"
	log.Println("准备返回")
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(fl)
	if err != nil {
		http.Error(w, "响应编码错误5", http.StatusInternalServerError)
	}

}

func Getfollower(username string) []userlogin.Usertype {
	dsn := "root:123456@tcp(127.0.0.1:3306)/relation?charset=utf8mb4&parseTime=True&loc=Local"
	var err error
	dbre, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			//解决查表的时候自动添加复数的问题，例如user变成了users
			SingularTable: true,
		},
	})

	err = dbre.AutoMigrate(&Lover{})
	if err != nil {
		//panic("Action->InitDB 报错：" + err.Error())
		log.Println("error in linkDB:", err)

	}

	dsnUser := "root:123456@tcp(127.0.0.1:3306)/userlist?charset=utf8mb4&parseTime=True&loc=Local"
	dbuser, _ := gorm.Open(mysql.Open(dsnUser), &gorm.Config{
		//Logger: logger.Default.LogMode(logger.Info),记录日志
		Logger: logger.Default.LogMode(logger.Silent), //禁用日志记录
	})
	log.Println("linkDB here")
	err = dbuser.AutoMigrate(&userlogin.Usertype{})
	if err != nil {
		log.Println("dbuser error")

	}

	var lovername []Lover
	dbre.Where("lovername = ?", username).Find(&lovername)
	log.Println("2")
	var res []userlogin.Usertype
	for _, i := range lovername {
		var user userlogin.Usertype
		dbuser.Where("username = ?", i.Username).First(&user)
		res = append(res, user)
	}
	return res
}

// FollowerListGet 粉丝列表
func FollowerListGet(w http.ResponseWriter, r *http.Request) {
	// 连接数据库

	//erro := linkDB()
	//if erro != nil {
	//	log.Println("db error in FollowerListGet")
	//	return
	//}

	queryValues := r.URL.Query()
	token := queryValues.Get("token")
	//userID := queryValues.Get("user_id")
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

	// 查询对应的 users
	var users []userlogin.Usertype
	users = Getfollower(username)
	var ferl FollowerList

	ferl.UserList = users
	ferl.StatusMsg = "返回粉丝列表成功"
	ferl.StatusCode = "0"

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(ferl)
	if err != nil {
		http.Error(w, "响应编码错误5", http.StatusInternalServerError)
	}

}

//http.HandleFunc("/douyin/relation/action", relation.Action)
//http.HandleFunc("/douyin/relation/follow/list", relation.FollowListGet)
//http.HandleFunc("/douyin/relation/follower/list", relation.FollowerListGet)
