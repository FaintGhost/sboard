FROM mcr.microsoft.com/playwright:v1.50.1-noble

WORKDIR /app

RUN npm install -g bun@1.3.9

COPY package.json ./
RUN bun install --frozen-lockfile

COPY playwright.config.ts ./
COPY tests/ ./tests/

COPY entrypoint.sh ./
RUN chmod +x entrypoint.sh

ENTRYPOINT ["./entrypoint.sh"]
CMD ["bunx", "playwright", "test"]
