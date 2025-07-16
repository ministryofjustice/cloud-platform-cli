# Build Cloud Platform tools (CLI)
FROM golang:1.24.3-bookworm AS cli_builder

ENV \
  CGO_ENABLED=0 \
  GOOS=linux \
  KUBECTL_VERSION=1.31.9 \
  CLOUD_PLATFORM_CLI_VERSION=DOCKER \
  TERRAFORM_VERSION=1.2.5

WORKDIR /build

RUN apt update

RUN \
  apt \
  install \
  curl \
  unzip \
  -y

# Build cli
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
# To get the latest build tag into the image please build using docker build --build-arg CLOUD_PLATFORM_CLI_VERSION=<latest-tag> .
RUN go build -ldflags "-X github.com/ministryofjustice/cloud-platform-cli/pkg/commands.Version=${CLOUD_PLATFORM_CLI_VERSION}" -o cloud-platform .

# Install kubectl
RUN curl -sLo ./kubectl https://dl.k8s.io/release/v${KUBECTL_VERSION}/bin/linux/amd64/kubectl

# Install terraform
RUN curl -sLo terraform.zip https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip && unzip terraform.zip

RUN curl -sLo opa "https://openpolicyagent.org/downloads/latest/opa_linux_amd64"

RUN chmod +x kubectl terraform opa

# ---

FROM debian:bookworm-20250520-slim

ENV AWSCLI_VERSION=2.7.6

RUN apt update

RUN apt install \
  unzip \
  groff \
  ca-certificates \
  curl \
  git-crypt \
  git \
  grep \
  openssl \
  parallel \
  python3  \
  jq \
  -y

RUN curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64-${AWSCLI_VERSION}.zip" -o "awscliv2.zip"
RUN unzip awscliv2.zip
RUN ./aws/install

COPY --from=cli_builder /build/cloud-platform /usr/local/bin/cloud-platform
COPY --from=cli_builder /build/kubectl /usr/local/bin/kubectl
COPY --from=cli_builder /build/terraform /usr/local/bin/terraform
COPY --from=cli_builder /build/opa /usr/local/bin/opa

CMD /bin/sh
