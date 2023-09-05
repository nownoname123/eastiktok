package viedo

import (
	"context"
	"fmt"
	"log"

	"github.com/tencentyun/cos-go-sdk-v5"
	"net/http"
	"net/url"
	"time"
)

const (
	CosBucketName = "easy-tiktok-1319919207"

	CosRegion    = "ap-guangzhou"
	CosSecretId  = "AKIDd6LM4XllWAbonX4NdeR0LcfZIo8KK7i5"
	CosSecretKey = "DByZx98uSmtmSwFLDwSjmcHjxwhFbjfP"
)

func Upload(filepath string, filename string) (string, error) {
	// 创建 COS 客户端
	u, _ := url.Parse(fmt.Sprintf("https://%s.cos.%s.myqcloud.com", CosBucketName, CosRegion))
	b := &cos.BaseURL{BucketURL: u}
	c := cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  CosSecretId,
			SecretKey: CosSecretKey,
		},
	})

	objectKey := filename // 设置对象键，包括完整的扩展名
	_, err := c.Object.PutFromFile(context.Background(), objectKey, filepath, nil)
	if err != nil {
		log.Println("upload error :", err)
		panic(err)
	}

	// 获取对象的URL
	presignedURL, err := c.Object.GetPresignedURL(context.Background(), http.MethodGet, objectKey, CosSecretId, CosSecretKey, time.Hour, nil)
	if err != nil {
		panic(err)
	}
	log.Println(filename, " 存放在：", presignedURL.String())
	return presignedURL.String(), nil
}
