#!/bin/bash

# 测试登录API
curl -X POST http://localhost:8084/api/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"123456"}'