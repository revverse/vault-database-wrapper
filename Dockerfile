FROM golang:1.24.1 as builder

WORKDIR /opt
COPY ./ /opt/

RUN go mod download
RUN CGO_ENABLED=0 go build -ldflags='-extldflags=-static' -o vlt_plugins/vault-database-wrapper  .

RUN ls /opt/vlt_plugins/
RUN ls .


FROM hashicorp/vault:1.19
COPY --from=builder /opt/vlt_plugins/vault-database-wrapper /vault/plugins/vault-database-wrapper

RUN chmod +x /vault/plugins/vault-database-wrapper
RUN ls /vault/plugins/vault-database-wrapper

CMD [ "server", "-dev", "-dev-plugin-dir=/vault/plugins" ]

# docker run --rm --name vault -e PLUGIN_DIR=/vault/plugins  -v $(pwd)/vlt_plugins:/vault/plugins:Z  -p 8200:8200 --cap-add IPC_LOCK hashicorp/vault server -dev -dev-plugin-dir=/vault/plugins