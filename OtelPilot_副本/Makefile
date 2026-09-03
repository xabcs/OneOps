IMAGE ?= registry.example.com/observability/otelpilot
TAG ?= v0.1.0

.PHONY: build test image push deploy undeploy example

build:
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o bin/otelpilot .

test:
	go test ./...

image:
	docker build -t $(IMAGE):$(TAG) .

push:
	docker push $(IMAGE):$(TAG)

# 按序部署：命名空间/RBAC/配置 -> 证书初始化 Job -> 组件 -> Webhook 配置
deploy:
	kubectl apply -f deploy/00-namespace.yaml
	kubectl apply -f deploy/01-rbac.yaml
	kubectl apply -f deploy/02-configmap.yaml
	kubectl apply -f deploy/03-cert-bootstrap-job.yaml
	kubectl -n otel-pilot wait --for=condition=complete job/otel-pilot-cert-bootstrap --timeout=120s
	kubectl apply -f deploy/04-deployment.yaml
	kubectl apply -f deploy/05-service.yaml
	kubectl apply -f deploy/06-mutatingwebhook.yaml
	kubectl -n otel-pilot rollout status deployment/otel-pilot --timeout=180s

undeploy:
	kubectl delete -f deploy/06-mutatingwebhook.yaml --ignore-not-found
	kubectl delete -f deploy/05-service.yaml --ignore-not-found
	kubectl delete -f deploy/04-deployment.yaml --ignore-not-found
	kubectl delete -f deploy/03-cert-bootstrap-job.yaml --ignore-not-found
	kubectl delete -f deploy/02-configmap.yaml --ignore-not-found
	kubectl delete -f deploy/01-rbac.yaml --ignore-not-found
	kubectl delete -f deploy/00-namespace.yaml --ignore-not-found

example:
	kubectl apply -f deploy/example-app.yaml
