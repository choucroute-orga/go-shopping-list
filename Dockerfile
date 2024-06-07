FROM golang:1.22-alpine AS builder

WORKDIR /usr/src/app/

RUN apk update

RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=bind,source=./go-shopping-list/go.sum,target=go.sum \
    --mount=type=bind,source=./go-shopping-list/go.mod,target=go.mod \
    go mod download

RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=bind,rw,source=./go-shopping-list,target=. \
    go build -ldflags "-s -w" -o /go/bin/shopping-list/ ./

FROM alpine

WORKDIR /usr/src/app/

COPY --from=builder /go/bin/shopping-list/ ./

EXPOSE ${SHOPPING_LIST_PORT}
ENTRYPOINT [ "./shopping-list" ]