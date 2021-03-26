FROM golang:alpine AS builder

# Set necessary environmet variables needed for our image
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /build

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN go build -o main .

FROM scratch

WORKDIR /bin

COPY --from=builder /build/main .

WORKDIR /conf
ADD config.yaml .

WORKDIR /

ENTRYPOINT ["/bin/main"]
