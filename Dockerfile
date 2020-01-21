# Build Cloud Platform tools (CLI)
FROM golang:1.13.0-stretch AS cp_tools_builder

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux

WORKDIR /build
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN go build -o cp-tools ./cmd/cp-tools/main.go 
RUN pwd && ls

FROM alpine:3.11.0

ENV \
  TERRAFORM_VERSION=0.12.17 \
  KUBECTL_VERSION=1.13.11 \
  HELM_VERSION=2.14.3 \
  TERRAFORM_PINGDOM_VERSION=1.1.1 \
  TERRAFORM_AUTH0_VERSION=0.2.1

RUN \
  apk add \
    --no-cache \
    --no-progress \
    --update \
    --virtual \
    build-deps \
    build-base \
    bash \
    ca-certificates \
    coreutils \
    curl \
    findutils \
    git \
    gnupg \
    grep \
    jq \
    openssl \
    openssl-dev \
    python3

RUN curl https://bootstrap.pypa.io/get-pip.py -o get-pip.py && \
    python3 get-pip.py && \
    pip3 install pygithub boto3 && \
    pip3 install awscli

COPY --from=cp_tools_builder /build/cp-tools /usr/local/bin/cp-tools

# Install git-crypt
RUN git clone https://github.com/AGWA/git-crypt.git \
  && cd git-crypt && make && make install && cd - && rm -rf git-crypt

# Install kubectl
RUN curl -sLo /usr/local/bin/kubectl https://storage.googleapis.com/kubernetes-release/release/v${KUBECTL_VERSION}/bin/linux/amd64/kubectl

# Install helm
RUN curl -sL https://storage.googleapis.com/kubernetes-helm/helm-v${HELM_VERSION}-linux-amd64.tar.gz | tar -xzC /usr/local/bin --strip-components 1 linux-amd64/helm

# Install terraform
RUN curl -sL https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip | unzip -d /usr/local/bin -

# Install aws-iam-authenticator (required for EKS)
RUN curl -sLo /usr/local/bin/aws-iam-authenticator https://amazon-eks.s3-us-west-2.amazonaws.com/1.14.6/2019-08-22/bin/linux/amd64/aws-iam-authenticator

# Ensure everything is executable
RUN chmod +x /usr/local/bin/*

# Create terraform plugins directory
RUN mkdir -p ~/.terraform.d/plugins

# Install terraform auth0 provider
RUN curl -sL https://github.com/yieldr/terraform-provider-auth0/releases/download/v${TERRAFORM_AUTH0_VERSION}/terraform-provider-auth0_v${TERRAFORM_AUTH0_VERSION}_linux_amd64.tar.gz | tar xzv  \
  && mv terraform-provider-auth0_v${TERRAFORM_AUTH0_VERSION} ~/.terraform.d/plugins/

# Install Pingdom provider
RUN wget https://github.com/russellcardullo/terraform-provider-pingdom/releases/download/v${TERRAFORM_PINGDOM_VERSION}/terraform-provider-pingdom_v${TERRAFORM_PINGDOM_VERSION}_linux_amd64_static \
  && chmod +x terraform-provider-pingdom_v${TERRAFORM_PINGDOM_VERSION}_linux_amd64_static \
  && mv terraform-provider-pingdom_v${TERRAFORM_PINGDOM_VERSION}_linux_amd64_static ~/.terraform.d/plugins/terraform-provider-pingdom_v${TERRAFORM_PINGDOM_VERSION}

CMD /bin/bash

