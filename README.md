## Bank service

- Create and manage bank accounts, which are composed of owner’s name, balance, and currency. 
- Record all balance changes to each of the account. So every time some money is added to or subtracted from the account, an account entry record will be created. 
- Perform a money transfer between 2 accounts. This should happen within a transaction, so that either both accounts’ balance are updated successfully or none of them are.

## Setup local development

### Install tools

- [Docker desktop](https://www.docker.com/products/docker-desktop)
- [PostgreSQL](https://www.postgresql.org)
- [Golang](https://golang.org/)
- [Migrate](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate)
- [sqlc](https://github.com/kyleconroy/sqlc#installation)
- [Gomock](https://github.com/golang/mock)

### Setup infrastructure

- Start postgres container:
    ```bash
    make postgres
    ```

- Create bank database:
    ```bash
    make create-db
    ```

- Run db migration up all versions:
    ```bash
    make migrate-up
    ```

- Run db migration up 1 version:
    ```bash
    make migrate-up1
    ```

- Run db migration down all versions:
    ```bash
    make migrate-down
    ```

- Run db migration down 1 version:
    ```bash
    make migrate-down1
    ```

### Documentation

### How to generate code
- Generate SQL CRUD with sqlc:
    ```bash
    make sqlc
    ```

- Generate DB mock with gomock:
    ```bash
    make mock
    ```

- Create a new db migration:
    ```bash
    make new_migration name=<migration_name>
    ```

### How to run
- Run server:
    ```bash
    make server
    ```

- Run test:
    ```bash
    make test
    ```