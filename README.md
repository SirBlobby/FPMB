# FPMB — Free Project Management Boards

FPMB is a self-hosted, open-source project management platform. It provides Kanban boards, task tracking, team collaboration, a canvas whiteboard, a knowledge base, a calendar, file management, webhook integrations, API key management, and role-based access control — all in one application.

## Features

- **Kanban Boards** — drag-and-drop cards with priority, color labels, due dates, assignees, subtasks, and Markdown descriptions
- **Personal & Team Projects** — create projects scoped to a team or privately for yourself
- **Whiteboard** — full-screen canvas with pen, rectangle, circle, and eraser tools; auto-saves after every stroke
- **Team Docs** — two-pane Markdown knowledge base editor per team
- **Calendar** — month and week views with per-team and per-project event creation
- **File Manager** — per-project, per-team, and personal file/folder browser with upload support
- **Webhooks** — integrations with Discord, GitHub, Gitea, Slack, and custom endpoints
- **Notifications** — inbox with unread indicators, badge count, and mark-as-read
- **API Keys** — personal API keys with granular scopes for programmatic access
- **API Documentation** — built-in interactive API reference page at `/api-docs`
- **RBAC** — hierarchical role flags (Viewer `1`, Editor `2`, Admin `4`, Owner `8`) for fine-grained permission control
- **User Settings** — profile management, avatar upload, password change, and API key management
- **Archived Projects** — projects can be archived; the board becomes read-only (no drag-drop, no card edits, no new cards or columns)
- **Docker Support** — single-command deployment with Docker Compose

## Tech Stack

| Layer | Technology |
|---|---|
| Frontend | SvelteKit + Svelte 5 (Runes), TypeScript |
| Styling | Tailwind CSS v4, JetBrains Mono |
| Icons | @iconify/svelte (Lucide + Simple Icons) |
| Markdown | marked + DOMPurify |
| Backend | Go 1.24 + GoFiber v2 |
| Database | MongoDB 7 |
| Auth | JWT (access 15 min, refresh 7 days) + personal API keys |
| Authorization | RBAC hierarchical role flags |
| Deployment | Docker multi-stage build + Docker Compose |

## Getting Started

### Prerequisites

- [Bun](https://bun.sh) 1.x (or Node.js 22+)
- Go 1.24+
- MongoDB 7+ (local or Atlas)

### Development

Install frontend dependencies and start the dev server:

```bash
bun install
bun run dev
```

In a separate terminal, start the backend:

```bash
cd server
cp example.env .env    # edit with your MongoDB URI and secrets
go run ./cmd/api/main.go
```

The frontend dev server runs on `http://localhost:5173` and proxies API requests.
The Go server runs on `http://localhost:8080` and serves both the API and the built frontend in production.

### Production Build (Manual)

```bash
bun run build
cd server && go build -o bin/fpmb ./cmd/api/main.go
./bin/fpmb
```

### Docker (Recommended)

The easiest way to deploy FPMB:

```bash
# Start everything (app + MongoDB)
docker compose up -d

# With custom secrets
JWT_SECRET=my-secret JWT_REFRESH_SECRET=my-refresh-secret docker compose up -d

# Rebuild after code changes
docker compose up -d --build

# View logs
docker compose logs -f app

# Stop
docker compose down
```

This starts:
- **fpmb** — the application on port `8080`
- **fpmb-mongo** — MongoDB 7 on port `27017`

Data is persisted in Docker volumes (`app_data` for uploads, `mongo_data` for the database).

### Environment Variables

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8080` | Server listen port |
| `MONGO_URI` | `mongodb://localhost:27017` | MongoDB connection string |
| `MONGO_DB_NAME` | `fpmb` | MongoDB database name |
| `JWT_SECRET` | `changeme-jwt-secret` | Secret for signing access tokens (**change in production**) |
| `JWT_REFRESH_SECRET` | `changeme-refresh-secret` | Secret for signing refresh tokens (**change in production**) |

## API Overview

All routes are under `/api`. Protected endpoints require a `Bearer` token (JWT access token or personal API key).

A full interactive reference is available in-app at `/api-docs`.

### Authentication

| Method | Route | Description |
|---|---|---|
| POST | `/auth/register` | Create a new account |
| POST | `/auth/login` | Login — returns access + refresh tokens |
| POST | `/auth/refresh` | Exchange refresh token for new tokens |
| POST | `/auth/logout` | Logout (requires auth) |

### API Keys

| Method | Route | Description |
|---|---|---|
| GET | `/users/me/api-keys` | List all active API keys |
| POST | `/users/me/api-keys` | Create a new key with scopes |
| DELETE | `/users/me/api-keys/:keyId` | Revoke an API key |

**Available scopes:** `read:projects`, `write:projects`, `read:boards`, `write:boards`, `read:teams`, `write:teams`, `read:files`, `write:files`, `read:notifications`

### Users

| Method | Route | Description |
|---|---|---|
| GET | `/users/me` | Get current user profile |
| PUT | `/users/me` | Update profile (name, email) |
| PUT | `/users/me/password` | Change password |
| POST | `/users/me/avatar` | Upload avatar (multipart) |
| GET | `/users/me/avatar` | Serve avatar image |
| GET | `/users/search?q=` | Search users by name/email |

### Teams

| Method | Route | Description |
|---|---|---|
| GET/POST | `/teams` | List or create teams |
| GET/PUT/DELETE | `/teams/:teamId` | Get, update, or delete a team |
| GET | `/teams/:teamId/members` | List team members |
| POST | `/teams/:teamId/members/invite` | Invite a member |
| PUT/DELETE | `/teams/:teamId/members/:userId` | Update role or remove member |
| GET/POST | `/teams/:teamId/projects` | List or create team projects |
| GET/POST | `/teams/:teamId/events` | List or create team events |
| GET/POST | `/teams/:teamId/docs` | List or create docs |
| GET | `/teams/:teamId/files` | List team files |
| POST | `/teams/:teamId/files/upload` | Upload file (multipart) |
| POST | `/teams/:teamId/avatar` | Upload team avatar |
| POST | `/teams/:teamId/banner` | Upload team banner |

### Projects

| Method | Route | Description |
|---|---|---|
| GET/POST | `/projects` | List all or create personal project |
| GET/PUT/DELETE | `/projects/:projectId` | Get, update, or delete project |
| PUT | `/projects/:projectId/archive` | Toggle archive state |
| GET/POST | `/projects/:projectId/members` | List or add members |
| GET | `/projects/:projectId/board` | Get board (columns + cards) |
| POST | `/projects/:projectId/columns` | Create a column |
| POST | `/projects/:projectId/columns/:columnId/cards` | Create a card |
| GET/POST | `/projects/:projectId/events` | List or create events |
| GET | `/projects/:projectId/files` | List project files |
| POST | `/projects/:projectId/files/upload` | Upload file (multipart) |
| GET/POST | `/projects/:projectId/webhooks` | List or create webhooks |
| GET/PUT | `/projects/:projectId/whiteboard` | Get or save whiteboard |

### Cards, Events, Files, Webhooks, Notifications

| Method | Route | Description |
|---|---|---|
| PUT/DELETE | `/cards/:cardId` | Update or delete a card |
| PUT | `/cards/:cardId/move` | Move card between columns |
| PUT/DELETE | `/events/:eventId` | Update or delete an event |
| GET | `/files/:fileId/download` | Download a file |
| DELETE | `/files/:fileId` | Delete a file |
| PUT/DELETE | `/webhooks/:webhookId` | Update or delete a webhook |
| PUT | `/webhooks/:webhookId/toggle` | Enable/disable a webhook |
| GET | `/notifications` | List notifications |
| PUT | `/notifications/read-all` | Mark all as read |
| PUT | `/notifications/:notifId/read` | Mark one as read |
| DELETE | `/notifications/:notifId` | Delete a notification |

## License

MIT
