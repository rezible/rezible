FROM oven/bun:1 AS base
WORKDIR /usr/src/app

FROM base AS build
# Bun's isolated installs link each workspace package's dependencies into its own node_modules, so the
# build runs where the install did.
COPY package.json bun.lock ./
COPY packages packages
RUN bun install --filter=@rezible/documents-server --frozen-lockfile
ENV NODE_ENV=production
RUN bun run --filter=@rezible/documents-server build

FROM base AS release

# The build bundles every dependency into index.js.
COPY --from=build /usr/src/app/packages/apps/documents-server/dist/index.js .

ENV NODE_ENV=production
ENV HOST=0.0.0.0
ENV PORT=7003

USER bun
EXPOSE 7003/tcp
ENTRYPOINT ["bun", "run", "index.js"]
