#!/bin/bash

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 项目信息
IMAGE_NAME="diagnostic-agent"
IMAGE_TAG="latest"
REGISTRY="" # 如果有私有镜像仓库，在这里填写

# 项目目录
PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$PROJECT_DIR" || exit 1

# 设置Go代理环境变量
export GOPROXY=https://goproxy.cn,direct
export GO111MODULE=on

echo -e "${GREEN}=== Diagnostic Agent 构建脚本 ===${NC}"

# 1. 检查Docker是否安装
if ! command -v docker &> /dev/null; then
    echo -e "${RED}错误: Docker未安装，请先安装Docker${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Docker检查通过${NC}"

# 2. 检查go.mod是否存在
if [ ! -f "go.mod" ]; then
    echo -e "${RED}错误: go.mod文件不存在${NC}"
    exit 1
fi

# 3. 如果go.sum不存在，尝试生成（可选）
if [ ! -f "go.sum" ]; then
    echo -e "${YELLOW}go.sum不存在，将在Docker构建时自动生成${NC}"
fi

# 4. 构建Docker镜像
echo -e "${GREEN}开始构建Docker镜像...${NC}"
docker build -t ${IMAGE_NAME}:${IMAGE_TAG} .

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Docker镜像构建成功: ${IMAGE_NAME}:${IMAGE_TAG}${NC}"
else
    echo -e "${RED}✗ Docker镜像构建失败${NC}"
    exit 1
fi

# 5. 如果有镜像仓库，推送到仓库
if [ ! -z "$REGISTRY" ]; then
    FULL_IMAGE="${REGISTRY}/${IMAGE_NAME}:${IMAGE_TAG}"
    echo -e "${GREEN}标记镜像: ${FULL_IMAGE}${NC}"
    docker tag ${IMAGE_NAME}:${IMAGE_TAG} ${FULL_IMAGE}

    echo -e "${GREEN}推送镜像到仓库...${NC}"
    docker push ${FULL_IMAGE}

    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ 镜像推送成功: ${FULL_IMAGE}${NC}"
    else
        echo -e "${RED}✗ 镜像推送失败${NC}"
        exit 1
    fi
else
    echo -e "${YELLOW}未配置镜像仓库，跳过推送步骤${NC}"
    echo -e "${YELLOW}本地镜像: ${IMAGE_NAME}:${IMAGE_TAG}${NC}"
fi

# 6. 保存镜像（可选）
echo -e "${GREEN}是否保存镜像为tar文件? (y/n)${NC}"
read -r save_response
if [[ "$save_response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
    TAR_FILE="${IMAGE_NAME}-${IMAGE_TAG}.tar"
    echo -e "${GREEN}保存镜像到 ${TAR_FILE}...${NC}"
    docker save ${IMAGE_NAME}:${IMAGE_TAG} -o ${TAR_FILE}
    echo -e "${GREEN}✓ 镜像已保存: ${TAR_FILE}${NC}"
fi

echo -e "${GREEN}=== 构建完成 ===${NC}"
echo -e "${GREEN}运行以下命令部署到K8s:${NC}"
echo -e "  kubectl apply -f k8s-deployment.yaml"