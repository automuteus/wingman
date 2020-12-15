# wingman
Helping your capture client find the perfect match

## Description
Wingman is the externally-facing Socket.io broker for automuteus.

Wingman receives messages from your capture client, and relays that information to other automuteus services, using
Redis as the storage/messaging mechanism.

Wingman also coordinates sending mute/deafen requests (that are provided by Galactus) to any available capture clients
that are running a Discord bot that can be used for mute/deafen

## Environment Variables

### Required

- `REDIS_ADDR`: The host and port at which your Redis database instance is accessible. Ex: `192.168.1.42:6379`

### Optional
- `WINGMAN_PORT`: The port at which Wingman can be reached. This is the externally-facing Socket.io port that capture clients will connect to.
Defaults to 8123.
- `REDIS_USER`: Username for authentication with Redis.
- `REDIS_PASS`: Password for authentication with Postgres.
