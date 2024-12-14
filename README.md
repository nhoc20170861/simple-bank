1. Create new migrate
```sh
migrate create -ext sql -dir db/migration -seq add_users
```

2. Apply the new migration
```sh
make migrateup
```