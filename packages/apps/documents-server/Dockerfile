FROM oven/bun:1 AS base
WORKDIR /usr/src/app

FROM base AS install
RUN mkdir -p /temp/dev
COPY package.json bun.lock /temp/dev/
COPY packages /temp/dev/packages
RUN cd /temp/dev && bun install --filter=@rezible/documents-server --save-dev

RUN mkdir -p /temp/prod
COPY package.json bun.lock /temp/prod/
COPY packages /temp/prod/packages
RUN cd /temp/prod && bun install --filter=@rezible/documents-server --frozen-lockfile --production

FROM base AS prerelease
COPY --from=install /temp/dev/node_modules node_modules
COPY package.json bun.lock ./
COPY packages packages

ENV NODE_ENV=production
RUN bun run --filter=@rezible/documents-server build

FROM base AS release

COPY --from=install /temp/prod/node_modules node_modules
COPY --from=prerelease /usr/src/app/packages/apps/documents-server/dist/index.js .
COPY --from=prerelease /usr/src/app/packages/apps/documents-server/package.json .

ENV NODE_ENV=production
ENV HOST=0.0.0.0
ENV PORT=7003

USER bun
EXPOSE 7003/tcp
ENTRYPOINT ["bun", "run", "index.js"]