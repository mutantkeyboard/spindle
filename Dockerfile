FROM golang:1.25-alpine

RUN apk add --no-cache gcc musl-dev git curl

# Install golangci-lint
RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b /usr/local/bin

WORKDIR /workspace

COPY go.mod go.sum ./
RUN go mod download

COPY . .
