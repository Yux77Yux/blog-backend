package aliyun

import (
	"fmt"
	"io"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/yux77yux/blog-backend/config"
)

func CreateBucket() {
	bucketName := "blog"

	client, err := oss.New(config.District, config.AccessKey_ID, config.AccessKey_Secret)
	if err != nil {
		fmt.Println("Error", err)
	}

	err = client.CreateBucket(bucketName)
	if err != nil {
		fmt.Println("Error", err)
	}
}

func UploadFile(file io.Reader, objectName string) (string, error) {
	bucketName := "20240802"

	// 创建OSSClient实例。
	client, err := oss.New(config.District, config.AccessKey_ID, config.AccessKey_Secret)
	if err != nil {
		return "", err
	}

	// 获取存储空间。
	bucket, err := client.Bucket(bucketName)
	if err != nil {
		return "", err
	}
	// 上传文件。
	err = bucket.PutObject(objectName, file)
	if err != nil {
		return "", err
	}

	presignedURL, err := bucket.SignURL(objectName, oss.HTTPGet, 31536000) // 31536000 秒有效期
	if err != nil {
		return "", err
	}

	return presignedURL, nil
}
