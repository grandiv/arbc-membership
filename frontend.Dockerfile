# Frontend container — Next.js static export served by nginx, with /api/
# proxied to the BFF container. Build context = arbc-membership/ root so we
# COPY both frontend/ and nginx.conf.
# node:22 — corepack's pnpm needs node:sqlite (Node 22+).
# --platform=$BUILDPLATFORM: build the static export natively (output is
# arch-neutral HTML/JS); only the final nginx stage is the target arch.
FROM --platform=$BUILDPLATFORM node:22-alpine AS builder
WORKDIR /app
RUN corepack enable
COPY frontend/package.json frontend/pnpm-lock.yaml* frontend/pnpm-workspace.yaml* ./
RUN pnpm install --frozen-lockfile || pnpm install
COPY frontend/ .
# Empty API base → client fetches go same-origin (/api/*), proxied by nginx.
ENV NEXT_PUBLIC_API_URL=""
RUN pnpm build

FROM nginx:alpine
COPY --from=builder /app/out /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
