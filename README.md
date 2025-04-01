# vault-database-wrapper
As simple as possible wrapper for Vault database secret engine

Can be used for services that implement access to the database via an API.

# Play on localhost

build custom image
```bash
docker build  -t vlt_plugins .
```
and run
```bash
docker run --rm --name vault -e PLUGIN_DIR=/vault/plugins  -p 8200:8200 --cap-add IPC_LOCK vlt_plugins server -dev -dev-plugin-dir=/vault/plugins
```

export token and url ( root token you can find in vault stdout )
```bash
export VAULT_ADDR=http://127.0.0.1:8200
export VAULT_TOKEN=hvs.KQd3DCCxt6Bf2hTyzt4gxJ6R
```
Register plugin
```bash
vault plugin register   -sha256=`shasum -a 256 vlt_plugins/vault-database-wrapper | cut -d' ' -f1`   database   vault-database-wrapper
```
enable in UI http://127.0.0.1:8200/ui/vault/secrets/

create new connection
```bash
vault write database/config/mydbXXX   plugin_name=vault-database-wrapper   connection_url="http://172.17.0.1:1234"   username="admin"   password="password"   allowed_roles="readonly"

vault write database/roles/readonly   db_name=mydbXXX    default_ttl="1m"   max_ttl="2h"
vault read database/creds/readonly
Key                Value
---                -----
lease_id           database/creds/readonly/tllALE2iCbvGp19B6fiE7tmh
lease_duration     1m
lease_renewable    true
password           RSMzxNozLj2eIZO-JiuW
username           root

```

To debug requests, a fake HTTP server can be used.

```bash
wget https://github.com/svenstaro/dummyhttp/releases/download/v1.0.3/dummyhttp-1.0.3-x86_64-unknown-linux-musl
dummyhttp-1.0.3-x86_64-unknown-linux-musl --port 8888 -vv
```

A sample debug request should be:
```log
2025-34-28 19:34:52 172.17.0.2:40040 POST /userAdd HTTP/1.1
┌─Incoming request
│ POST /userAdd HTTP/1.1
│ Accept-Encoding: gzip
│ Content-Length: 223
│ Content-Type: application/json
│ Host: 172.17.0.1:8888
│ User-Agent: Go-http-client/1.1
│ Body:
│ {
│   "admin_password": "password",
│   "admin_username": "admin",
│   "commands": null,
│   "connection_url": "http://172.17.0.1:1234",
│   "expiration": "2025-03-28T17:35:57Z",
│   "password": "DsTjO2RBcbXvIE9-0CZE",
│   "role_name": "readonly",
│   "username": "root"
│ }
┌─Outgoing response
│ HTTP/1.1 200 OK
│ Content-Length: 9
│ Content-Type: text/plain; charset=utf-8
│ Date: Fri, 28 Mar 2025 19:34:52 +0200
│ Body:
│ dummyhttp
```
When lease time expired
```log
2025-35-28 19:35:50 172.17.0.2:41738 POST /userDelete HTTP/1.1
┌─Incoming request
│ POST /userDelete HTTP/1.1
│ Accept-Encoding: gzip
│ Content-Length: 114
│ Content-Type: application/json
│ Host: 172.17.0.1:8888
│ User-Agent: Go-http-client/1.1
│ Body:
│ {
│   "admin_password": "password",
│   "admin_username": "admin",
│   "connection_url": "http://172.17.0.1:1234",
│   "username": "root"
│ }
┌─Outgoing response
│ HTTP/1.1 200 OK
│ Content-Length: 9
│ Content-Type: text/plain; charset=utf-8
│ Date: Fri, 28 Mar 2025 19:35:50 +0200
│ Body:
│ dummyhttp
```

After that, you can create any service that will receive requests for /userAdd, /userDelete, /userUpdate, and then pass them on to the necessary database/service.