FROM golang:alpine AS builder
WORKDIR /app
COPY . .
RUN go get && go build .

FROM alpine
COPY --from=builder /app/proxmox-zfs-exporter /proxmox-zfs-exporter
ENTRYPOINT /proxmox-zfs-exporter