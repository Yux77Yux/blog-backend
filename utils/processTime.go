package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const ipinfoURL = "https://ipinfo.io/"

type IPInfo struct {
	IP       string `json:"ip"`
	City     string `json:"city"`
	Region   string `json:"region"`
	Country  string `json:"country"`
	Location string `json:"loc"`
	Timezone string `json:"timezone"`
}

func TransfromTime(r *http.Request, timezone string) string {
	clientLocation, err := time.LoadLocation(timezone)
	if err != nil {
		return fmt.Sprintf("Invalid time zone: %s", timezone)
	}

	now := time.Now()

	clientTime := now.In(clientLocation)

	mysqlDatetime := clientTime.Format("2006-01-02 15:04:05")

	return mysqlDatetime
}

func GetIPInfo(ip string) (*IPInfo, error) {
	resp, err := http.Get(fmt.Sprintf("%s%s?token=c60e5a7e2bb42d", ipinfoURL, ip))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var info IPInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, err
	}

	return &info, nil
}

func GetClientIP(r *http.Request) string {
	// 尝试从 X-Forwarded-For 中获取
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		// 尝试从 X-Real-IP 中获取
		ip = r.Header.Get("X-Real-IP")
	}
	if ip == "" {
		// 默认从 RemoteAddr 获取
		ip = r.RemoteAddr
	}
	return ip
}
