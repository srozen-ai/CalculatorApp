# Build stage: compile a static binary so the final image needs no libc.
FROM golang:1.26-alpine AS build

WORKDIR /src

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -o /out/api ./cmd/api

# Runtime stage: distroless static has no shell and no package manager.
FROM gcr.io/distroless/static-debian12

COPY --from=build /out/api /api

EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/api"]
