# Deployment (Render)

This project deploys to [Render](https://render.com) using a Render Blueprint defined in [`render.yaml`](../render.yaml). No container registry, image push or deploy hook is used; Render builds from the repository Dockerfiles.

## One-time setup: apply the Blueprint

1. Make sure [`render.yaml`](../render.yaml) is on `main` before applying the Blueprint: merge the change into `dev`, then merge `dev` into `main`, then apply the Blueprint.
2. In the [Render Dashboard](https://dashboard.render.com), go to **Blueprints → New Blueprint Instance**.
3. Connect the repository and select the branch `main`.
4. Render will detect [`render.yaml`](../render.yaml) and create the services and database defined by the Blueprint.
5. Set the required secret values in the Render dashboard for any variables marked `sync: false` (see below). Render will not store secret values in the repository.
6. Trigger the first deploy after the secrets are set.

## Services and database

The Blueprint defines:
- One backend web service (Docker, build context `backend/`, health check at `/health`), on the `free` plan.
- One frontend web service (Docker, build context `frontend/`), on the `free` plan; public origin that proxies `/api/` to the backend at its public URL.
- One Render Postgres database, on the `free` plan. The backend reads `DATABASE_URL` from it.

All resources use Render's free plans. A free Render Postgres expires 30 days after creation, has 1 GB of storage, and each workspace can have only one.

## Networking between services

Free Render web services cannot receive private network traffic, so the frontend reaches the backend over the public internet: `API_UPSTREAM` is the backend's public URL, `https://learn-ai-engineering-backend.onrender.com`. The Caddy server in the frontend image proxies `/api/` to that URL and rewrites the `Host` header to the upstream host, so the backend accepts the request. The browser stays same-origin and no CORS is needed.

A free backend web service spins down after 15 minutes without traffic. The first request after that starts it again and is slow (tens of seconds), so a slow first API call after an idle period is expected.

## Environment variables

### Backend (web service)
| Variable | Source | Notes |
| --- | --- | --- |
| `DATABASE_URL` | Render Postgres (from database) | Provided by Render from the linked database. |
| `ENV` | Blueprint | Set to `production`. |
| `HOST` | Blueprint | Set to `0.0.0.0`. |
| `PORT` | Blueprint | Set to `8080`. |
| `SEED_MAINTAINER_EMAIL` | Render dashboard (secret) | Marked `sync: false`. No value stored in the repository. |
| `SEED_MAINTAINER_PASSWORD` | Render dashboard (secret) | Marked `sync: false`. No value stored in the repository. |

### Frontend (web service)
| Variable | Source | Notes |
| --- | --- | --- |
| `API_UPSTREAM` | Blueprint | Backend public URL, `https://learn-ai-engineering-backend.onrender.com`. The frontend's Caddy proxies `/api/` to it. |
| `VITE_API_MODE` | Blueprint | Set to `http`. Build-time value for the frontend bundle. |
| `VITE_API_BASE_URL` | Blueprint | Set to `/api`. Build-time value for the frontend bundle. |
| `PORT` | Blueprint | Set to `80`. Matches the port exposed by the frontend image. |

### Database
| Variable | Source | Notes |
| --- | --- | --- |
| Render Postgres connection | Render | Exposed to the backend via `DATABASE_URL`. Credentials live only in Render environment variables. |

## Promotion

The maintainer merges `dev` into `main` by hand. With `autoDeployTrigger: checksPass`, Render will deploy from `main` when checks pass.

## Rollback

To roll back, redeploy an earlier commit from the Render dashboard (or revert and merge to `main`). No custom deploy hook is used.

## Notes

- The deploy gate and first verification are handled separately (T34b).
- No registry, no image push and no deploy hook are used.
- Secrets live only in Render environment variables.
