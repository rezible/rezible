# Documents server

The backend encrypts document-session tokens with `DOCUMENTS__SESSION_KEY_HEX`; this service decrypts them with the same 32-byte key encoded as hex.

Tokens are validated when authenticating. Authenticated connections close with reconnectable code `4001` when their session expires, and the frontend fetches fresh credentials for reconnects.
