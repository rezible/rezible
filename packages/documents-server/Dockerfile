FROM oven/bun:1 AS build

WORKDIR /app

COPY . .

RUN bun install --frozen-lockfile

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
