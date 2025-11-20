# use the official Bun image
# see all versions at https://hub.docker.com/r/oven/bun/tags
FROM oven/bun:1.3 AS build
WORKDIR /app

ARG GOOGLE_CLIENT_ID
ARG TURNSTILE_CLIENT_TOKEN

ENV GOOGLE_CLIENT_ID=$GOOGLE_CLIENT_ID
ENV TURNSTILE_CLIENT_TOKEN=$TURNSTILE_CLIENT_TOKEN

COPY package.json bun.lock ./

# Re-install only production dependencies, ensuring they are correctly resolved
RUN bun install --frozen-lockfile --production --ignore-scripts

# Copy the entire project
COPY . .

# Run bun build to generate the output
RUN bun --bun run build

# copy production dependencies and source code into final image
FROM oven/bun:1.3 AS production
WORKDIR /app

COPY package.json bun.lock ./

# Re-install only production dependencies, ensuring they are correctly resolved
RUN bun install --frozen-lockfile --production --ignore-scripts

# Only .output folder is needed from the build stage
COPY --from=build /app/.output /app/.output

# run the app and suppress deprecation warnings
EXPOSE 3000/tcp
ENTRYPOINT [ "bun", "--bun", "run", "--no-deprecation", "/app/.output/server/index.mjs" ]
