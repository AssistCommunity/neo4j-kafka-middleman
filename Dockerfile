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

RUN mkdir /log

FROM scratch

WORKDIR /bin

COPY --from=builder /build/main .

COPY --from=builder /log /log

WORKDIR /conf
ADD config.yaml .

ENTRYPOINT ["/bin/main"]
