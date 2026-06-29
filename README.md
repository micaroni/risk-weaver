# Risk Weaver
An AI-assisted Kubernetes security analytics platform that correlates policy violations, vulnerabilities, identity exposure, and runtime anomalies into explainable workload risk assessments.

# Usage Guide
## PostgreSQL startup
PostgreSQL runs locally in a Docker container.
From the project root:
1. Start PostgreSQL:

   ```bash
   docker compose up -d
   ```

2. Verify that the container is running:

   ```bash
   docker compose ps
   ```

3. View PostgreSQL logs:

   ```bash
   docker compose logs postgres
   ```

4. Open a PostgreSQL shell:

   ```bash
   docker compose exec postgres \
     psql -U risk_weaver_app -d risk_weaver
   ```

5. Exit the PostgreSQL shell:

   ```text
   \q
   ```

6. Stop PostgreSQL:

   ```bash
   docker compose down
   ```

Running `docker compose down` does not delete the PostgreSQL data volume.
To stop the container and delete all local PostgreSQL data:

```bash
docker compose down -v
```

> **Warning:** The `-v` option permanently deletes the local PostgreSQL volume and all data stored in it.

## Environment configuration
Create a `.env` file in the project root with the following database settings:

```dotenv
POSTGRES_DB=risk_weaver
POSTGRES_USER=risk_weaver_app
POSTGRES_PASSWORD=local-development-only

DATABASE_URL=postgresql://risk_weaver_app:local-development-only@localhost:5433/risk_weaver?sslmode=disable

GOOSE_DRIVER=postgres
GOOSE_DBSTRING=host=localhost port=5433 user=risk_weaver_app password=local-development-only dbname=risk_weaver sslmode=disable
GOOSE_MIGRATION_DIR=./db/migrations
```

Ensure `.env` is included in `.gitignore`:

```gitignore
.env
```

## Using migrations with Goose
Run all Goose commands from the project root.
### Create a migration
Create a new SQL migration file:

```bash
goose -dir db/migrations create <migration_name> sql
```

Example:

```bash
goose -dir db/migrations create create_workloads sql
```

This creates a timestamped migration file similar to:

```text
db/migrations/20260626153000_create_workloads.sql
```

Goose creates a migration file, not a database table. Define the schema change inside the generated migration file.

Example:

```sql
-- +goose Up

CREATE TABLE workloads (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    namespace TEXT NOT NULL,
    environment TEXT NOT NULL,
    owner TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down

DROP TABLE workloads;
```

### Apply migrations manually

Load the environment variables:

```bash
set -a
source .env
set +a
```

Apply all pending migrations:

```bash
goose \
  -dir db/migrations \
  "$GOOSE_DRIVER" \
  "$GOOSE_DBSTRING" \
  up
```

The database package also automatically applies any pending migrations when the application starts.

### Check migration status

```bash
goose \
  -dir db/migrations \
  "$GOOSE_DRIVER" \
  "$GOOSE_DBSTRING" \
  status
```

### Check the current migration version

```bash
goose \
  -dir db/migrations \
  "$GOOSE_DRIVER" \
  "$GOOSE_DBSTRING" \
  version
```

### Roll back the latest migration

```bash
goose \
  -dir db/migrations \
  "$GOOSE_DRIVER" \
  "$GOOSE_DBSTRING" \
  down
```

### Apply one migration at a time

```bash
goose \
  -dir db/migrations \
  "$GOOSE_DRIVER" \
  "$GOOSE_DBSTRING" \
  up-by-one
```

## Verify the database schema

List all tables:

```bash
docker compose exec postgres \
  psql -U risk_weaver_app -d risk_weaver \
  -c "\dt"
```

Inspect the `workloads` table:

```bash
docker compose exec postgres \
  psql -U risk_weaver_app -d risk_weaver \
  -c "\d workloads"
```

Open an interactive PostgreSQL shell:

```bash
docker compose exec postgres \
  psql -U risk_weaver_app -d risk_weaver
```

Useful PostgreSQL shell commands:

```text
\dt
\d workloads
\q
```

## Migration guidelines
### Do
* Create a new migration for every schema change.
* Keep migrations small and focused.
* Include both `Up` and `Down` operations when practical.
* Review generated migration files before applying them.
* Run `goose status` to verify migration state.
* Commit migration files to source control.
* Trust Goose to skip migrations that have already been applied.
### Do not
* Edit an old migration after it has been applied to a shared or persistent database.
* Manually change the database schema outside migrations.
* Delete an applied migration file to undo a schema change.
* Assume that creating a migration file applies it immediately.
* Delete the PostgreSQL volume unless the local database should be completely reset.

Goose records applied migrations in the `goose_db_version` table and skips migrations that have already been applied.

# Project Contract
## Version 0.1
The system accepts a workload and security findings, calculates a deterministic risk score, stores results, and returns an assessment with an explanation.
 ### What is "done"?
A user can submit a JSON object to the `/workloads` endpoint and receive a workload ID in response.

