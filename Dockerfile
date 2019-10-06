FROM golang:1-buster AS builder
WORKDIR /src
COPY . .
RUN go build -o /bin/bot

FROM debian:buster AS downloader
RUN apt-get -y update
RUN apt-get -y install curl
RUN curl -o keybase_amd64.deb https://prerelease.keybase.io/keybase_amd64.deb

FROM debian:buster
COPY --from=downloader keybase_amd64.deb ./
RUN apt-get -y update
RUN apt-get -y install ca-certificates build-essential
RUN apt-get -y install ./keybase_amd64.deb
RUN useradd -m keybase
USER keybase
COPY ./entrypoint.sh ./
COPY --from=builder /bin/bot ./bot
CMD ["./entrypoint.sh"]
