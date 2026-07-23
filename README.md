<div align="center">

  <h1>⚡ Outlook 365 — ticDesk Workstation</h1>
  <p><strong>Enterprise IT Helpdesk, Schedule Calendar & Shift Notes Platform</strong></p>

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
> **ticDesk Outlook 365 Workstation** is a complete, from-scratch redesign modeled directly after **Microsoft Outlook Web (Outlook 365)**. It features an Outlook 365 navigation app rail, a secondary folder pane, an interactive **Schedule Calendar (`/calendar`)**, and a **Shift Notes Scratchpad (`/notes`)** designed for Admins and Support agents to manage maintenance windows, agent shifts, SLA targets, and handoff notes.

### ✨ Microsoft Outlook 365 Features

- 📅 **Outlook Schedule Calendar (`/calendar`)**: Schedule and track server maintenance windows, support shifts, and SLA deadline targets.
- 📝 **Outlook Shift Notes & Scratchpad (`/notes`)**: Sticky notes for support agents & admins to store quick reference IP addresses, credential snippets, and shift handoff checklists with pin/unpin controls.
- ✉️ **Outlook Mail Tickets Inbox (`/tickets`)**: Server-rendered ticket inbox with HTMX v2 zero-reload status, priority, and assignee badge swaps.
- 🛡️ **Double-Barrier RBAC Security**: Role permissions enforced at both HTTP middleware (`RequireRole`) and PostgreSQL repository query filters.
- 📊 **Outlook Today Analytics (`/dashboard`)**: Real-time aggregate metrics for open tickets, in-progress issues, SLA resolution averages, and agent workload.
- 📧 **Outlook HTML Email Notifications**: In-process goroutine worker sending Microsoft Outlook-styled HTML emails via SMTP (MailHog).

---

## 🏗️ System Architecture

```mermaid
graph TD
    Client["💻 Web Browser<br/>(Microsoft Outlook 365 Layout + HTMX v2)"]
    
    subgraph GoServer["🚀 Go Backend Server (net/http + chi)"]
        Router["Chi Router & RBAC Middleware"]
        Handlers["HTTP Handlers<br/>(Tickets, Calendar, Notes, Admin)"]
        EmailWorker["📬 Async Email Worker Queue<br/>(chan EmailJob)"]
        StorageService["📁 Local Storage Service<br/>(10MB File Validation)"]
    end
    
    Database[("🐘 PostgreSQL 16<br/>(Tickets, Events, Notes, Users)")]
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
| **Create & View Support Tickets** | ✅ | ✅ | ✅ |
| **Manage Outlook Calendar Schedule** | ✅ | ✅ | ❌ |
| **Manage Shift Notes & Scratchpad** | ✅ | ✅ | ❌ |
| **Inline Status & Priority Swaps** | ✅ | ✅ | ❌ |
| **Assign / Reassign Agents** | ✅ | ✅ | ❌ |
| **Post Staff Internal Notes** | ✅ | ✅ | ❌ *(Hidden)* |
| **User & Role Management (`/admin/users`)** | ✅ | ❌ | ❌ |

---

## 🔑 Pre-seeded Test Credentials

| Name | Email Address | Password | Role | Description |
|:---|:---|:---:|:---:|:---|
| **Suvesh** | **`admin@ticdesk.com`** | **`password123`** | `Admin` | System Administrator (Full Access) |
| **Alex Rivera** | **`alex.support@ticdesk.com`** | **`password123`** | `Support` | Customer Support Agent 1 |
| **Sarah Chen** | **`sarah.support@ticdesk.com`** | **`password123`** | `Support` | Customer Support Agent 2 |
| **Test Customer** | **`cust@ticdesk.com`** | **`password123`** | `Customer` | End User / Customer Account |

---

## 🛠️ API & Web Route Sitemap

| Method | Route | Description | Target Page |
|:---:|:---|:---|:---:|
| `GET` | `/dashboard` | Today Overview & Workstation | `dashboard.html` |
| `GET` | `/tickets` | Outlook Mail Ticket Inbox | `ticket_list.html` |
| `GET` | `/calendar` | Outlook Schedule & Calendar | `calendar.html` |
| `POST` | `/calendar/events` | Create Maintenance / Shift Event | `calendar.html` |
| `GET` | `/notes` | Outlook Shift Notes Scratchpad | `notes.html` |
| `POST` | `/notes` | Create Sticky Note | `notes.html` |
| `POST` | `/notes/{id}/pin` | Pin/Unpin Sticky Note | `notes.html` |
| `GET` | `/admin/users` | User & Role Management | `admin_users.html` |

---

<div align="center">

  <sub>Built with ❤️ by Suvesh • Microsoft Outlook 365 Workstation Engine</sub>

</div>
