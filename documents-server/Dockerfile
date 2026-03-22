FROM oven/bun:1 AS build

WORKDIR /app

# Copy workspace manifests first so dependency installation can be cached.
COPY package.json bun.lock ./
COPY documents-server/package.json documents-server/package.json
COPY shared/tiptap-extensions/package.json shared/tiptap-extensions/package.json

RUN bun install --frozen-lockfile

# Copy only the sources needed to build the documents server bundle.
COPY tsconfig.json tsconfig.json
COPY documents-server documents-server
COPY shared/tiptap-extensions shared/tiptap-extensions

WORKDIR /app/documents-server
RUN bun run build

FROM oven/bun:1-slim AS runtime

WORKDIR /app

COPY --from=build /app/documents-server/dist ./dist

ENV NODE_ENV=production
ENV HOST=0.0.0.0
ENV PORT=7003

EXPOSE 7003

CMD ["bun", "run", "./dist/index.js"]
