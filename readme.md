# `db_explorer` Program

This simple web service is a MySQL database manager that allows CRUD operations (create, read, update, delete) over HTTP.

In this task, we continue practicing HTTP and database interaction skills.

*In this assignment, global variables are not allowed. Store everything you need in struct fields captured by a closure.*

## API behavior (from a user perspective)

- `GET /` — returns the list of all tables (that can be used in further requests)
- `GET /$table?limit=5&offset=7` — returns 5 records (`limit`) starting from record 7 (`offset`) from table `$table`. Defaults: `limit=5`, `offset=0`
- `GET /$table/$id` — returns the record details or `404`
- `PUT /$table` — creates a new record; record data is provided in the request body (POST parameters)
- `POST /$table/$id` — updates an existing record; data is provided in the request body (POST parameters)
- `DELETE /$table/$id` — deletes a record
- `GET`, `PUT`, `POST`, `DELETE` are the HTTP methods used to send requests

## Program requirements

- Implement routing manually — do not use external libraries.
- Full dynamic behavior:
  - During `NewDbExplorer` initialization, read table and column metadata from the database (queries are below).
  - Use this metadata for validation and processing.
  - No hardcoded validation rules for specific tables/fields.
  - If a third table is added, the service should work with it automatically.
- Assume that the list of tables does not change while the program is running.
- Build SQL queries dynamically, and extract data dynamically as well — there is no fixed list of fields.
- Validation should be basic: `string`, `int`, `float`, `null`.
  - Remember: when JSON is unmarshaled into `interface{}`, numbers become `float64` unless special options are used.
- Use only `database/sql`. You will receive a ready-to-use database connection as input.
  - No ORM and no similar abstractions.
- Field names in the API must match database field names exactly.
- On unexpected errors, return HTTP `500`.
- Do not forget about SQL injection protection.
- Unknown fields should be ignored.
- Global variables are forbidden in this assignment.

## Useful SQL queries

```sql
SHOW TABLES;
SHOW FULL COLUMNS FROM `$table_name`;
```

## Hints

- `row` objects contain not only values, but also metadata:
  - https://golang.org/pkg/database/sql/#Rows.ColumnTypes
- You will actively use empty interfaces (`interface{}`).
- Pay attention to `null` handling (there is a test case where a value is missing and no DB default exists).
- You will need to extract an unknown number of fields from rows — think about how to use empty interfaces for that.
- The easiest way to run MySQL locally is via Docker:

```bash
docker run -p 3306:3306 -v $(PWD):/docker-entrypoint-initdb.d -e MYSQL_ROOT_PASSWORD=1234 -e MYSQL_DATABASE=golang -d mysql
```
