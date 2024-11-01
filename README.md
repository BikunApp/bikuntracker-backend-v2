## Getting started

First, copy the content in `.env.sample` to a new `.env` file. Then, compose the docker container that will run the database server on port specified on `.env`

```
make up
```

or

```
docker compose up -d
```

Lastly, to run the development server

```
make dev
```

or

```
air
```

## Interfaces

Golang does not allow import cycles, to counter that we define interfaces for each **Handler, Service, Repository and Util** in`app/interfaces`. See `app/interfaces/auth.go` for some example, any other reference to another module's instance will use this `interfaces.SomeInstance` interface type

For example, if a module called `problem` uses any sort of authentication, say `AuthService`, we will define the `authService` typing as `interfaces.AuthService` so that the problem module does not have any dependency to the auth module

## DTO (Data Transfer Object)

Any definition of request body types and response types will be defined on `app/dto/auth.go`

## Utility

- `config.go` extracts the `.env` file and initialized a `Config` object with the datas extracted from the environment file
- `logger.go` this middleware logs any request that goes in the server
- `parser.go` this utility file handles any sort of request parsing and response encoding
- `utils/middleware` folder hosts all the middlewares used within the app

## Migrations

A migration is a series of changes to a database (be it of the table, of the schema, or anything related to the database)

A migration consists of an up migration and a down migration, alongside a sequence id of the migration

To create a migration, run

```
make migration [name_of_migration]
```

This will then create an up migration file and down migration file, like so:

```
db/migrations/000001_create_user_table.down.sql
db/migrations/000001_create_user_table.up.sql
```

The up migration file specifies how the database should handle an update/change in the database, and the down migration file specifies how to undo said changes

Here is the example of a up migration file

```psql
CREATE TABLE account (
  id SERIAL NOT NULL UNIQUE,
  name VARCHAR(256) NOT NULL,
  npm VARCHAR(64),
  email VARCHAR(64),
  role VARCHAR(16) DEFAULT 'default',
  created_at BIGINT DEFAULT EXTRACT(EPOCH FROM CURRENT_TIMESTAMP),
  updated_at BIGINT DEFAULT EXTRACT(EPOCH FROM CURRENT_TIMESTAMP),
  PRIMARY KEY (id, npm)
);

CREATE TRIGGER update_account_updated_at BEFORE UPDATE ON account FOR EACH ROW EXECUTE PROCEDURE update_modified_column();
```

And an example of a down migration file

```psql
DROP TRIGGER IF EXISTS update_account_updated_at ON account;
DROP TABLE IF EXISTS account;
```

To apply unapplied migration(s), do

```
make migrate
```

To undo migration(s), do

```
make migrate-down
```

To summarize, here are some of the available commands

```
# Create a migration
make migration [name_of_migration]

# Apply all migrations
make migrate

# Apply n migrations
make migrate -- -steps n

# Revert all migrations
make migrate-down

# Revert n migrations
make migrate-down -- -steps n
```

## Database Conventions

Use `snake_case` as the primary convention, each table name and table attributes should be written in snake case

In the migration files, each table will contain an INTEGER `id`, BIGINT `created_at` and BIGINT `updated_at` field

```sql
id SERIAL NOT NULL UNIQUE
created_at BIGINT DEFAULT EXTRACT(EPOCH FROM CURRENT_TIMESTAMP)
updated_at BIGINT DEFAULT EXTRACT(EPOCH FROM CURRENT_TIMESTAMP)
```

- `id` will be the serial primary key (simple incremental id)
- `created_at` and `updated_at` is a unix timestamp, **make sure this field is `BIGINT`** (Refer [here](https://stackoverflow.com/questions/11799160/postgresql-field-type-for-unix-timestamp) to see more details)

The first migration file contains a trigger function that changes the field `obj.updated_at` to the current time in unix. Each of the migration files after that has a table that stores `created_at` and `updated_at` should create a trigger that runs on row updates

```sql
CREATE TRIGGER [trigger_name] BEFORE UPDATE ON [table_name] FOR EACH ROW EXECUTE PROCEDURE update_modified_column();
```

Keep in mind that the `update_modified_column()` function is defined `db/migrations/000001_create_trigger_for_updated_at_field.up.sql`, change the field name in this file

### Cleaning up

To drop the trigger functions in the down migrations, make sure that you delete the trigger function first before you delete the table or the migration will throw an error.

#### Example down migration

```sql
DROP TRIGGER update_achievement_updated_at ON achievement;
DROP TABLE achievement;

DROP TRIGGER update_achievement_category_updated_at ON achievement_category;
DROP TABLE achievement_category;
```

## Handler, Service, Repository, Util files

- Handler files are essentially controllers, these handlers will handle incoming requests to certain urls (each request object can be accessed through `r *http.Request`)
- Service files handle the bussiness logic part of the application
- Repository files handle any sort of database logic, `GetAllBuses`, `CreateBus`, `UpdateBus`, etc
- Util files handle any sort of additional logic that makes use of repository or service files, since we can use dependency injection to inject service/repo files to the utility entity

## Custom Route Parsing

Check out `utils/middleware` to see all the available middlewares, some of note are the one that enables authentication to be used within the app, `utils/middleware/jwt.go` and `utils/middleware/role-protect.go`. See each file for more details and see the example of its usage in `main.go`
