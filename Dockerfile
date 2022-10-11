# Build Cloud Platform tools (CLI)
FROM golang:1.19.2-alpine AS cli_builder

ENV \
    CGO_ENABLED=0 \
    GOOS=linux \
    KUBECTL_VERSION=1.21.5 \
    CLOUD_PLATFORM_CLI_VERSION=DOCKER \
    TERRAFORM_VERSION=0.14.8

WORKDIR /build

RUN \
  apk add \
    --no-cache \
    --no-progress \
    --update \
    curl \
    unzip

# Build cli
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
# To get the latest build tag into the image please build using docker build --build-arg CLOUD_PLATFORM_CLI_VERSION=<latest-tag> .
RUN go build -ldflags "-X github.com/ministryofjustice/cloud-platform-cli/pkg/commands.Version=${CLOUD_PLATFORM_CLI_VERSION}" -o cloud-platform .

# Install kubectl
RUN curl -sLo ./kubectl https://storage.googleapis.com/kubernetes-release/release/v${KUBECTL_VERSION}/bin/linux/amd64/kubectl

# Install terraform
RUN curl -sLo terraform.zip https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip && unzip terraform.zip

RUN chmod +x kubectl terraform

# ---

FROM alpine:3.16.2

ENV AWSCLI_VERSION=2.7.6
ENV GLIBC_VER=2.31-r0

RUN apk add --update --no-cache \
  groff \
  bash \
  ca-certificates \
  coreutils \
  findutils \
  git-crypt \
  git \
  gnupg \
  grep \
  openssl


# AWS cli installation taken from https://github.com/aws/aws-cli/issues/4685#issuecomment-941927371
RUN apk add --no-cache --virtual .dependencies binutils curl \
    && curl -sL https://alpine-pkgs.sgerrand.com/sgerrand.rsa.pub -o /etc/apk/keys/sgerrand.rsa.pub \
    && curl -sLO https://github.com/sgerrand/alpine-pkg-glibc/releases/download/${GLIBC_VER}/glibc-${GLIBC_VER}.apk \
    && curl -sLO https://github.com/sgerrand/alpine-pkg-glibc/releases/download/${GLIBC_VER}/glibc-bin-${GLIBC_VER}.apk \
    && curl -sLO https://github.com/sgerrand/alpine-pkg-glibc/releases/download/${GLIBC_VER}/glibc-i18n-${GLIBC_VER}.apk \
    && apk add --no-cache --virtual .glibc \
        glibc-${GLIBC_VER}.apk \
        glibc-bin-${GLIBC_VER}.apk \
        glibc-i18n-${GLIBC_VER}.apk \
    && /usr/glibc-compat/bin/localedef -i en_US -f UTF-8 en_US.UTF-8 \
    && curl -sL https://awscli.amazonaws.com/awscli-exe-linux-x86_64-${AWSCLI_VERSION}.zip -o awscliv2.zip \
    && unzip awscliv2.zip \
    && aws/install \
    && rm -rf \
        awscliv2.zip \
        aws \
        /usr/local/aws-cli/v2/*/dist/aws_completer \
        /usr/local/aws-cli/v2/*/dist/awscli/data/ac.index \
        /usr/local/aws-cli/v2/*/dist/awscli/examples \
        glibc-*.apk \
    && apk del --purge .dependencies

COPY --from=cli_builder /build/cloud-platform /usr/local/bin/cloud-platform
COPY --from=cli_builder /build/kubectl /usr/local/bin/kubectl
COPY --from=cli_builder /build/terraform /usr/local/bin/terraform

CMD /bin/sh
