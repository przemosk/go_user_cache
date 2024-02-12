# USER CACHE

## REQUIRMENTS

- Go
- Redis
- MySQL

## PREPARATION

In file `user.sql` you will find out commands which should be executed in your mysql cli

in `main()` function there is method `fillRedisCacheDatabase(ctx, rdb, 2)` which filling up local Redis server with dummy data. Last argument is reponsible for number of created cached users.

## RUNNING

Running script:

```bash
go run main.go --race
```

## SPECS

Running specs:

```bash
go test --race
```
