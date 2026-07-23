# ticDesk — IT Helpdesk & Ticketing System
### Full Architecture + Antigravity Build Prompt

Stack: **Go** (backend) + **HTMX** + **Alpine.js** + **Tailwind CSS** (frontend, server-rendered) + **PostgreSQL** (DB)

---

## 1. High-Level Architecture

This is a classic **server-rendered HTMX app**, not an SPA — Go renders HTML fragments, HTMX swaps them into the DOM, Alpine.js handles tiny client-side interactivity (dropdowns, modals, toggles). No separate frontend build step, no JSON API layer needed for the UI itself (though you'll expose a couple of JSON endpoints for the dashboard charts).

```
┌─────────────────────────────────────────────────────────────┐
│                        Browser                              │
│  HTMX (swaps HTML fragments) + Alpine.js (local UI state)   │
│  Tailwind CSS (compiled once, static file)                   │
└───────────────────────────┬───────────────────────────────────┘
                             │ HTTP (HTML fragments + full pages)
┌───────────────────────────▼───────────────────────────────────┐
│                     Go Backend (net/http + chi router)         │
│  ┌───────────┐ ┌────────────┐ ┌─────────────┐ ┌─────────────┐ │
│  │  Handlers  │ │ Middleware │ │  Templates  │ │  Services   │ │
│  │ (routes)   │ │ (auth/RBAC)│ │ (html/tmpl) │ │ (business)  │ │
│  └───────────┘ └────────────┘ └─────────────┘ └─────────────┘ │
│  ┌───────────────────────┐  ┌────────────────────────────┐    │
│  │  Repository layer     │  │  Background worker(s)       │    │
│  │  (pgx / sqlc queries) │  │  (email queue, notifications)│   │
│  └───────────────────────┘  └────────────────────────────┘    │
└───────────────────────────┬───────────────────────────────────┘
                             │
                 ┌───────────▼────────────┐      ┌─────────────────┐
                 │      PostgreSQL         │      │  Local disk /   │
                 │  (tickets, users, etc.) │      │  S3-compatible  │
                 └─────────────────────────┘      │  file storage   │
                                                    └─────────────────┘
                             │
                 ┌───────────▼────────────┐
                 │   SMTP (email notifs)   │
                 │  e.g. Mailhog (dev) /   │
                 │  Resend/SendGrid (prod) │
                 └─────────────────────────┘
```

**Why this stack works well together:**
- Go gives you a single static binary, no Node runtime needed in prod.
- HTMX keeps almost all logic server-side — less duplicated validation logic vs a React+API split.
- Alpine.js fills the small gaps HTMX doesn't cover (client-only toggle state, character counters, etc.).
- Tailwind is compiled once at build time via the standalone CLI (no Node dependency required).

---

## 2. Folder Structure

```
ticDesk/
├── cmd/
│   └── server/
│       └── main.go                 # entrypoint, wires everything together
├── internal/
│   ├── auth/
│   │   ├── session.go              # session cookie management
│   │   ├── password.go             # bcrypt hash/verify
│   │   └── middleware.go           # RequireAuth, RequireRole(...)
│   ├── config/
│   │   └── config.go               # env var loading
│   ├── db/
│   │   ├── migrations/             # SQL migration files (goose or golang-migrate)
│   │   │   ├── 0001_init.sql
│   │   │   ├── 0002_tickets.sql
│   │   │   └── ...
│   │   ├── queries/                # sqlc query files (.sql)
│   │   └── sqlc/                   # generated Go code (do not hand-edit)
│   ├── handlers/
│   │   ├── auth_handler.go         # login/logout/register
│   │   ├── ticket_handler.go       # CRUD + status/priority changes
│   │   ├── comment_handler.go      # replies on tickets
│   │   ├── attachment_handler.go   # upload/download files
│   │   ├── dashboard_handler.go    # stats views + JSON for charts
│   │   ├── admin_handler.go        # user management, role assignment
│   │   └── search_handler.go       # filtered ticket list (HTMX partial)
│   ├── models/
│   │   └── models.go               # Go structs mirroring DB rows
│   ├── repository/
│   │   ├── user_repo.go
│   │   ├── ticket_repo.go
│   │   ├── comment_repo.go
│   │   └── attachment_repo.go
│   ├── services/
│   │   ├── ticket_service.go       # business rules (e.g. SLA calc, auto-assign)
│   │   ├── notification_service.go # decides when/what to email
│   │   └── storage_service.go      # local disk or S3 abstraction
│   ├── mailer/
│   │   └── mailer.go               # SMTP client wrapper + templates
│   └── router/
│       └── router.go               # chi routes, grouped by role
├── web/
│   ├── templates/
│   │   ├── layouts/
│   │   │   └── base.html
│   │   ├── pages/
│   │   │   ├── login.html
│   │   │   ├── dashboard.html
│   │   │   ├── ticket_list.html
│   │   │   ├── ticket_detail.html
│   │   │   └── admin_users.html
│   │   └── partials/               # HTMX fragment targets
│   │       ├── ticket_row.html
│   │       ├── ticket_table.html
│   │       ├── comment_item.html
│   │       ├── comment_list.html
│   │       └── toast.html
│   └── static/
│       ├── css/
│       │   ├── input.css           # Tailwind directives
│       │   └── output.css          # compiled (gitignored)
│       ├── js/
│       │   └── app.js              # tiny Alpine components/helpers
│       └── uploads/                # dev-only local attachment storage
├── .env.example
├── go.mod
├── go.sum
├── tailwind.config.js
├── Makefile                        # make dev / make migrate / make build
└── docker-compose.yml              # postgres + mailhog for local dev
```

---

## 3. Database Schema (PostgreSQL)

```sql
-- users
CREATE TYPE user_role AS ENUM ('admin', 'support', 'customer');

CREATE TABLE users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name          TEXT NOT NULL,
    email         TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    role          user_role NOT NULL DEFAULT 'customer',
    is_active     BOOLEAN NOT NULL DEFAULT true,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ticket enums
CREATE TYPE ticket_priority AS ENUM ('low', 'medium', 'high');
CREATE TYPE ticket_status   AS ENUM ('open', 'in_progress', 'resolved', 'closed');

CREATE TABLE categories (
    id    SERIAL PRIMARY KEY,
    name  TEXT UNIQUE NOT NULL          -- e.g. "Hardware", "Network", "Software"
);

CREATE TABLE tickets (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ticket_number SERIAL UNIQUE,        -- human-friendly #1024 style reference
    title         TEXT NOT NULL,
    description   TEXT NOT NULL,
    category_id   INT REFERENCES categories(id),
    priority      ticket_priority NOT NULL DEFAULT 'medium',
    status        ticket_status NOT NULL DEFAULT 'open',
    created_by    UUID NOT NULL REFERENCES users(id),   -- customer who raised it
    assigned_to   UUID REFERENCES users(id),            -- support agent
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    resolved_at   TIMESTAMPTZ
);

CREATE TABLE comments (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ticket_id   UUID NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
    author_id   UUID NOT NULL REFERENCES users(id),
    body        TEXT NOT NULL,
    is_internal BOOLEAN NOT NULL DEFAULT false, -- support-only notes, hidden from customer
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE attachments (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ticket_id   UUID REFERENCES tickets(id) ON DELETE CASCADE,
    comment_id  UUID REFERENCES comments(id) ON DELETE CASCADE,
    uploaded_by UUID NOT NULL REFERENCES users(id),
    file_name   TEXT NOT NULL,
    file_path   TEXT NOT NULL,          -- disk path or S3 key
    file_size   BIGINT NOT NULL,
    mime_type   TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE ticket_status_history (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ticket_id   UUID NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
    changed_by  UUID NOT NULL REFERENCES users(id),
    old_status  ticket_status,
    new_status  ticket_status NOT NULL,
    changed_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE notifications (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id),
    ticket_id   UUID REFERENCES tickets(id) ON DELETE CASCADE,
    type        TEXT NOT NULL,          -- 'ticket_assigned', 'new_comment', 'status_changed'
    is_read     BOOLEAN NOT NULL DEFAULT false,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- indexes that matter once ticket volume grows
CREATE INDEX idx_tickets_status ON tickets(status);
CREATE INDEX idx_tickets_priority ON tickets(priority);
CREATE INDEX idx_tickets_assigned_to ON tickets(assigned_to);
CREATE INDEX idx_tickets_created_by ON tickets(created_by);
CREATE INDEX idx_comments_ticket_id ON comments(ticket_id);
CREATE INDEX idx_tickets_title_trgm ON tickets USING gin (title gin_trgm_ops); -- needs pg_trgm ext, for search
```

---

## 4. Role-Based Access Matrix

| Action                              | Admin | Support | Customer |
|--------------------------------------|:-----:|:-------:|:--------:|
| Create ticket                        |  ✅   |   ✅    |    ✅    |
| View own tickets                     |  ✅   |   ✅    |    ✅    |
| View all tickets                     |  ✅   |   ✅    |    ❌    |
| Assign ticket to agent                |  ✅   |   ✅*   |    ❌    |
| Change priority                       |  ✅   |   ✅    |    ❌    |
| Change status                         |  ✅   |   ✅    |    ❌ (only re-open own resolved) |
| Comment (public)                      |  ✅   |   ✅    |    ✅    |
| Comment (internal note)               |  ✅   |   ✅    |    ❌    |
| Upload attachment                     |  ✅   |   ✅    |    ✅ (own tickets only) |
| Manage users / roles                  |  ✅   |   ❌    |    ❌    |
| View dashboard (org-wide stats)       |  ✅   |   ✅ (own queue) |  ❌ |
| Delete ticket                          |  ✅   |   ❌    |    ❌    |

\* Support can self-assign or reassign within their team; only Admin can reassign across teams.

Enforce this with a single `RequireRole(roles ...string)` middleware plus **row-level checks in the repository layer** (e.g. a customer's query for ticket detail must also filter `created_by = current_user_id`, never trust the route alone).

---

## 5. Core Routes

```
GET  /login                          → login page
POST /login                          → authenticate, set session cookie
POST /logout                         → destroy session

GET  /dashboard                      → role-aware dashboard (stats + charts)
GET  /dashboard/stats.json           → JSON for chart rendering (Alpine + a tiny chart lib)

GET  /tickets                        → ticket list page (filters via query params)
GET  /tickets/table                  → HTMX partial: filtered/paginated table body
GET  /tickets/new                    → new ticket form
POST /tickets                        → create ticket
GET  /tickets/{id}                   → ticket detail page
PATCH /tickets/{id}/status            → HTMX: change status (returns updated status badge partial)
PATCH /tickets/{id}/priority          → HTMX: change priority
PATCH /tickets/{id}/assign            → HTMX: assign to agent

GET  /tickets/{id}/comments           → HTMX partial: comment thread
POST /tickets/{id}/comments           → add comment/reply, returns new comment_item partial

POST /tickets/{id}/attachments        → upload file (multipart)
GET  /attachments/{id}                → download file (auth + ownership checked)

GET  /admin/users                     → user management page (admin only)
POST /admin/users/{id}/role           → change a user's role
POST /admin/users/{id}/deactivate     → disable account

GET  /notifications                   → HTMX partial: notification bell dropdown
POST /notifications/{id}/read         → mark as read
```

**HTMX pattern to lean on throughout:** every action that mutates state (status change, new comment, assign) returns the **already-updated HTML partial**, not a redirect or JSON. E.g. `PATCH /tickets/{id}/status` re-renders `partials/ticket_status_badge.html` and HTMX swaps it via `hx-target` — no full page reload, no client-side state to keep in sync.

---

## 6. Notifications (Email)

- `notification_service.go` listens for domain events (ticket created, assigned, commented, status changed) and:
  1. Writes a row to `notifications` (for the in-app bell icon).
  2. Enqueues an email job (simple in-process buffered channel + goroutine worker is enough for this scale — no need for a message broker).
- Dev: use **Mailhog** (in `docker-compose.yml`) to catch outgoing mail locally without hitting a real inbox.
- Prod: swap the SMTP host/creds to a real provider (SendGrid, Resend, or plain Gmail SMTP for a portfolio demo).
- Templates live in `internal/mailer/templates/` as simple `html/template` files (ticket_assigned.html, new_reply.html, status_changed.html).

---

## 7. Dashboard Stats

Query examples the dashboard handler needs (all filterable by role — support sees only their queue, admin sees org-wide):

- Open / In Progress / Resolved counts (grouped `COUNT(*) ... GROUP BY status`)
- Tickets by priority (for a bar chart)
- Average resolution time: `AVG(resolved_at - created_at)` for resolved tickets
- Tickets by category (for a pie/donut chart)
- Agent workload: open ticket count grouped by `assigned_to`

Render these with a lightweight approach: either server-side SVG (Go can generate simple bar/pie SVGs with no JS dependency) **or** a small JS chart lib (Chart.js via CDN) fed by the `/dashboard/stats.json` endpoint — either is fine for a portfolio piece; server-side SVG is more "in the spirit" of the HTMX-first approach if you want to minimize JS.

---

## 8. Search & Filters

`GET /tickets/table?status=open&priority=high&q=printer&assigned_to=<uuid>&page=2`

- Implement as a single handler that builds a dynamic SQL WHERE clause based on which query params are present (use `squirrel` or hand-roll a small query builder — avoid raw string concatenation for anything with user input).
- Debounce the search input on the client with a small Alpine/HTMX combo: `hx-trigger="keyup changed delay:400ms"` on the search box, targeting the `#ticket-table` partial.
- Full-text-ish search on title/description: `pg_trgm` extension + `ILIKE` or `similarity()` is enough at this scale — no need for Elasticsearch.

---

## 9. Auth

- Session-based auth (not JWT) — simplest fit for a server-rendered HTMX app. Store session ID in an HttpOnly, Secure cookie; session data (`user_id`, `role`) in a `sessions` table or in-memory store (e.g. `scs` session library for Go) backed by Postgres for persistence across restarts.
- Passwords hashed with bcrypt (cost 12).
- Middleware chain: `RequireAuth` → `RequireRole(...)` → handler.
- CSRF protection on all POST/PATCH/DELETE forms (HTMX sends the token as a header via a small Alpine/JS snippet reading a meta tag).

---

## 10. File Attachments

- Store files on local disk in dev (`web/static/uploads/{ticket_id}/{uuid}_{filename}`), abstracted behind a `storage_service.go` interface so swapping to S3-compatible storage later (e.g. MinIO, Backblaze) is a one-file change.
- Validate: max file size (e.g. 10MB), whitelist MIME types (images, PDFs, common office docs, logs/text).
- Serve downloads through a handler that checks ownership/role — never serve `uploads/` as a static directory directly, or access control is bypassed.

---

## 11. Suggested Build Order

1. Project scaffold + Docker Compose (Postgres + Mailhog) + migrations for `users`
2. Auth: register/login/logout/session middleware
3. Tickets: create + list + detail (no HTMX yet — plain page loads)
4. Layer in HTMX: convert list filters and status/priority changes to partial swaps
5. Comments + attachments
6. Role-based access matrix enforcement (middleware + repo-level checks)
7. Dashboard stats page
8. Email notifications (Mailhog first, real SMTP later)
9. Search/filter polish + pagination
10. Tailwind pass for visual polish (you already have a design language from your portfolio work — carry that navy/accent aesthetic if you want ticDesk to look distinctive rather than generic admin-panel)

---

## 12. Paste-Ready Prompt for Antigravity

Copy everything in the block below into Antigravity as your first scaffolding prompt.

```
I'm building "ticDesk", an IT Helpdesk & Ticketing System, as a portfolio project.
Location: C:\Users\Suvesh\Desktop\projects\ticDesk

TECH STACK (mandatory, do not substitute):
- Go (standard library net/http + chi router for routing)
- HTMX for all dynamic UI updates (server returns HTML partials, no JSON API for the UI)
- Alpine.js for small client-only interactivity (dropdowns, modals, toasts)
- Tailwind CSS (standalone CLI, no Node/PostCSS build step required)
- PostgreSQL as the database, accessed via pgx and sqlc for typed queries
- golang-migrate (or goose) for SQL migrations
- Session-based auth using the `scs` session library, bcrypt for password hashing
- Docker Compose for local Postgres + Mailhog (fake SMTP for dev email testing)

CORE FEATURES:
1. Auth: login/logout, session cookies, bcrypt password hashing, CSRF protection on mutating requests
2. Role-based access control with 3 roles: admin, support, customer
   - admin: full access, manage users/roles, org-wide dashboard, delete tickets
   - support: manage tickets (assign/reassign within team, change status/priority, internal notes), view own queue dashboard
   - customer: create tickets, view/comment on own tickets only, upload attachments to own tickets
3. Tickets: create/list/detail, priority (low/medium/high), status (open/in_progress/resolved/closed), category, assignment to a support agent, full status-change history log
4. Comments: threaded replies on a ticket, with an "internal note" flag visible only to admin/support (never to the customer)
5. File attachments: upload on tickets and comments, size limit 10MB, MIME whitelist (images, PDF, common office docs, plain text/log files), stored on local disk behind a storage_service abstraction so it can later swap to S3-compatible storage
6. Dashboard: role-aware stats — ticket counts by status/priority/category, average resolution time, per-agent workload for support/admin views
7. Search & filters: filter ticket list by status/priority/category/assigned agent, plus a debounced text search on title/description (use Postgres pg_trgm extension + ILIKE/similarity, not a separate search engine)
8. Email notifications: on ticket created, assigned, new comment, and status changed — queue via an in-process goroutine worker (no external message broker needed at this scale), send through SMTP (Mailhog in dev)

DATABASE:
Use this schema as the starting point (I've already designed it — please generate the corresponding migration files exactly, then generate sqlc query files for standard CRUD + the filtered/paginated ticket list query + dashboard aggregate queries):

[PASTE THE SQL SCHEMA FROM SECTION 3 OF THIS DOCUMENT HERE]

ARCHITECTURE / FOLDER STRUCTURE:
Follow this exact structure (I've planned it out already):

[PASTE THE FOLDER STRUCTURE FROM SECTION 2 OF THIS DOCUMENT HERE]

ROUTES:
Implement these routes first, using the pattern where every mutating action (status change, priority change, new comment, assignment) returns the already-updated HTML partial for HTMX to swap in — never a redirect or raw JSON for UI-facing actions:

[PASTE THE ROUTES FROM SECTION 5 OF THIS DOCUMENT HERE]

BUILD ORDER — please work through this incrementally and pause for my review after each numbered phase rather than generating everything at once:
1. Project scaffold + docker-compose.yml (Postgres + Mailhog) + go.mod + first migration for `users` table
2. Auth: register/login/logout/session middleware/CSRF
3. Tickets: create + list + detail as plain server-rendered pages (no HTMX yet)
4. Convert ticket list filters and status/priority changes to HTMX partial swaps
5. Comments + attachments
6. Enforce the role-based access matrix at both the middleware layer and the repository query layer (never trust route-level checks alone — e.g. a customer's ticket detail query must also filter by created_by)
7. Dashboard stats page
8. Email notifications wired to Mailhog
9. Search/filter polish + pagination
10. Tailwind visual pass

VISUAL DIRECTION:
Clean, modern SaaS admin-panel aesthetic — not generic Bootstrap-look. Off-white/light background option AND a dark mode toggle if reasonably easy with Tailwind. Sidebar nav + topbar with notification bell (HTMX-polled or SSE, your call on complexity). Status badges color-coded (open=blue, in_progress=amber, resolved=green, closed=gray), priority badges (low=gray, medium=orange, high=red).

Please start with Phase 1 only, and confirm the scaffold runs (`docker compose up` + `go run ./cmd/server` + migrations applied) before moving to Phase 2.
```

---

### Notes for you
- I split the prompt into phases on purpose — a single mega-prompt for a full-stack app like this tends to produce shakier output than incremental scaffolding with review checkpoints, especially for the RBAC and HTMX-partial patterns which are easy to get subtly wrong in one shot.
- If Antigravity's context window struggles with the full schema + folder structure + routes in one message, you can split this into 3 separate messages (schema → folder structure → routes → build order) and it'll still work fine since each phase references the earlier ones.
- Worth adding to your resume/portfolio README once built: a short section on *why* session auth over JWT here, and *why* HTMX over a SPA — interviewers for support-engineer-adjacent dev roles tend to like seeing that the tool choice was deliberate, not default.
