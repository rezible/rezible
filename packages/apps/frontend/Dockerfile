FROM oven/bun:1 AS base
WORKDIR /usr/src/app

FROM base AS install
RUN mkdir -p /temp/dev
COPY package.json bun.lock /temp/dev/
COPY packages /temp/dev/packages
RUN cd /temp/dev && bun install --filter=@rezible/frontend --frozen-lockfile

FROM base AS prerelease
COPY --from=install /temp/dev/node_modules node_modules
COPY package.json bun.lock ./
COPY packages packages

RUN cat <<'EOF' > ./Caddyfile
{
  auto_https off
}
:80 {
	root * /srv
	try_files {path} /index.html
	file_server
}
EOF

RUN bun run --filter "@rezible/frontend" build

FROM library/caddy:2-alpine AS runtime

COPY --from=prerelease /usr/src/app/Caddyfile /etc/caddy/Caddyfile
COPY --from=prerelease /usr/src/app/packages/apps/frontend/dist /srv

EXPOSE 80
