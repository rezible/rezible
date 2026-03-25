FROM oven/bun:1 AS base

WORKDIR /usr/src/app

FROM base AS install
RUN mkdir -p /temp/dev
COPY package.json bun.lock packages /temp/dev/
RUN cd /temp/dev && bun install --filter=@rezible/documents-server --frozen-lockfile

RUN mkdir -p /temp/prod
COPY package.json bun.lock packages /temp/prod/
RUN cd /temp/prod && bun install --filter=@rezible/documents-server --frozen-lockfile --production

FROM base AS prerelease
COPY --from=install /temp/dev/node_modules node_modules

# [optional] tests & build
ENV NODE_ENV=production
RUN bun run --filter=@rezible/documents-server build

FROM base AS release
COPY --from=install /temp/prod/node_modules node_modules
COPY --from=prerelease /usr/src/app/index.ts .
COPY --from=prerelease /usr/src/app/package.json .

ENV NODE_ENV=production
ENV HOST=0.0.0.0
ENV PORT=7003

# run the app
USER bun
EXPOSE 7003/tcp
ENTRYPOINT [ "bun", "run", "index.ts" ]