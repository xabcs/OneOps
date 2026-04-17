import requests
import json

# 测试登录API
def test_login():
    url = "http://localhost:8085/api/login"
    headers = {"Content-Type": "application/json"}
    data = {
        "username": "admin",
        "password": "123456"
    }
    
    try:
        response = requests.post(url, headers=headers, json=data)
        print(f"状态码: {response.status_code}")
        print(f"响应内容: {response.text}")
        
        if response.status_code == 200:
            result = response.json()
            if result.get("success"):
                print("✅ 登录成功！")
                print(f"Token: {result.get('data', {}).get('token')}")
            else:
                print("❌ 登录失败:", result.get("message"))
        else:
            print("❌ HTTP错误:", response.status_code)
            
    except requests.exceptions.ConnectionError:
        print("❌ 无法连接到服务器，请确保后端服务正在运行")
    except Exception as e:
        print(f"❌ 请求异常: {e}")

if __name__ == "__main__":
    test_login()