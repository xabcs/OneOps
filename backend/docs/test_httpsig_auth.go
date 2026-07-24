package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"gopkg.in/twindagger/httpsig.v1"
)

const (
	TestServerURL    = "http://192.168.4.13"
	TestAccessKeyID  = "96dd1976-7c61-483e-ab4a-d597eb0aa626"
	TestAccessKeySecret = "Cg0zPrtOfCDyIPNmI95ubbGcU1yezDC4o2qR"
)

// 使用 httpsig 库进行签名认证
func testHttpsAuth(accessKey, secret, jmsURL string) error {
	endpoint := "/api/v1/users/users/"
	fullURL := jmsURL + endpoint

	// 创建请求
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置必需的 headers
	gmtFmt := "Mon, 02 Jan 2006 15:04:05 GMT"
	req.Header.Set("Date", time.Now().Format(gmtFmt))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-JMS-ORG", "00000000-0000-0000-0000-000000000002")

	// 定义需要签名的 header 列表
	headers := []string{"(request-target)", "date"}

	// 创建签名器
	signer, err := httpsig.NewRequestSigner(accessKey, secret, "hmac-sha256")
	if err != nil {
		return fmt.Errorf("创建签名器失败: %w", err)
	}

	// 执行签名
	if err := signer.SignRequest(req, headers, nil); err != nil {
		return fmt.Errorf("签名请求失败: %w", err)
	}

	fmt.Printf("Authorization: %s\n", req.Header.Get("Authorization"))
	fmt.Printf("Date: %s\n", req.Header.Get("Date"))
	fmt.Printf("请求URL: %s\n", fullURL)

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}

	fmt.Printf("\n状态码: %d\n", resp.StatusCode)
	fmt.Printf("响应内容:\n%s\n", string(body))

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API 请求失败，状态码: %d", resp.StatusCode)
	}

	return nil
}

func main() {
	fmt.Println("====================================")
	fmt.Println("JumpServer HttpSig 认证测试")
	fmt.Println("====================================")
	fmt.Println()

	err := testHttpsAuth(TestAccessKeyID, TestAccessKeySecret, TestServerURL)
	if err != nil {
		fmt.Printf("❌ 测试失败: %v\n", err)
	} else {
		fmt.Printf("✅ 测试成功！\n")
	}
}