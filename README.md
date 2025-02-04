# Socialat backend
## Setup

### Database ([Posgresql](https://www.postgresql.org/))

When you have Posgresql, you need create new db with name **socialat**

### Config env

Create new yaml config in `./main/config.yaml`. a sample config file can be located at `sample/sample_config.yaml`

in your config.yaml edit the db section to match your environment settings
`db:
  dns: "host=<host> user=<user> password=<password> dbname=socialat port=<port> sslmode=disable TimeZone=Asia/Shanghai"`

## Running socialat (Linux | MacOS | Window):

### Terminal

Run `go run ./cmd/socialat --config=./main/config.yaml`

### Makefile

You can create `Makefile` and add command to makefile like this

```
.PHONY:
up:
	go run ./cmd/socialat --config=./main/config.yaml

```
after that run `make up`
