# build builder
FROM golang:1.16 as builder

WORKDIR /workspace

# Copy the Go Modules manifests
COPY . .
# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GOPROXY=https://mirrors.aliyun.com/goproxy/ go build -a -o cable cmd/cable/app.go

# build server
FROM openanolis/anolisos:8.4-x86_64

RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

WORKDIR /app

COPY --from=builder /workspace/cable .
RUN chmod +x /app/cable

ENV TINI_VERSION v0.19.0
ADD https://github.com/krallin/tini/releases/download/${TINI_VERSION}/tini /tini
RUN chmod +x /tini

RUN useradd -m admin
ENV OBPROXY_VERSION 3.2.3
ENV OBPROXY_RELEASE 2
RUN rpm -ivh https://mirrors.aliyun.com/oceanbase/community/stable/el/8/x86_64/obproxy-ce-${OBPROXY_VERSION}-${OBPROXY_RELEASE}.el8.x86_64.rpm
RUN mv /home/admin/obproxy-${OBPROXY_VERSION} /home/admin/obproxy
RUN chown -R admin:admin /home/admin/obproxy

ENTRYPOINT ["/tini", "--"]
CMD ["/app/cable"]
