# FPMB — Build & Architecture Reference

This document is the authoritative reference for agentic coding sessions. Read it fully before making changes.

---

## Quick-start commands

```bash
# Frontend dev server
bun run dev

# Frontend production build (source of truth for TS errors)
bun run build

# Backend run (from server/)
go run ./cmd/api/main.go

# Backend compile check (source of truth for Go errors)
go build ./...
```

- Go binary: `/usr/bin/go` (1.22.2)
- Package manager: `bun` (not npm/pnpm/yarn)
- Go module path: `github.com/fpmb/server`
- Backend must be run from `server/` — relative paths like `../data/` resolve from there

---

## Repository layout

```
openboard/
├── src/                          SvelteKit frontend (static SPA adapter)
│   ├── lib/
│   │   ├── api/
│   │   │   ├── client.ts         apiFetch, apiFetchFormData, token management
│   │   │   └── index.ts          all typed API methods grouped by resource
│   │   ├── components/
│   │   │   ├── Markdown/         Markdown.svelte — renders marked + DOMPurify
│   │   │   └── Modal/            Modal.svelte — generic overlay
│   │   ├── stores/
│   │   │   └── auth.svelte.ts    authStore singleton (init, login, register, logout, setUser)
│   │   ├── types/
│   │   │   └── api.ts            TypeScript interfaces mirroring Go models
│   │   └── utils/
│   │       └── fileRefs.ts       resolveFileRefs() — converts $file:<name> refs in Markdown
│   └── routes/
│       ├── (auth)/               login, register pages (no auth guard)
│       └── (app)/
│           ├── +layout.svelte    auth guard, top navbar (avatar, logout button)
│           ├── +page.svelte      dashboard
│           ├── board/[id]/       Kanban board
│           ├── calendar/         month/week calendar
│           ├── notifications/    notification inbox
│           ├── projects/         project list + project settings
│           ├── settings/user/    user profile + avatar upload + password change
│           ├── team/[id]/        team overview + team settings (avatar/banner upload)
│           └── whiteboard/[id]/  canvas whiteboard
├── server/
│   ├── cmd/api/main.go           Fiber app bootstrap, all route registrations
│   └── internal/
│       ├── database/db.go        MongoDB connection, GetCollection helper
│       ├── handlers/             one file per resource group (auth, teams, projects, …)
│       ├── middleware/auth.go    JWT Protected() middleware
│       ├── models/models.go      all MongoDB document structs (source of truth for field names)
│       ├── routes/               (unused legacy dir — ignore)
│       └── utils/                shared Go utilities
├── static/                       fonts, favicon
├── data/                         runtime file storage (gitignored)
│   ├── teams/<teamID>/avatar.<ext>
│   ├── teams/<teamID>/banner.<ext>
│   ├── users/<userID>/avatar.<ext>
│   └── projects/<projectID>/files/…
├── build/                        SvelteKit production output (served by Go)
├── package.json
├── svelte.config.js
├── vite.config.ts
└── tsconfig.json
```

---

## Backend (Go + GoFiber v2)

### Route registration — `server/cmd/api/main.go`

All routes are registered here. Add new routes to the appropriate group. Every group except `/auth` is wrapped in `middleware.Protected()`.

```
/api/health                       GET   — liveness check
/api/auth/register                POST
/api/auth/login                   POST
/api/auth/refresh                 POST
/api/auth/logout                  POST  (Protected)

/api/users/me                     GET PUT
/api/users/me/password            PUT
/api/users/me/avatar              POST GET   (multipart upload / serve)
/api/users/me/files               GET
/api/users/me/files/folder        POST
/api/users/me/files/upload        POST
/api/users/search                 GET  ?q=

/api/teams                        GET POST
/api/teams/:teamId                GET PUT DELETE
/api/teams/:teamId/members        GET
/api/teams/:teamId/members/invite POST
/api/teams/:teamId/members/:userId PUT DELETE
/api/teams/:teamId/projects       GET POST
/api/teams/:teamId/events         GET POST
/api/teams/:teamId/docs           GET POST
/api/teams/:teamId/files          GET
/api/teams/:teamId/files/folder   POST
/api/teams/:teamId/files/upload   POST
/api/teams/:teamId/avatar         POST GET
/api/teams/:teamId/banner         POST GET

/api/projects                     GET POST
/api/projects/:projectId          GET PUT DELETE
/api/projects/:projectId/archive  PUT
/api/projects/:projectId/members  GET POST
/api/projects/:projectId/members/:userId PUT DELETE
/api/projects/:projectId/board    GET
/api/projects/:projectId/columns  POST
/api/projects/:projectId/columns/:columnId PUT DELETE
/api/projects/:projectId/columns/:columnId/position PUT
/api/projects/:projectId/columns/:columnId/cards    POST
/api/projects/:projectId/events   GET POST
/api/projects/:projectId/files    GET
/api/projects/:projectId/files/folder POST
/api/projects/:projectId/files/upload POST
/api/projects/:projectId/webhooks GET POST
/api/projects/:projectId/whiteboard GET PUT

/api/cards/:cardId                PUT DELETE
/api/cards/:cardId/move           PUT

/api/events/:eventId              PUT DELETE

/api/notifications                GET
/api/notifications/read-all       PUT
/api/notifications/:notifId/read  PUT
/api/notifications/:notifId       DELETE

/api/docs/:docId                  GET PUT DELETE

/api/files/:fileId/download       GET
/api/files/:fileId                DELETE

/api/webhooks/:webhookId          PUT DELETE
/api/webhooks/:webhookId/toggle   PUT
```

### MongoDB models — `server/internal/models/models.go`

This file is the single source of truth for all field names and types. Always check here before referencing a field.

| Struct | Collection | Key fields |
|---|---|---|
| `User` | `users` | `_id`, `name`, `email`, `password_hash` (json:`-`), `avatar_url` |
| `Team` | `teams` | `_id`, `name`, `workspace_id`, `avatar_url`, `banner_url`, `created_by` |
| `TeamMember` | `team_members` | `team_id`, `user_id`, `role_flags`, `invited_by` |
| `Project` | `projects` | `_id`, `team_id`, `name`, `description`, `visibility`, `is_public`, `is_archived`, `created_by` |
| `ProjectMember` | `project_members` | `project_id`, `user_id`, `role_flags` |
| `BoardColumn` | `columns` | `project_id`, `title`, `position` |
| `Card` | `cards` | `column_id`, `project_id`, `title`, `description`, `priority`, `color`, `due_date`, `assignees []string`, `subtasks []Subtask`, `position` |
| `Subtask` | (embedded) | `id int`, `text`, `done` |
| `Event` | `events` | `title`, `date`, `time`, `color`, `description`, `scope`, `scope_id` |
| `Notification` | `notifications` | `user_id`, `type`, `message`, `project_id`, `card_id`, `read` |
| `Doc` | `docs` | `team_id`, `title`, `content`, `created_by` |
| `File` | `files` | `project_id`, `team_id`, `user_id`, `name`, `type`, `size_bytes`, `parent_id`, `storage_url` |
| `Webhook` | `webhooks` | `project_id`, `name`, `type`, `url`, `secret_hash` (json:`-`), `status`, `last_triggered` |
| `Whiteboard` | `whiteboards` | `project_id`, `data` |

### RBAC

Roles are **hierarchical integers**, not bitflags. Use `>=` comparisons.

```
Viewer  = 1
Editor  = 2
Admin   = 4
Owner   = 8
```

Example: `member.RoleFlags >= 2` means Editor or above.

### Image upload/serve pattern (handler)

Reference implementation: `server/internal/handlers/teams.go` (UploadTeamAvatar / ServeTeamAvatar).

```go
// Validate extension
allowedImageExts = map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true}

// Build storage path
dir := fmt.Sprintf("../data/<resource>/%s", id.Hex())
os.MkdirAll(dir, 0755)

// Delete old file (glob by base name, any extension)
old, _ := filepath.Glob(filepath.Join(dir, "avatar.*"))
for _, f := range old { os.Remove(f) }

// Save new file
c.SaveFile(fh, filepath.Join(dir, "avatar"+ext))

// Update DB with static URL string
database.GetCollection("<coll>").UpdateOne(ctx, bson.M{"_id": id},
    bson.M{"$set": bson.M{"avatar_url": "/api/<resource>/avatar", "updated_at": time.Now()}})

// Serve
matches, _ := filepath.Glob(filepath.Join(dir, "avatar.*"))
return c.SendFile(matches[0])
```

Key points:
- `allowedImageExts` is declared once in `teams.go` and reused across the `handlers` package (same package, no redeclaration needed in `users.go`)
- Avatar URL stored in DB is a **static string** (e.g. `/api/users/me/avatar`), not per-extension — the serve endpoint globs at runtime
- Max upload size: GoFiber default (4MB). Adjust `fiber.Config{BodyLimit}` in `main.go` if needed

### Background jobs

`startDueDateReminder()` in `main.go` runs every hour and creates `due_soon` notifications for cards due within 1 or 3 days. It deduplicates within a 24-hour window.

---

## Frontend (SvelteKit + Svelte 5 Runes)

### Rules — non-negotiable

- **Svelte 5 Runes only**: `$state`, `$derived`, `$props`, `$bindable`. `onMount` is also acceptable.
- **No legacy Svelte**: no `export let`, no reactive `$:`, no `writable` stores.
- `$state` may only be used inside `.svelte`, `.svelte.ts`, or `.svelte.js` files — never in plain `.ts` modules.
- No code comments anywhere.
- No extra files or folders.
- `$page.params.id` must always be written as `$page.params.id ?? ''`.

### LSP / type errors

The LSP cache is often stale. **Do not** treat LSP errors as real until confirmed by `bun run build`. Pre-existing LSP errors on `$page`, `Project.visibility`, `Project.is_public`, etc. are known stale artefacts.

### API client — `src/lib/api/client.ts`

| Export | Purpose |
|---|---|
| `apiFetch<T>(path, options?)` | JSON fetch, auto-refreshes token on 401, throws on error |
| `apiFetchFormData<T>(path, formData)` | Multipart POST (method hardcoded to POST), same token/retry logic |
| `getAccessToken()` | Returns current in-memory access token |
| `setAccessToken(token)` | Updates in-memory token and localStorage |

Tokens:
- `access_token` — localStorage + in-memory, 15-minute JWT
- `refresh_token` — localStorage only, 7-day JWT
- `user_id` — localStorage only (used for convenience reads)

### API methods — `src/lib/api/index.ts`

All API calls are grouped by resource: `auth`, `users`, `teams`, `projects`, `board`, `cards`, `events`, `notifications`, `docs`, `files`, `webhooks`. Import the group you need:

```ts
import { users, teams, projects, board, cards } from '$lib/api';
```

### Auth store — `src/lib/stores/auth.svelte.ts`

```ts
authStore.user        // User | null (reactive)
authStore.loading     // boolean (reactive)
authStore.init()      // call once in root layout onMount
authStore.login(email, password)
authStore.register(name, email, password)
authStore.logout()    // calls API, clears tokens, nulls user
authStore.setUser(u)  // update user after profile/avatar changes
```

### TypeScript interfaces — `src/lib/types/api.ts`

Mirrors Go models exactly. Always import from here; never inline interfaces.

Key interfaces: `User`, `Team`, `TeamMember`, `Project`, `ProjectMember`, `Card`, `Subtask`, `Column`, `BoardData`, `Event`, `Notification`, `Doc`, `FileItem`, `Webhook`, `Whiteboard`, `AuthResponse`.

### File ref resolution — `src/lib/utils/fileRefs.ts`

`resolveFileRefs(text, files)` replaces `$file:<name>` tokens in Markdown with links. Unmatched refs render as `` `unknown file: <name>` ``.

### Tailwind notes

- Use Tailwind v4 utility classes only.
- `ring-neutral-750` does not exist — use `ring-neutral-800`.
- Never combine `inline-block` with `flex` on the same element — `inline-block` wins and kills flex behaviour. Use `flex` alone.

---

## Common patterns

### Adding a new API route (end-to-end checklist)

1. Add handler function in the appropriate `server/internal/handlers/*.go` file.
2. Register the route in `server/cmd/api/main.go` under the right group.
3. Add the typed method to `src/lib/api/index.ts`.
4. If a new response shape is needed, add an interface to `src/lib/types/api.ts`.
5. Call from a `.svelte` page; use `apiFetchFormData` for multipart, `apiFetch` for everything else.
6. Verify: `go build ./...` from `server/` and `bun run build` from root.

### Archived project read-only enforcement

`src/routes/(app)/board/[id]/+page.svelte`:
- `isArchived` state derived from `project.is_archived` on mount.
- Yellow banner shown when archived.
- Cards: `draggable={!isArchived}`, click/keydown handlers suppressed, edit button hidden.
- Add Card and Add Column buttons hidden when archived.
- `handleDrop` returns early when `isArchived`.

### User / team avatar upload (frontend)

```ts
const updated = await users.uploadAvatar(file);   // returns updated User
authStore.setUser(updated);

const updatedTeam = await teams.uploadAvatar(teamId, file);  // returns updated Team
```

Display: show `<img src={user.avatar_url}>` if `avatar_url` is set, otherwise show the user's initial letter in a coloured circle.

---

## Environment variables

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8080` | Server listen port |
| `MONGO_URI` | `mongodb://localhost:27017` | MongoDB connection string |
| `MONGO_DB_NAME` | `fpmb` | MongoDB database name |
| `JWT_SECRET` | — | Required in production |

`.env` is loaded automatically via `godotenv` in `main.go`. System env vars are used as fallback.

---

## Data directory layout

Runtime files are stored under `data/` relative to the repository root. The Go server runs from `server/`, so it accesses `../data/`.

```
data/
├── teams/<teamID>/
│   ├── avatar.<ext>
│   └── banner.<ext>
├── users/<userID>/
│   └── avatar.<ext>
└── projects/<projectID>/files/…
```

`data/` is gitignored.

Goal
Build out and improve the FPMB (Free Project Management Boards) self-hosted project management app at /home/coder/openboard. The current session focused on: user avatar upload, logout button, archived project read-only mode, task card UI fixes, and $file: unknown file handling. The final task (still in progress) is updating README.md and creating BUILD.md with full coding/architecture details for future agent sessions.
---
Instructions
- Project is at /home/coder/openboard
- No code comments
- No extra files or folders
- Frontend: Svelte 5 Runes only ($state, $derived, $props, $bindable) — onMount is acceptable; no legacy Svelte patterns
- Backend: Go, module path github.com/fpmb/server
- Use bun (not npm): bun run build, bun run dev
- Go is at /usr/bin/go (1.22.2); run from server/ via go run ./cmd/api/main.go
- LSP has a stale cache — use bun run build / go build ./... as the source of truth for real errors; ignore LSP errors that were pre-existing
- $page.params.id must use ?? ''
- RBAC: Viewer=1, Editor=2, Admin=4, Owner=8 — roles are hierarchical (>= not bitwise)
- Server runs from server/ so ../data/ resolves to /home/coder/openboard/data/
- File storage: ../data/teams/<teamID>/avatar.<ext>, ../data/users/<userID>/avatar.<ext>
- Use apiFetchFormData in frontend for multipart uploads
- $state may only be used inside .svelte or .svelte.ts/.svelte.js files, not plain .ts modules
---
Discoveries
Architecture
- allowedImageExts map is defined in server/internal/handlers/teams.go — shared across the handlers package, so users.go reuses it directly (same package, no redeclaration)
- authStore (src/lib/stores/auth.svelte.ts) already has a logout() method that calls the API, clears tokens, and nulls the user — no need to duplicate logic
- apiFetchFormData always uses POST method (hardcoded in client.ts)
- LSP errors on board page ($page, Project.visibility, Project.is_public, etc.) are all pre-existing stale cache issues — builds pass cleanly
- ring-neutral-750 is not a valid Tailwind class (causes white ring fallback); use ring-neutral-800 for card backgrounds
- inline-block + flex conflict — inline-block wins and kills flex centering; use flex alone
- $file:<name> refs are resolved in src/lib/utils/fileRefs.ts via resolveFileRefs() — unmatched refs previously returned raw  `$file:name`  syntax
- Board page is_archived enforcement: drag-drop guarded in handleDrop, template uses {#if !isArchived} to hide Add Card / Add Column, draggable={!isArchived} on cards
Key patterns
- Team/user image upload: validate ext with allowedImageExts, glob-delete old file, SaveFile, UpdateOne with URL, return updated document
- Serve image: glob-find by extension, SendFile
- Avatar URL stored as /api/users/me/avatar (static string, not per-extension) — the serve endpoint finds the file at runtime
---
Accomplished
✅ Completed this session
1. User avatar upload — Full end-to-end:
   - server/internal/handlers/users.go: Added UploadUserAvatar and ServeUserAvatar handlers
   - server/cmd/api/main.go: Registered POST /users/me/avatar and GET /users/me/avatar
   - src/lib/api/index.ts: Added users.uploadAvatar(file)
   - src/routes/(app)/settings/user/+page.svelte: Added avatar display (<img> if set, else initial letter), file input upload button, success/error feedback, calls authStore.setUser(updated)
2. Logout button — Added to top navbar in src/routes/(app)/+layout.svelte:
   - Arrow-out icon button to the right of the user avatar (desktop only, hidden md:flex)
   - Calls authStore.logout() then goto('/login')
3. Archived project read-only mode — src/routes/(app)/board/[id]/+page.svelte:
   - Added isArchived state, populated from project.is_archived on mount
   - Yellow archived banner shown when isArchived is true
   - Cards: draggable={!isArchived}, click/keydown suppressed, edit ... button hidden
   - "Add Card" button hidden per column when archived
   - "Add Column" button hidden when archived
   - handleDrop returns early when isArchived
4. Task card assignee ring fix — ring-neutral-750 → ring-neutral-800, inline-block removed (was conflicting with flex centering)
5. $file: unknown file fallback — src/lib/utils/fileRefs.ts: unmatched refs now render as  `unknown file: <name>` 
🔄 In Progress — Documentation update
- README.md — needs updating to reflect: user avatar upload, team avatar/banner, logout button, archived read-only mode, new API routes
- BUILD.md — needs to be created as a comprehensive reference for future agentic coding sessions, covering all architecture, conventions, patterns, file layout, API routes, models, and coding rules
---
Relevant files / directories
Backend (Go)
server/
├── cmd/api/main.go                              routes registration (recently: user avatar routes added)
└── internal/
    ├── handlers/
    │   ├── auth.go
    │   ├── teams.go                             reference: allowedImageExts, uploadTeamImage/serveTeamImage pattern
    │   ├── users.go                             recently added: UploadUserAvatar, ServeUserAvatar
    │   ├── projects.go
    │   ├── board.go
    │   ├── cards.go
    │   ├── files.go
    │   ├── notifications.go
    │   ├── docs.go
    │   ├── events.go
    │   ├── webhooks.go
    │   └── whiteboard.go
    ├── middleware/
    │   └── auth.go
    ├── models/
    │   └── models.go                            source of truth for all field names/types
    └── database/
        └── db.go
Frontend
src/
├── lib/
│   ├── api/
│   │   ├── client.ts                            apiFetch, apiFetchFormData, token management
│   │   └── index.ts                             all API methods (recently: users.uploadAvatar added)
│   ├── components/
│   │   ├── Markdown/Markdown.svelte
│   │   └── Modal/Modal.svelte
│   ├── stores/
│   │   └── auth.svelte.ts                       authStore: init, login, register, logout, setUser
│   ├── types/
│   │   └── api.ts                               all TypeScript interfaces
│   └── utils/
│       └── fileRefs.ts                          resolveFileRefs (recently: unknown file fallback fixed)
└── routes/
    └── (app)/
        ├── +layout.svelte                       navbar, auth guard (recently: logout button added)
        ├── board/[id]/+page.svelte              kanban board (recently: archived read-only mode)
        ├── settings/user/+page.svelte           user settings (recently: avatar upload UI)
        ├── team/[id]/
        │   ├── +page.svelte                     team overview (shows avatar/banner)
        │   └── settings/+page.svelte            team settings (avatar/banner upload)
        └── projects/
            ├── +page.svelte                     project list (shows is_archived badge)
            └── [id]/settings/+page.svelte       project settings (archive/delete)
Docs (in progress)
/home/coder/openboard/README.md                  needs update
/home/coder/openboard/BUILD.md                   needs to be created
Data directory (runtime)
/home/coder/openboard/data/
├── teams/<teamID>/avatar.<ext>
├── teams/<teamID>/banner.<ext>
└── users/<userID>/avatar.<ext>