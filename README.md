# GoAI

HTTP service: Gemini-backed **audit** APIs (TheTVDB series id hint, torrent season/episode).  
Consumers call this over HTTP with a shared Bearer secret; they do **not** need to import this module.

## Layout (hexagonal, alinhado a GoAnimes)

| Camada | Pacote |
|--------|--------|
| Composição | `internal/app` (`Wire`) |
| Casos de uso | `internal/core/services` |
| Portas | `internal/core/ports` (`TextCompletion`, `AuditService`) |
| Domínio | `internal/core/domain` |
| Driving adapter (HTTP) | `internal/adapters/http/ginapi` |
| Driven adapter (Gemini) | `internal/adapters/gemini` |
| Binário | `cmd/goai` |

## GitHub Actions

Padrão alinhado ao **GoAnimes**: `checkout@v6`, `setup-go@v6` (sem cache de módulos), `go vet`, **golangci-lint** `v2.11` (`.golangci.yml`), `go test -count=1`.

| Workflow | Quando | O quê |
|----------|--------|--------|
| **ci** | PR para `main` / push em branches que **não** sejam `main` ou `master` | vet + lint + test + build (em `main`/`master` não corre — evita duplicar com **oracle-deploy**) |
| **oracle-deploy** | push em `main`/`master` ou *Run workflow* | Igual ao **GoAnimes**: jobs **test** → **image** (GHCR `:main`, `:sha`) → **deploy** (SSH, `.env.goai.deploy`, recreate **goai**) |
| **release** | tag `v*` | vet + lint + test + build + artefacto `goai-linux-amd64` |

**Repository secrets** (SSH, iguais aos do GoTV): `OCI_VM_HOST`, `OCI_VM_USER`, `OCI_SSH_PRIVATE_KEY`, opcional `OCI_DEPLOY_ROOT`, opcional `GHCR_*`.

**Environment `prd` (repo GoAI):** `GOAI_ADMIN_API_KEY`, `GOAI_GEMINI_API_KEYS`; variable opcional `GOAI_GEMINI_MODEL`. O GoTV **não** grava `GOAI_*` — deploys são independentes.

## Environment

| Variable | Required | Description |
|----------|----------|-------------|
| `GOAI_ADMIN_API_KEY` | yes | Bearer token for `/v1/*` (igual ideia a `GOANIMES_ADMIN_API_KEY` / `GOTV_ADMIN_API_KEY`) |
| `GOAI_INTERNAL_API_KEY` | legacy | Alias opcional se `GOAI_ADMIN_API_KEY` estiver vazio |
| `GOAI_GEMINI_API_KEYS` | yes | Comma-separated Gemini API keys (rotation on 429) |
| `GOAI_GEMINI_MODEL` | no | Default `gemini-2.0-flash` |
| `GOAI_GEMINI_BASE_URL` | no | Default `https://generativelanguage.googleapis.com` |
| `GOAI_GEMINI_KEY_COOLDOWN_SEC` | no | Cooldown after quota error per key (default 60) |
| `GOAI_HTTP_ADDR` or `PORT` | no | Listen address (default `:8088` or `:$PORT`) |

## API

- `GET /healthz` — no auth
- `POST /v1/audit/series` — `Authorization: Bearer <GOAI_ADMIN_API_KEY>`

  Body (example):

  ```json
  {
    "series_name": "Show Title",
    "series_id": "optional-stremio-series-id",
    "mal_id": 0,
    "imdb_id": "tt1234567",
    "year": 2024,
    "title_preferred": "",
    "title_native": "",
    "existing_tvdb_series_id": 0
  }
  ```

  Response: `thetvdb_series_id`, `confidence`, `notes`, `raw_response_json`, `prompt_version`.

- `POST /v1/audit/release` — same auth

  Body (example):

  ```json
  {
    "torrent_title": "[Group] Show - 05 [1080p]",
    "series_name": "Show",
    "series_id": "",
    "current_season": 1,
    "current_episode": 5,
    "is_special": false
  }
  ```

  Response: `season`, `episode`, `is_special`, `confidence`, `notes`, `raw_response_json`, `prompt_version`.

## Run

```bash
export GOAI_ADMIN_API_KEY=devsecret
export GOAI_GEMINI_API_KEYS=your_key
go run ./cmd/goai
```
