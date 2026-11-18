FROM oven/bun:1 AS build
WORKDIR /usr/src/app

# Bun's isolated installs link each workspace package's dependencies into its own node_modules, so the
# build runs where the install did.
COPY package.json bun.lock ./
COPY packages packages
RUN bun install --filter=@rezible/frontend --frozen-lockfile
RUN bun run --filter "@rezible/frontend" build

FROM library/caddy:2-alpine AS runtime

# Caddy writes only its state and autosaved configuration, under /data and /config. The image's caddy binary
# has the capability to bind port 80 without root.
RUN addgroup -S -g 10001 caddy \
    && adduser -S -D -H -u 10001 -G caddy caddy \
    && chown -R 10001:10001 /data /config

COPY --from=build /usr/src/app/packages/apps/frontend/Caddyfile /etc/caddy/Caddyfile
COPY --from=build /usr/src/app/packages/apps/frontend/dist /srv

USER 10001:10001

EXPOSE 80
