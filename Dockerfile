FROM golang:1.17 AS cli_builder

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY . .

RUN go build -o cloud-platform ./cmd/cloud-platform/main.go

# ---------------------------------------------------------------------------------------------------------------------
FROM debian:bullseye as tools_builder
ENV \
  KUBECTL_VERSION=1.20.7 \
  TERRAFORM_VERSION=0.14.8

WORKDIR /app

RUN apt-get update && apt-get install -y gnupg software-properties-common curl unzip

RUN curl -sLo terraform.zip https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip && unzip terraform.zip

RUN curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl" && chmod +x kubectl

# ---------------------------------------------------------------------------------------------------------------------
FROM debian:bullseye-slim

WORKDIR /app

RUN apt-get update && apt-get install -y \
    gnupg-agent \
    git-crypt \
    git \
    awscli


COPY --from=cli_builder /app/cloud-platform /usr/local/bin/cloud-platform

COPY --from=tools_builder /app/terraform /usr/local/bin/terraform

COPY --from=tools_builder /app/kubectl /usr/local/bin/kubectl

CMD [/bin/sh]
