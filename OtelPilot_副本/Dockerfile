# 多阶段构建：静态 Go 二进制 + distroless 非 Root 运行
FROM golang:1.24 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/otelpilot .

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/otelpilot /otelpilot
USER 65532:65532
ENTRYPOINT ["/otelpilot"]
