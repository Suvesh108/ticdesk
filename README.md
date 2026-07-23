# ticDesk — IT Helpdesk & Ticketing System

![Go](https://img.shields.io/badge/Go-1.22-00ADD8?style=flat&logo=go)
![HTMX](https://img.shields.io/badge/HTMX-1.9-3366CC?style=flat)
![Alpine.js](https://img.shields.io/badge/Alpine.js-3.x-8BC0D0?style=flat&logo=alpine.js)
![TailwindCSS](https://img.shields.io/badge/Tailwind_CSS-3.x-38B2AC?style=flat&logo=tailwind-css)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-4169E1?style=flat&logo=postgresql)
![Docker](https://img.shields.io/badge/Docker-Compose-2496ED?style=flat&logo=docker)

**ticDesk** is a high-performance, server-rendered IT Helpdesk & Ticket Management System built with **Go**, **HTMX**, **Alpine.js**, **Tailwind CSS**, and **PostgreSQL**. Designed as a modern alternative to legacy admin portals, it leverages an HTML-first architecture with zero-reload HTMX partial swaps, double-barrier RBAC security, and real-time aggregate analytics.

---

## 🚀 Key Features

- 🔒 **Authentication & Session Security**: Password hashing via `bcrypt`, HTTP-only session cookies via `scs`, and CSRF protection.
- 🛡️ **Role-Based Access Control (RBAC)**:
  - **Admin**: Full system access, User & Role Management (`/admin/users`), org-wide dashboard, ticket deletion.
  - **Support**: Queue management, team reassignment, status/priority changes, staff-only internal notes.
  - **Customer**: Ticket creation, view/comment on own tickets only, file attachments.
- ⚡ **Zero-Reload HTMX Partial Swaps**: Instant mutations for ticket status, priority levels, and support agent assignment returning pure HTML fragments.
- 💬 **Threaded Discussion & Internal Notes**: Public comment threads for customer communication + staff-only internal notes (`is_internal = true`) hidden from customer accounts.
- 📎 **Secure File Attachment Storage**: 10MB file upload validation, MIME whitelist, and disk storage abstraction (`storage_service.go`).
- 📊 **Real-time Analytics & Stats JSON**: Live PostgreSQL aggregate queries for status counts, priority distributions, SLA resolution averages, and agent workload queues.
- 📧 **Asynchronous Email Worker Queue**: Non-blocking in-process goroutine queue (`chan EmailJob`) sending HTML notification emails via SMTP (MailHog in dev).
- 🔍 **Fuzzy Trigram Search & Pagination**: PostgreSQL `pg_trgm` trigram search on titles and descriptions with 300ms debounced HTMX search inputs.
- 🎨 **Apple/Linear-Grade Design System**: Dark slate glassmorphism palette (`#090d16`), Inter typography, custom SVG icons, and responsive layouts.

---

## 🏗️ High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        Browser                              │
│  HTMX (swaps HTML fragments) + Alpine.js (local UI state)   │
│  Tailwind CSS (compiled static styling)                    │
└───────────────────────────┬─────────────────────────────────┘
                            │ HTTP (HTML fragments + full pages)
┌───────────────────────────▼─────────────────────────────────┐
│                     Go Backend (net/http + chi router)       │
│  ┌───────────┐ ┌────────────┐ ┌─────────────┐ ┌───────────┐ │
│  │  Handlers  │ │ Middleware │ │  Templates  │ │ Services  │ │
│  │ (routes)   │ │ (auth/RBAC)│ │ (html/tmpl) │ │(business) │ │
│  └───────────┘ └────────────┘ └─────────────┘ └───────────┘ │
│  ┌───────────────────────┐  ┌──────────────────────────┐   │
│  │   Repository layer    │  │  Background email queue  │   │
│  │   (pgx pool queries)  │  │  (chan EmailJob worker)  │   │
│  └───────────────────────┘  └──────────────────────────┘   │
└───────────────────────────┬─────────────────────────────────┘
                            │
               ┌────────────▼───────────┐      ┌────────────────┐
               │       PostgreSQL       │      │   Local Disk   │
               │ (tickets, users, etc.) │      │ File Attachments│
               └────────────────────────┘      └────────────────┘
                            │
               ┌────────────▼───────────┐
               │    MailHog SMTP Server │
               │   (dev email testing)  │
               └────────────────────────┘
```

---

## 📂 Project Structure

```
ticDesk/
├── cmd/
│   └── server/
│       └── main.go                 # Application entrypoint & dependency wiring
├── internal/
│   ├── auth/
│   │   ├── session.go              # Session cookie management (SCS)
│   │   ├── password.go             # Bcrypt hashing & verification
│   │   └── middleware.go           # RequireAuth & RequireRole middleware
│   ├── config/
│   │   └── config.go               # Environment variables configuration
│   ├── db/
│   │   └── migrations/             # SQL migrations (0001 to 0004)
│   │       ├── 0001_init.sql
│   │       ├── 0002_tickets.sql
│   │       ├── 0003_comments_and_attachments.sql
│   │       └── 0004_search_and_indices.sql
│   ├── handlers/
│   │   ├── auth_handler.go         # Authentication handlers (Login/Register/Logout)
│   │   ├── ticket_handler.go       # Core ticket CRUD + HTMX mutations
│   │   ├── comment_handler.go      # Discussion replies & internal notes
│   │   ├── attachment_handler.go   # Secure attachment download handler
│   │   ├── dashboard_handler.go    # Analytics stats + JSON endpoint
│   │   └── admin_handler.go        # User management & role assignment
│   ├── models/
│   │   └── models.go               # Go models & data structures
│   ├── repository/
│   │   ├── user_repo.go            # User database operations
│   │   ├── ticket_repo.go          # Ticket database operations & analytics
│   │   ├── comment_repo.go         # Comment database operations
│   │   └── attachment_repo.go      # Attachment database operations
│   ├── services/
│   │   ├── storage_service.go      # Local file storage abstraction
│   │   └── email_service.go        # Async SMTP worker queue
│   └── router/
│       └── router.go               # Chi router setup & RBAC route groups
├── web/
│   ├── templates/
│   │   ├── layouts/
│   │   │   └── base.html           # Main HTML base layout & design system
│   │   ├── pages/
│   │   │   ├── login.html
│   │   │   ├── register.html
│   │   │   ├── dashboard.html
│   │   │   ├── ticket_list.html
│   │   │   ├── ticket_new.html
│   │   │   ├── ticket_detail.html
│   │   │   └── admin_users.html
│   │   └── partials/               # HTMX fragment targets
│   │       ├── ticket_status_badge.html
│   │       ├── ticket_priority_badge.html
│   │       ├── ticket_assignee.html
│   │       ├── ticket_table.html
│   │       └── comment_list.html
│   └── static/
│       └── uploads/                # Local attachment storage directory
├── .env.example
├── docker-compose.yml              # PostgreSQL 16 + MailHog dev environment
├── Dockerfile                      # Multi-stage production Go container build
├── go.mod
├── go.sum
└── Makefile                        # Project build & management commands
```

---

## ⚡ Quick Start & Setup

### Prerequisites
- [Docker & Docker Compose](https://www.docker.com/)
- [Go 1.22+](https://go.dev/) (for running locally outside Docker)

### Running with Docker Compose (Recommended)

1. **Clone the Repository**:
   ```bash
   git clone https://github.com/Suvesh108/ticdesk.git
   cd ticdesk
   ```

2. **Start Services via Docker**:
   ```bash
   docker compose up -d --build
   ```

3. **Access Services**:
   - **ticDesk Web Application**: [http://localhost:8081](http://localhost:8081)
   - **MailHog Web UI (Dev Emails)**: [http://localhost:8025](http://localhost:8025)
   - **PostgreSQL Database**: `localhost:5432` (User: `ticdesk`, Password: `ticdesk_secret`, DB: `ticdesk`)

---

## 🔐 Default Credentials

| Email | Password | Role | Description |
|:---|:---|:---:|:---|
| `admin@ticdesk.com` | `password123` | `Admin` | System Administrator with full access |

*You can also register a new account on the `/register` page, which defaults to the `Customer` role.*

---

## 🛠️ API & Web Routes

| Method | Endpoint | Description | Access |
|:---:|:---|:---|:---:|
| `GET` | `/login` | Render login page | Public |
| `POST` | `/login` | Authenticate & set session cookie | Public |
| `POST` | `/logout` | Invalidate session | Authenticated |
| `GET` | `/dashboard` | Render role-aware dashboard | Authenticated |
| `GET` | `/dashboard/stats.json` | Returns JSON aggregate stats | Authenticated |
| `GET` | `/tickets` | List tickets with search & filters | Authenticated |
| `GET` | `/tickets/new` | Render ticket creation form | Authenticated |
| `POST` | `/tickets` | Create ticket | Authenticated |
| `GET` | `/tickets/{id}` | Ticket detail & discussion thread | Authenticated |
| `PATCH` | `/tickets/{id}/status` | Update status (HTMX partial swap) | Support / Admin |
| `PATCH` | `/tickets/{id}/priority` | Update priority (HTMX partial swap) | Support / Admin |
| `PATCH` | `/tickets/{id}/assign` | Update assignee (HTMX partial swap) | Support / Admin |
| `GET` | `/tickets/{id}/comments` | Render comment thread partial | Authenticated |
| `POST` | `/tickets/{id}/comments` | Post reply or internal note | Authenticated |
| `GET` | `/attachments/{id}` | Download file attachment | Authenticated |
| `GET` | `/admin/users` | Render user management console | Admin Only |
| `POST` | `/admin/users/{id}/role` | Change user role | Admin Only |
| `POST` | `/admin/users/{id}/deactivate` | Toggle account active status | Admin Only |

---

## 📄 License

This project is open-source under the [MIT License](LICENSE).
