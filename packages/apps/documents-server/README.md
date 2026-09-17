# Documents server

The backend signs document-session tokens with `DOCUMENTS__SESSION_SIGNING_SEED_HEX`; this service verifies them with the matching `DOCUMENTS__SESSION_PUBLIC_KEY_HEX`. Each value is 32 bytes encoded as hex.

Generate a pair with `bun run --silent --cwd packages/apps/documents-server generate-session-keys`. Workspace setup persists the generated pair. Deploy matching keys to both services; old encrypted tokens are no longer accepted.

Tokens are validated when authenticating. Expiry enforcement for established connections and frontend credential refresh remain follow-up work.
