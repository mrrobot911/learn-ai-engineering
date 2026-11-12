# Deployment (Render)

This project deploys to [Render](https://render.com) using a Render Blueprint defined in [`render.yaml`](../render.yaml). No container registry, image push or deploy hook is used; Render builds one image from the root [`Dockerfile`](../Dockerfile).

## One-time setup: apply the Blueprint

1. Make sure [`render.yaml`](../render.yaml) is on `main` before applying the Blueprint: merge the change into `dev`, then merge `dev` into `main`, then apply the Blueprint.
2. In the [Render Dashboard](https://dashboard.render.com), go to **Blueprints → New Blueprint Instance**.
3. Connect the repository and select the branch `main`.
4. Render will detect [`render.yaml`](../render.yaml) and create the web service and database defined by the Blueprint.
5. Set the required secret values in the Render dashboard for any variables marked `sync: false` (see below). Render will not store secret values in the repository.
6. Trigger the first deploy after the secrets are set.
7. If an earlier Blueprint created the separate frontend service, delete the old `learn-ai-engineering-frontend` web service in the Render dashboard after the new deployment is live. The single service replaces it.

## Services and database

The Blueprint defines:
- One web service (Docker, build context `.`, root `Dockerfile`, health check at `/health`), on the `free` plan. The image contains the frontend build and the Go backend, so one URL serves the app and the API.
- One Render Postgres database, on the `free` plan. The backend reads `DATABASE_URL` from it.

All resources use Render's free plans. A free Render Postgres expires 30 days after creation, has 1 GB of storage, and each workspace can have only one.

## Same-origin serving

The root `Dockerfile` builds the frontend with `VITE_API_MODE=http` and `VITE_API_BASE_URL=/api` and copies the build into the image. The backend serves the frontend from the directory in `STATIC_DIR` (set to `/app/static` in the image): known files are served as-is, unknown non-API paths fall back to `index.html`, and the whole API is served under `/api` with the prefix stripped. `GET /health` stays available without the prefix for Render's health check. The browser stays same-origin and no CORS is needed.

A free web service spins down after 15 minutes without traffic. The first request after that cold start takes about a minute, so a slow first page load after an idle period is expected.

## Environment variables

### Web service
| Variable | Source | Notes |
| --- | --- | --- |
| `DATABASE_URL` | Render Postgres (from database) | Provided by Render from the linked database. |
| `ENV` | Blueprint | Set to `production`. |
| `HOST` | Blueprint | Set to `0.0.0.0`. |
| `PORT` | Blueprint | Set to `8080`. |
| `SEED_MAINTAINER_EMAIL` | Render dashboard (secret) | Marked `sync: false`. No value stored in the repository. |
| `SEED_MAINTAINER_PASSWORD` | Render dashboard (secret) | Marked `sync: false`. No value stored in the repository. |
| `STATIC_DIR` | Dockerfile | Set to `/app/static`; points the backend at the frontend build in the image. |

### Database
| Variable | Source | Notes |
| --- | --- | --- |
| Render Postgres connection | Render | Exposed to the web service via `DATABASE_URL`. Credentials live only in Render environment variables. |

## Promotion

The maintainer merges `dev` into `main` by hand. With `autoDeployTrigger: checksPass`, Render will deploy from `main` when checks pass.

## Rollback

To roll back, redeploy an earlier commit from the Render dashboard (or revert and merge to `main`). No custom deploy hook is used.

## Notes

- The deploy gate and first verification are handled separately (T34b).
- No registry, no image push and no deploy hook are used.
- Secrets live only in Render environment variables.
- Local use is unchanged: `docker compose up --build` still runs the frontend Caddy image and the backend as separate services.
