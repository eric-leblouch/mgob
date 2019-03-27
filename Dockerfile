FROM golang:1.11

ARG APP_VERSION=unkown

ADD . /go/src/github.com/stefanprodan/mgob

WORKDIR /go/src/github.com/stefanprodan/mgob

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-X main.version=$APP_VERSION" \
    -a -installsuffix cgo -o mgob github.com/stefanprodan/mgob

FROM alpine:edge

ARG BUILD_DATE
ARG VCS_REF
ARG VERSION

ENV MONGODB_TOOLS_VERSION 4.0.6-r0
ENV GOOGLE_CLOUD_SDK_VERSION 235.0.0
ENV AZURE_CLI_VERSION 2.0.58
ENV PATH /root/google-cloud-sdk/bin:$PATH

LABEL org.label-schema.build-date=$BUILD_DATE \
      org.label-schema.name="mgob" \
      org.label-schema.description="MongoDB backup automation tool" \
      org.label-schema.url="https://github.com/eric-leblouch/mgob" \
      org.label-schema.vcs-ref=$VCS_REF \
      org.label-schema.vcs-url="https://github.com/eric-leblouch/mgob" \
      org.label-schema.vendor="mydatamodels.com" \
      org.label-schema.version=$VERSION \
      org.label-schema.schema-version="1.0"

RUN apk add --no-cache ca-certificates mongodb-tools=${MONGODB_TOOLS_VERSION}
ADD https://dl.minio.io/client/mc/release/linux-amd64/mc /usr/bin
RUN chmod u+x /usr/bin/mc

WORKDIR /root/

#install gcloud
# https://github.com/GoogleCloudPlatform/cloud-sdk-docker/blob/69b7b0031d877600a9146c1111e43bc66b536de7/alpine/Dockerfile
RUN apk --no-cache add \
        curl \
        python3 \
        py3-crcmod \
        bash \
        libc6-compat \
        openssh-client \
        git \
    && curl -O https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-sdk-${GOOGLE_CLOUD_SDK_VERSION}-linux-x86_64.tar.gz && \
    tar xzf google-cloud-sdk-${GOOGLE_CLOUD_SDK_VERSION}-linux-x86_64.tar.gz && \
    rm google-cloud-sdk-${GOOGLE_CLOUD_SDK_VERSION}-linux-x86_64.tar.gz && \
    ln -s /lib /lib64 && \
    gcloud config set core/disable_usage_reporting true && \
    gcloud config set component_manager/disable_update_check true && \
    gcloud config set metrics/environment github_docker_image && \
    gcloud --version

# install azure-cli
RUN apk add py3-pip && \
  apk add --virtual=build gcc libffi-dev musl-dev openssl-dev python3-dev make && \
  pip3 install --upgrade pip && \
  pip3 install cffi && \
  pip3 install azure-cli==${AZURE_CLI_VERSION} && \
  sed -i 's/python -m/python3 -m/' /usr/bin/az && \
  apk del --purge build

COPY --from=0 /go/src/github.com/stefanprodan/mgob/mgob .

VOLUME ["/config", "/storage", "/tmp", "/data", "/secrets"]

ENTRYPOINT [ "./mgob" ]
