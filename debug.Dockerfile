FROM golang:1.23 AS builder

ARG LD_FLAGS='-s -w -linkmode external -extldflags "-static"'
ARG TARGETPLATFORM

WORKDIR /app

RUN export GOOS=$(echo ${TARGETPLATFORM} | cut -d / -f1) && \
    export GOARCH=$(echo ${TARGETPLATFORM} | cut -d / -f2)

COPY go.* ./
RUN go env && go mod download

RUN go install github.com/go-delve/delve/cmd/dlv@latest

COPY . .

RUN CGO_ENABLED=1 go build -tags="sqlite_unlock_notify" -o /app/policyreporter -v

EXPOSE 8080

ENTRYPOINT ["/app/policyreporter", "run"]
