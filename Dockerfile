# Build Cloud Platform tools (CLI)
FROM golang:1.13.0-stretch AS cloud_platform_cli_builder

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
  KUBECTL_VERSION=1.18.16

RUN \
  apk add \
    --no-cache \
    --no-progress \
    --update \
    --virtual \
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
COPY --from=hashicorp/terraform:0.14.7 /bin/terraform /usr/local/bin/terraform

# Install aws-iam-authenticator (required for EKS)
RUN curl -sLo /usr/local/bin/aws-iam-authenticator https://amazon-eks.s3-us-west-2.amazonaws.com/1.14.6/2019-08-22/bin/linux/amd64/aws-iam-authenticator

# Ensure everything is executable
RUN chmod +x /usr/local/bin/*

CMD /bin/bash
