VERSION --use-cache-command 0.7
FROM golang:1.19.6
WORKDIR /workdir

all:
    WAIT
        BUILD +lint
        BUILD +test
    END
    WAIT
        BUILD +gosec
    END

deps:
    COPY go.mod go.sum ./
    RUN go mod download

deps-test:
    FROM +deps
    COPY Makefile ./
    RUN make mockgen setup-envtest

lint:
    FROM earthly/dind:alpine
    COPY . ./workdir
    WITH DOCKER --pull golangci/golangci-lint:v1.51.0
        RUN docker run -w /workdir -v /workdir:/workdir golangci/golangci-lint:v1.51.0 golangci-lint run --timeout 240s
    END

gosec:
    FROM earthly/dind:alpine
    COPY . ./workdir
    WITH DOCKER --pull securego/gosec:2.15.0
        RUN docker run -w /workdir -v /workdir:/workdir securego/gosec:2.15.0 -exclude-dir=example -exclude-generated ./...
    END

test:
    FROM +deps-test
    CACHE $HOME/.cache/go-build
    COPY . ./
    RUN make _test
