package relation

import (
	"demo-tiktok/userlogin"
	"encoding/json"
	"github.com/dgrijalva/jwt-go"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"log"
	"net/http"
	"strconv"
)

//ActionTry 关注操作

type ActionTry struct {
	Token      string `json:"token" `       // 用户鉴权token
	ActionType string `json:"action_type" ` // 1-关注，2-取消关注
	ToUserID   string `json:"to_user_id" `  // 对方用户id
}

// ActionReturn 关注操作返回状态
type ActionReturn struct {
	StatusCode int64  `json:"status_code"` // 状态码，0-成功，其他值-失败
	StatusMsg  string `json:"status_msg"`  // 返回状态描述
}
type Lover struct {
	Username  string `gorm:"column:username"`
	Lovername string `gorm:"column:lovername"`
}

var dbUser *gorm.DB

func getuser(userid int64) (userlogin.Usertype, error) {
	dsnUser := "root:123456@tcp(127.0.0.1:3306)/userlist?charset=utf8mb4&parseTime=True&loc=Local"
	var err error
	dbUser, err = gorm.Open(mysql.Open(dsnUser), &gorm.Config{})
	var user userlogin.Usertype
	if err != nil {
		return user, err
	}
	err = dbUser.AutoMigrate(&userlogin.Usertype{})

	dbUser.Where("id = ?", userid).First(&user)
	log.Println("关注了", user.Username)
	return user, nil
}
func addlover(username string, lovername string) {
	var res Lover
	dsn := "root:123456@tcp(127.0.0.1:3306)/relation?charset=utf8mb4&parseTime=True&loc=Local"

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			//解决查表的时候自动添加复数的问题，例如user变成了users
			SingularTable: true,
		},
	})
	if err != nil {
		return
	}
	err = db.AutoMigrate(&Lover{})
	res.Lovername = lovername
	res.Username = username
	db.Create(&res)
	log.Println(username, "关注了", lovername)

	err = db.AutoMigrate(&userlogin.Usertype{})
	var u, l userlogin.Usertype
	dbUser.Where("username = ?", username).First(&u)
	dbUser.Where("username = ?", lovername).First(&l)
	l.FollowerCount++
	u.FollowCount++
	dbUser.Save(&u)
	dbUser.Save(&l)
	return
}
func Dellover(username string, lovername string) error {

	dsn := "root:123456@tcp(127.0.0.1:3306)/relation?charset=utf8mb4&parseTime=True&loc=Local"

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			//解决查表的时候自动添加复数的问题，例如user变成了users
			SingularTable: true,
		},
	})
	if err != nil {
		return err
	}
	err = db.AutoMigrate(&Lover{})

	if err != nil {
		return err
	}
	var a Lover
	db.Where("username = ? AND lovername = ?", username, lovername).Delete(&a)

	err = db.AutoMigrate(&userlogin.Usertype{})
	var u, l userlogin.Usertype
	dbUser.Where("username = ?", username).First(&u)
	dbUser.Where("username = ?", lovername).First(&l)
	l.FollowerCount--
	u.FollowCount--
	dbUser.Save(&u)
	dbUser.Save(&l)
	return nil
}
func Action(w http.ResponseWriter, r *http.Request) {

	queryValues := r.URL.Query()
	token := queryValues.Get("token")
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
	//
	loverid := queryValues.Get("to_user_id")

	n, _ := strconv.Atoi(loverid)
	lid := int64(n)
	lover, _ := getuser(lid)
	ActionType := queryValues.Get("action_type")
	w.Header().Set("Content-Type", "application/json")
	if ActionType == "1" {

		addlover(username, lover.Username) //把关注关系存到数据库中

		// 完成后返回响应
		response := ActionReturn{
			StatusCode: 0,
			StatusMsg:  "关注成功",
		}

		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			http.Error(w, "响应编码错误3", http.StatusInternalServerError)
		}
	} else if ActionType == "2" {

		err := Dellover(username, lover.Username)
		if err != nil {
			log.Println("delete relation error:", err)
			response := ActionReturn{
				StatusCode: 0,
				StatusMsg:  "取消关注成功",
			}
			err = json.NewEncoder(w).Encode(response)
			return
		}
		response := ActionReturn{
			StatusCode: 0,
			StatusMsg:  "取消关注成功",
		}
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			log.Println("json error in action :", err)
		}
		return
	}
}
