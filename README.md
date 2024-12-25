1. Create new migrate
```sh
migrate create -ext sql -dir db/migration -seq <migration_name>
```

2. Apply the new migration
```sh
make migrateup
```

3. Regenerate gRPC code
``` sh
protoc --proto_path=proto --go_out=pb --go_opt=paths=source_relative \
    --go-grpc_out=pb --go-grpc_opt=paths=source_relative \
    --grpc-gateway_out=pb --grpc-gateway_opt=paths=source_relative \
    proto/*.proto
```