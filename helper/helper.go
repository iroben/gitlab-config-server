package helper

import (
	"crypto/tls"
	"net/http"
	"time"

	"encoding/base64"
	"sync"

	"crypto/md5"
	"crypto/rand"
	"fmt"
	"io"
	"os"
)

/**
缓存连接对象
*/
var (
	clients = map[string]*http.Client{}
	lock    = sync.Mutex{}
)

func Client(name string) *http.Client {
	lock.Lock()
	defer lock.Unlock()
	client, ok := clients[name]
	if !ok {
		client = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig:    &tls.Config{InsecureSkipVerify: true},
				DisableCompression: false,
				MaxIdleConns:       100,
				/*Proxy: func(request *http.Request) (*url.URL, error) {
					return url.Parse("http://127.0.0.1:8888")
				},*/
				ResponseHeaderTimeout: time.Second * 5,
			},
		}
		clients[name] = client
	}
	return client
}

// Md5加密
func Md5(str string) string {
	data := fmt.Sprintf("%x", md5.Sum([]byte(str)))
	return data
}

//生成Guid字串(32位)
func GetGuid() string {
	b := make([]byte, 48)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}
	return Md5(base64.URLEncoding.EncodeToString(b))
}

// 检查文件或目录是否存在
// 如果由 filename 指定的文件或目录存在则返回 true，否则返回 false
func Exist(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || os.IsExist(err)
}
