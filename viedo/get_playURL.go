package viedo

import (
	"context"
	"fmt"
	"github.com/tencentyun/cos-go-sdk-v5"
	"log"
	"net/http"
	"net/url"
	"time"
)

func GetPlayUrl(filename string) (string, error) {

	u, _ := url.Parse(fmt.Sprintf("https://%s.cos.%s.myqcloud.com", CosBucketName, CosRegion))
	b := &cos.BaseURL{BucketURL: u}
	c := cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  CosSecretId,
			SecretKey: CosSecretKey,
		},
	})
	objectKey := filename // 设置对象键，包括完整的扩展名
	// 获取预签名 URL
	presignedURL, err := c.Object.GetPresignedURL(context.Background(), http.MethodGet, objectKey, CosSecretId, CosSecretKey, 3*time.Hour, nil)
	if err != nil {
		log.Println("获取预签名URL失败：", err)
		panic(err)
	}
	log.Println(filename, " 预签名URL为：：", presignedURL.String())
	return presignedURL.String(), nil
}
