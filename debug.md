```
docker exec -it vault sh
/ # cd /vault/plugins/
/vault/plugins # ldd vault-database-wrapper 
	/lib64/ld-linux-x86-64.so.2 (0x7cade9a09000)
	libc.so.6 => /lib64/ld-linux-x86-64.so.2 (0x7cade9a09000)
Error relocating vault-database-wrapper: __vfprintf_chk: symbol not found
Error relocating vault-database-wrapper: __fprintf_chk: symbol not found
/vault/plugins # rm vault-database-wrapper 
/vault/plugins # ldd vault-database-wrapper 
/lib/ld-musl-x86_64.so.1: vault-database-wrapper: Not a valid dynamic program
/vault/plugins # rm vault-database-wrapper 
/vault/plugins # ldd vault-database-wrapper 
/lib/ld-musl-x86_64.so.1: vault-database-wrapper: Not a valid dynamic program
/vault/plugins #  
 ```
 
 fix
 ```
 CGO_ENABLED=0 GOOS=linux  go build .
 or
 CGO_ENABLED=0 go build -ldflags='-extldflags=-static'  .
 ```