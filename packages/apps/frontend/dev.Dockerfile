FROM oven/bun:1 AS base

WORKDIR /usr/src/app

COPY package.json bun.lock ./
COPY packages/ packages
RUN bun install --filter=@rezible/frontend --frozen-lockfile

EXPOSE 7000

CMD ["bun", "run", "--filter=@rezible/frontend", "dev"]