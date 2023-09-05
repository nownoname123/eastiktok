package viedo

import (
	"demo-tiktok/userlogin"
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type Req struct {
	Token string `protobuf:"bytes,1,opt,name=token" json:"token"`
	Data  []byte `protobuf:"bytes,2,opt,name=data" json:"data"`
	Title string `protobuf:"bytes,3,opt,name=title" json:"title"`
}

type Res struct {
	StatusCode int64  `json:"status_code"`
	StatusMsg  string `json:"status_msg"`
}

var (
	dbVideo *gorm.DB
	dbUser  *gorm.DB
)

func linksql() error {
	// 初始化数据库连接
	dsnVideo := "root:123456@tcp(127.0.0.1:3306)/videolist?charset=utf8mb4&parseTime=True&loc=Local"
	dsnUser := "root:123456@tcp(127.0.0.1:3306)/userlist?charset=utf8mb4&parseTime=True&loc=Local"
	var err error
	dbVideo, err = gorm.Open(mysql.Open(dsnVideo), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			//解决查表的时候自动添加复数的问题，例如user变成了users
			SingularTable: true,
		},
	})
	dbUser, _ = gorm.Open(mysql.Open(dsnUser), &gorm.Config{
		//Logger: logger.Default.LogMode(logger.Info),记录日志
		Logger: logger.Default.LogMode(logger.Silent), //禁用日志记录
	})

	if err != nil {
		return err
	}
	err = dbVideo.AutoMigrate(&Video{})
	if err != nil {
		return err
	}

	err = dbUser.AutoMigrate(&userlogin.Usertype{})
	if err != nil {
		return err
	}
	return nil
}
func getCover(title string, videoPath string) (coverPath string) {
	outputDir := "C:\\Users\\Administrator\\Desktop\\demo-tiktok\\file\\covers\\"
	outputImagePath := filepath.Join(outputDir, fmt.Sprintf("%s.jpg", title))

	// 确保输出目录存在
	err := os.MkdirAll(outputDir, os.ModePerm)
	if err != nil {
		log.Fatal("Error creating output directory:", err)
	}
	log.Println(videoPath)
	log.Println(outputImagePath)

	// 构建 FFmpeg 命令  需要事先下载ffmpeq
	//ffmpegPath := "ffmpeg.exe"
	ffmpegPath := "C:\\Users\\Administrator\\Desktop\\demo-tiktok\\ffmpeg-6.0\\ffmpeg-6.0-essentials_build\\bin\\ffmpeg.exe"

	cmd := exec.Command(ffmpegPath, "-i", videoPath, "-ss", "00:00:01", "-frames:v", "1", outputImagePath)
	log.Println("ready to ffmpeg")
	// 运行 FFmpeg 命令
	err = cmd.Run()
	if err != nil {
		log.Println("Error while extracting video frame:", err)
	}
	log.Println("ffmpeg over")
	log.Println("Video frame extracted and saved as", outputImagePath)
	u, err := Upload(outputImagePath, title+".jpg")
	return u
}
func Videopublish(w http.ResponseWriter, r *http.Request) {

	error1 := linksql() //连接数据库
	if error1 != nil {
		log.Println("linksql error :", error1)
		return
	}
	log.Println("111")
	// 读取请求 body

	// 提取表单字段值
	token := r.PostFormValue("token")
	data, header, err := r.FormFile("data")
	title := r.FormValue("title")
	log.Println("token is :", token)
	// 现在你可以使用这些值来构建你的 Req 结构体实例

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
	var user userlogin.Usertype

	log.Println("username is :", username)

	if result := dbUser.Where("username = ?", username).Find(&user); result.Error != nil {
		log.Println("db error :", result.Error)
		return
	} //查找作者信息

	log.Println("user get it")

	log.Println("作者是：", user.Username)
	var video Video

	// 获取视频数据

	// 提取文件名和后缀
	filename := filepath.Base(header.Filename)
	ext := filepath.Ext(filename)
	newFilename := title + ext

	// 构建保存路径
	videoDir := "C:\\Users\\Administrator\\Desktop\\demo-tiktok\\file\\"
	filePath := filepath.Join(videoDir, newFilename)

	outFile, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "Error creating file", http.StatusInternalServerError)
		return
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, data) //将文件写入创造好的文件夹中
	if err != nil {
		log.Println("error in write file")
		http.Error(w, "Error writing file", http.StatusInternalServerError)
		return
	}
	log.Printf("File saved at: %s\n", filePath)
	w.Header().Set("Content-Type", "application/json") //设置Http响应头，它告诉客户端接收到的数据是以 JSON 格式表示的。

	if err != nil {
		log.Println("Error creating video file:", err)
		statumsg := "视频上传错误"
		response := Res{
			StatusCode: 1,
			StatusMsg:  statumsg,
		}

		err = json.NewEncoder(w).Encode(response)
		http.Error(w, "Error creating video file", http.StatusInternalServerError)
		return
	}

	video.CreatTime = time.Now()

	videourl, err := Upload(filePath, newFilename)
	video.PlayURL = videourl
	video.Author = user.ID
	video.Title = title
	video.FavoriteCount = 0
	video.CommentCount = 0

	video.CoverURL = getCover(title, filePath)

	video.IsFavorite = false

	// 将文件路径存储到数据库
	result := dbVideo.Create(&video)
	if result.Error != nil {
		log.Println("Error while inserting into the database:", result.Error)

		http.Error(w, "Database error", http.StatusInternalServerError)
		return

	}

	user.WorkCount++
	result = dbUser.Save(&user)
	if result.Error != nil {
		log.Println("数据库中作品数修改失败:", result.Error)
		response := Res{
			StatusCode: 1,
			StatusMsg:  title + "上传失败",
		}
		err = json.NewEncoder(w).Encode(response)
		return
	}
	// 返回上传成功的消息或其他响应
	response := Res{
		StatusCode: 0,
		StatusMsg:  title + "上传成功",
	}
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "Error writing video file", http.StatusInternalServerError)
		return
	}
	log.Println("视频上传完成")
}
