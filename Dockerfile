FROM golang:alpine AS builder
WORKDIR /app
COPY . .
RUN go get && go build .

FROM alpine
COPY --from=builder /app/proxmox-zfs-exporter /proxmox-zfs-exporter
RUN adduser -D -H promexporter
USER promexporter
ENTRYPOINT /proxmox-zfs-exporter
EXPOSE 9000