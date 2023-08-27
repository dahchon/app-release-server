FROM golang:1.20 as build

WORKDIR /workspace

# add go modules lockfiles
COPY go.mod go.sum ./
RUN go mod download

# prefetch the binaries, so that they will be cached and not downloaded on each change
RUN go run github.com/steebchen/prisma-client-go prefetch

COPY . ./

RUN go run github.com/steebchen/prisma-client-go generate

ENV CGO_ENABLED=0

RUN go build -o /app-release-server .

CMD ["/app"]

FROM debian:bullseye-slim as production

WORKDIR /app

COPY --from=build /app-release-server /app

COPY schema.prisma .
COPY migrations .

COPY bin/entrypoint.sh .

# CMD ["/app/app-release-server"]
ENTRYPOINT ["/app/entrypoint.sh"]
