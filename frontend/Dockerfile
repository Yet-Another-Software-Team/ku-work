# use the official Bun image
# see all versions at https://hub.docker.com/r/oven/bun/tags
FROM oven/bun:1 AS build
WORKDIR /app

ARG GOOGLE_CLIENT_ID
# ARG API_BASE_URL

ENV GOOGLE_CLIENT_ID=$GOOGLE_CLIENT_ID
# ENV API_BASE_URL=$API_BASE_URL

COPY package.json bun.lock ./


# use ignore-scripts to avoid building node modules like better-sqlite3
RUN bun install --frozen-lockfile --ignore-scripts

# Copy the entire project
COPY . .

# Run bun build to generate the output
RUN bun --bun run build

# copy production dependencies and source code into final image
FROM oven/bun:1 AS production
WORKDIR /app

COPY package.json bun.lock ./

# Re-install only production dependencies, ensuring they are correctly resolved
RUN bun install --frozen-lockfile --production --ignore-scripts

# Only .output folder is needed from the build stage
COPY --from=build /app/.output /app/.output

# run the app and suppress deprecation warnings
EXPOSE 3000/tcp
ENTRYPOINT [ "bun", "--bun", "run", "--no-deprecation", "/app/.output/server/index.mjs" ]