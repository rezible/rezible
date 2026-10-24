FROM oven/bun:1 AS build
WORKDIR /usr/src/app

# Bun's isolated installs link each workspace package's dependencies into its own node_modules, so the
# build runs where the install did.
COPY package.json bun.lock ./
COPY packages packages
RUN bun install --filter=@rezible/frontend --frozen-lockfile
RUN bun run --filter "@rezible/frontend" build

FROM library/caddy:2-alpine AS runtime

COPY --from=build /usr/src/app/packages/apps/frontend/Caddyfile /etc/caddy/Caddyfile
COPY --from=build /usr/src/app/packages/apps/frontend/dist /srv

EXPOSE 80
