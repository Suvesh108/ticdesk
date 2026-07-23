<div align="center">

  <h1>⚡ ticDesk</h1>
  <p><strong>A Modern, High-Performance IT Helpdesk & Ticketing Platform</strong></p>

  <p>
    Built with <strong>Go 1.22</strong> • <strong>HTMX v2</strong> • <strong>Alpine.js</strong> • <strong>Tailwind CSS</strong> • <strong>PostgreSQL 16</strong>
  </p>

  <p>
    <a href="https://golang.org/"><img src="https://img.shields.io/badge/Backend-Go_1.22-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go"></a>
    <a href="https://htmx.org/"><img src="https://img.shields.io/badge/Frontend-HTMX_v2.0-3366CC?style=for-the-badge&logo=htmx&logoColor=white" alt="HTMX"></a>
    <a href="https://alpinejs.dev/"><img src="https://img.shields.io/badge/UI_State-Alpine.js_3.x-8BC0D0?style=for-the-badge&logo=alpine.js&logoColor=white" alt="Alpine.js"></a>
    <a href="https://tailwindcss.com/"><img src="https://img.shields.io/badge/Styling-Tailwind_CSS-38B2AC?style=for-the-badge&logo=tailwind-css&logoColor=white" alt="Tailwind CSS"></a>
    <a href="https://www.postgresql.org/"><img src="https://img.shields.io/badge/Database-PostgreSQL_16-4169E1?style=for-the-badge&logo=postgresql&logoColor=white" alt="PostgreSQL"></a>
    <a href="https://www.docker.com/"><img src="https://img.shields.io/badge/DevOps-Docker_Compose-2496ED?style=for-the-badge&logo=docker&logoColor=white" alt="Docker"></a>
  </p>

  <br />

</div>

---

## 🌟 Overview

> [!NOTE]
> **ticDesk** is an HTML-first IT Helpdesk system engineered to demonstrate modern, server-rendered web application architecture without heavy single-page app (SPA) complexity. Go renders dynamic HTML partials, HTMX handles zero-reload DOM updates, Alpine.js powers client UI states, and PostgreSQL 16 provides trigram search and analytical aggregates.

### ✨ Key Design & Architectural Highlights

- ⚡ **Zero-Reload Dynamic UI**: State-mutating actions (`PATCH` status, priority, assignment) return pre-rendered HTML partials swapped seamlessly by **HTMX v2** (`hx-swap="outerHTML"`).
- 🛡️ **Double-Barrier RBAC Enforcement**: Role permissions enforced at both HTTP middleware (`RequireRole`) and PostgreSQL repository query filters (`created_by = current_user_id`).
- 🔒 **Session Security & Bcrypt Hashing**: Password protection with `bcrypt` (cost 12), secure HTTP-only cookies via `scs`, and CSRF protection.
- 💬 **Staff Internal Notes**: Threaded ticket discussion featuring staff-only internal notes (`is_internal = true`) automatically excluded from customer accounts.
- 📊 **Real-time Analytics**: SQL aggregate queries for ticket statuses, priority breakdown, SLA resolution averages (`AVG(resolved_at - created_at)`), and agent workload queues.
- 📧 **Asynchronous Email Worker Queue**: Non-blocking in-process goroutine worker channel (`chan EmailJob`) sending HTML notification emails via SMTP (MailHog in dev).
- 🔍 **PostgreSQL Trigram Search**: `pg_trgm` fuzzy text matching on titles and descriptions with 300ms debounced HTMX search inputs.
- 🎨 **Apple/Linear-Grade Aesthetic**: Dark slate glassmorphism palette (`#090d16`), Inter typography, custom inline SVG icons, and responsive layouts.

---

## 🏗️ System Architecture

```mermaid
graph TD
    Client["💻 Web Browser<br/>(HTMX v2 + Alpine.js + Tailwind CSS)"]
    
    subgraph GoServer["🚀 Go Backend Server (net/http + chi)"]
        Router["Chi Router & RBAC Middleware"]
        Handlers["HTTP Handlers<br/>(Auth, Tickets, Comments, Admin)"]
        EmailWorker["📬 Async Email Worker Queue<br/>(chan EmailJob)"]
        StorageService["📁 Local Storage Service<br/>(10MB File Validation)"]
    end
    
    Database[("🐘 PostgreSQL 16<br/>(Tickets, Users, History, pg_trgm)")]
    MailHog["✉️ MailHog SMTP Server<br/>(Dev Email Testing)"]

    Client <-->|"HTTP / HTML Partial Swaps"| Router
    Router --> Handlers
    Handlers <-->|"pgx Pool Queries"| Database
    Handlers --> StorageService
    Handlers --> EmailWorker
    EmailWorker -->|"SMTP (Port 1025)"| MailHog
```

---

## 🔐 Role-Based Access Control (RBAC) Matrix

| Action | 👑 Admin | 🛠️ Support | 👤 Customer |
|:---|:---:|:---:|:---:|
| **Create Support Ticket** | ✅ | ✅ | ✅ |
| **View Own Tickets** | ✅ | ✅ | ✅ |
| **View All Org Tickets** | ✅ | ✅ | ❌ |
| **Inline Status & Priority Swaps** | ✅ | ✅ | ❌ |
| **Assign / Reassign Agents** | ✅ | ✅ | ❌ |
| **Post Public Comments** | ✅ | ✅ | ✅ |
| **Post Staff Internal Notes** | ✅ | ✅ | ❌ *(Hidden)* |
| **Upload File Attachments (10MB)** | ✅ | ✅ | ✅ |
| **User & Role Management (`/admin/users`)** | ✅ | ❌ | ❌ |
| **View Org Analytics Dashboard** | ✅ | ✅ | ❌ |

---

## 🚀 Quick Start Guide

> [!TIP]
> The easiest way to get ticDesk up and running locally is using **Docker Compose**.

### 1. Clone the Repository
```bash
git clone https://github.com/Suvesh108/ticdesk.git
cd ticdesk
```

### 2. Launch with Docker Compose
```bash
docker compose up -d --build
```

### 3. Service Endpoints

| Service | Access URL | Description |
|:---|:---|:---|
| **ticDesk Web App** | [`http://localhost:8081`](http://localhost:8081) | Main IT Helpdesk & Ticketing portal |
| **MailHog Web UI** | [`http://localhost:8025`](http://localhost:8025) | Inbox for capturing outgoing SMTP emails |
| **PostgreSQL Database** | `localhost:5432` | DB: `ticdesk` \| User: `ticdesk` \| Pass: `ticdesk_secret` |

---

## 🔑 Pre-seeded Test Credentials

> [!IMPORTANT]
> The application includes the following pre-seeded test accounts for Admin, Support, and Customer roles:

| Name | Email Address | Password | Role | Description |
|:---|:---|:---:|:---:|:---|
| **Suvesh** | **`admin@ticdesk.com`** | **`password123`** | `Admin` | System Administrator with full access |
| **Alex Rivera** | **`alex.support@ticdesk.com`** | **`password123`** | `Support` | Customer Support Agent 1 |
| **Sarah Chen** | **`sarah.support@ticdesk.com`** | **`password123`** | `Support` | Customer Support Agent 2 |
| **Test Customer** | **`cust@ticdesk.com`** | **`password123`** | `Customer` | End User / Customer Account |

*You can also register additional accounts on the `/register` page.*

---

## 🛠️ API & Web Route Sitemap

<details>
<summary><strong>👉 Click to expand full Route Table</strong></summary>
<br />

| Method | Route | Description | Target / Partial Response | Auth Level |
|:---:|:---|:---|:---|:---:|
| `GET` | `/login` | Render login form | `login.html` | Public |
| `POST` | `/login` | Authenticate & set session cookie | `303` Redirect $\rightarrow$ `/dashboard` | Public |
| `POST` | `/logout` | Destroy session & cookie | `303` Redirect $\rightarrow$ `/login` | Authenticated |
| `GET` | `/dashboard` | Render role-aware dashboard | `dashboard.html` | Authenticated |
| `GET` | `/dashboard/stats.json` | JSON aggregate stats API | `application/json` | Authenticated |
| `GET` | `/tickets` | Filterable ticket list | `ticket_list.html` | Authenticated |
| `GET` | `/tickets/new` | Ticket submission page | `ticket_new.html` | Authenticated |
| `POST` | `/tickets` | Create new ticket | `303` Redirect $\rightarrow$ `/tickets/{id}` | Authenticated |
| `GET` | `/tickets/{id}` | Ticket detail & comment thread | `ticket_detail.html` | Authenticated |
| `PATCH` | `/tickets/{id}/status` | Inline status badge swap | `ticket_status_badge.html` | Support / Admin |
| `PATCH` | `/tickets/{id}/priority` | Inline priority badge swap | `ticket_priority_badge.html` | Support / Admin |
| `PATCH` | `/tickets/{id}/assign` | Inline agent assignment swap | `ticket_assignee.html` | Support / Admin |
| `GET` | `/tickets/{id}/comments` | Render comment thread partial | `comment_list.html` | Authenticated |
| `POST` | `/tickets/{id}/comments` | Post reply or internal note | `comment_list.html` | Authenticated |
| `GET` | `/attachments/{id}` | Secure file download handler | Serve File Payload | Authenticated |
| `GET` | `/admin/users` | User management console | `admin_users.html` | Admin Only |
| `POST` | `/admin/users/{id}/role` | Update user role | `303` Redirect $\rightarrow$ `/admin/users` | Admin Only |
| `POST` | `/admin/users/{id}/deactivate` | Toggle account active status | `303` Redirect $\rightarrow$ `/admin/users` | Admin Only |

</details>

---

## 📂 Folder Structure

```
ticDesk/
├── cmd/
│   └── server/
│       └── main.go                 # Server entrypoint & dependency wiring
├── internal/
│   ├── auth/                       # Session management, bcrypt & middleware
│   ├── config/                     # Environment configuration loader
│   ├── db/
│   │   └── migrations/             # SQL Migrations (0001 to 0004)
│   ├── handlers/                   # Auth, Ticket, Comment, Admin, Dashboard handlers
│   ├── models/                     # Go models & data structures
│   ├── repository/                 # PostgreSQL query layer (pgx pool)
│   ├── services/                   # Storage abstraction & Async Email Worker
│   └── router/                     # Chi router setup & RBAC groups
├── web/
│   ├── templates/
│   │   ├── layouts/                # Base HTML layout & Tailwind styling
│   │   ├── pages/                  # Full-page templates
│   │   └── partials/               # HTMX zero-reload target partials
│   └── static/
│       └── uploads/                # Local attachment storage
├── docker-compose.yml              # Go App + PostgreSQL 16 + MailHog
├── Dockerfile                      # Production multi-stage Docker build
├── Makefile                        # Dev automation commands
└── README.md
```

---

<div align="center">

  <sub>Built with ❤️ by Suvesh • Powered by Go, HTMX v2, Alpine.js, Tailwind CSS & PostgreSQL</sub>

</div>
