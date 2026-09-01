# ViComp

Stands for Victim's Compensation and is used to track and send invoices.

Single-user web app: Go (`net/http` + `html/template`), HTMX, Postgres,
Gotenberg for PDF rendering. Runs entirely in containers.

## Models

| Model      | Notes                                                                 |
|------------|----------------------------------------------------------------------|
| County     | Payer agency: name, contact, address, phone, email.                  |
| Clinician  | Provider: name, credentials, NPI, tax id, address, phone, email. Has clients. |
| Client     | Linked to one County and one Clinician. Has sessions and invoices.   |
| Session    | Dated appointment for a client: `amount`, `duration_minutes`, `cpt_code` (defaults to 90834). |
| Invoice    | Created by rolling up every not-yet-invoiced session for a client.   |

**Settings** (`/settings`, a single row):

- **Practice / billing details** — practice name, address, phone, email, tax id,
  NPI. These form the invoice "From" block. Any field a clinician fills in on
  their own record overrides the setting for that clinician's invoices; fields
  the clinician leaves blank fall back to Settings.
- **Make checks payable to** — printed in the invoice PDF footer.
- **Session defaults** — rate, length in minutes, CPT code — pre-fill a new
  session; each stays editable per session.

Once a session is attached to an invoice it is locked: it can't be edited,
deleted, or invoiced again. Deleting an invoice releases its sessions (they
become uninvoiced and can be billed on a new invoice).

## Run

```sh
make up        # build + start everything in the background
make logs      # follow app logs
make down      # stop (keeps data)   /   make reset  (stop + wipe db)
```

`make` on its own lists every target. Plain `docker compose up --build` works too.

- App:       http://localhost:8090
- Postgres:  localhost:5433  (user/pass/db all `vicomp`)
- Gotenberg: http://localhost:3001

Schema is created automatically on startup from `migrations/*.sql`.

The `app` container runs [air](https://github.com/air-verse/air): the repo is
bind-mounted into it, so saving a `.go`, `.html`, `.css`, or `.sql` file
rebuilds and restarts the server in place — no `make` command needed.

## Invoice flow

1. Create a County and a Clinician.
2. Create a Client, linking it to that County and Clinician.
3. Add Sessions for the Client (date + amount).
4. Invoices → New Invoice → pick the Client. The uninvoiced sessions and a
   total preview load via HTMX. Click **Generate Invoice**.
5. On the invoice page, **Download PDF** renders it through Gotenberg. The PDF
   carries the clinician's details, the county's details, the client info, one
   line per session (date, minutes, amount) and the total.

A Gmail OAuth "send" step is intentionally out of scope for now — download only.

## Local development (run the Go app on the host instead of in Docker)

```sh
make dev
```

Starts db + gotenberg in Docker and runs `go run ./cmd/vicomp` on the host
against them. Equivalent to:

```sh
docker compose up -d db gotenberg
DATABASE_URL='postgres://vicomp:vicomp@localhost:5433/vicomp?sslmode=disable' \
GOTENBERG_URL='http://localhost:3001' \
go run ./cmd/vicomp
```

## Layout

```
cmd/vicomp/main.go      entrypoint: config, db connect, migrate, serve
internal/db             pool + migration runner
internal/models         plain structs
internal/store          one file per model, raw SQL via pgx
internal/pdf            Gotenberg HTML->PDF client
internal/web            routes + handlers (one file per model)
templates/              layout.html + pages/*.html (html/template)
static/app.css
migrations/*.sql        001 schema · 002 settings + session cpt_code · 003 practice details
```
