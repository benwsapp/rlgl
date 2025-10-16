FROM --platform=$BUILDPLATFORM golang:1.25.3-alpine AS builder

ARG TARGETOS
ARG TARGETARCH

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build \
    -ldflags="-w -s -extldflags '-static'" \
    -a \
    -o rlgl . && \
    echo "nonroot:x:65532:65532:nonroot:/:/sbin/nologin" > /tmp/passwd

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /tmp/passwd /etc/passwd
COPY --from=builder /build/rlgl /rlgl

USER 65532:65532

EXPOSE 8080

ENTRYPOINT ["/rlgl"]
CMD ["serve"]
