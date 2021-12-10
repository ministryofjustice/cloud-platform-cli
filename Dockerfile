# Build Cloud Platform tools (CLI)
FROM golang:1.16.0-stretch AS cloud_platform_cli_builder

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux

WORKDIR /build
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN go build -o cloud-platform ./cmd/cloud-platform/main.go 
RUN pwd && ls

FROM alpine:3.11.0

ENV \
  KUBECTL_VERSION=1.20.7 \
  TERRAFORM_VERSION=0.14.8

RUN \
  apk add \
    --no-cache \
    --no-progress \
    --update \
    bash \
    ca-certificates \
    coreutils \
    curl \
    findutils \
    git-crypt \
    git \
    gnupg \
    grep \
    openssl \
    python3

RUN curl https://bootstrap.pypa.io/get-pip.py -o get-pip.py && \
    python3 get-pip.py && \
    pip3 install pygithub boto3 && \
    pip3 install awscli

COPY --from=cloud_platform_cli_builder /build/cloud-platform /usr/local/bin/cloud-platform

# Install kubectl
RUN curl -sLo /usr/local/bin/kubectl https://storage.googleapis.com/kubernetes-release/release/v${KUBECTL_VERSION}/bin/linux/amd64/kubectl

# Install terraform
RUN curl -sLo terraform.zip https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip && unzip terraform.zip && mv terraform /usr/local/bin && rm -f terraform.zip

# Ensure everything is executable
RUN chmod +x /usr/local/bin/*

CMD /bin/bash
