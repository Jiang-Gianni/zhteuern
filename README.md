# zhteuern

Tools needed:

* Golang: https://go.dev/dl/

* templ: https://github.com/a-h/templ

```bash
go install github.com/a-h/templ/cmd/templ@latest
```

* sqlc: https://github.com/sqlc-dev/sqlc

```bash
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
```

To start, clone the repo:

```bash
go mod tidy && templ generate lazy -v && sqlc generate
```

```bash
go run watch/fw/fw.go
```