# Deployment (Render)

This project deploys to [Render](https://render.com) using a Render Blueprint defined in [`render.yaml`](../render.yaml). No container registry, image push or deploy hook is used; Render builds from the repository Dockerfiles.

## One-time setup: apply the Blueprint

1. In the [Render Dashboard](https://dashboard.render.com), go to **Blueprints → New Blueprint Instance**.
2. Connect the repository and select the branch `main`.
3. Render will detect [`render.yaml`](../render.yaml) and create the services and database defined by the Blueprint.
4. Set the required secret values in the Render dashboard for any variables marked `sync: false` (see below). Render will not store secret values in the repository.
5. Trigger the first deploy after the secrets are set.

## Services and database

The Blueprint defines:
- One backend web service (Docker, build context `backend/`, health check at `/health`).
- One frontend web service (Docker, build context `frontend/`), public origin that proxies `/api/` to the backend over the private network.
- One Render Postgres database. The backend reads `DATABASE_URL` from it.

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
| `API_UPSTREAM` | Render service reference (backend) | Backend host and port on the private network. |
| `VITE_API_MODE` | Blueprint | Set to `http`. Build-time value for the frontend bundle. |
| `VITE_API_BASE_URL` | Blueprint | Set to `/api`. Build-time value for the frontend bundle. |

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
