VERSION 0.6
FROM golang:1.19.6
WORKDIR /workdir

deps:
    COPY Makefile ./
    RUN make mockgen setup-envtest
    COPY go.mod go.sum ./
    RUN go mod tidy
    RUN go mod download
    SAVE ARTIFACT go.mod AS LOCAL go.mod
    SAVE ARTIFACT go.sum AS LOCAL go.sum

lint:
    FROM earthly/dind:alpine
    WORKDIR /workdir
    COPY . ./
    WITH DOCKER --pull golangci/golangci-lint:v1.49.0
        RUN docker run -w $PWD -v $PWD:$PWD golangci/golangci-lint:v1.49.0 golangci-lint run --timeout 240s
    END

gosec:
    FROM earthly/dind:alpine
    WORKDIR /workdir
    COPY . ./
    WITH DOCKER --pull securego/gosec:2.15.0
        RUN docker run -w $PWD -v $PWD:$PWD securego/gosec:2.15.0 -exclude-dir=example -exclude-generated ./...
    END

test:
    FROM +deps
    COPY . ./
    RUN make _test
