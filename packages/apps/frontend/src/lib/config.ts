import * as rawEnv from '$env/dynamic/public';
import { z } from 'zod';

const env = z.object({
    PUBLIC_API_PATH_BASE: z.string().default("/api/v1"),
    PUBLIC_AUTH_OIDC_ISSUER_PATH: z.string().default("/auth"),
    PUBLIC_AUTH_OIDC_CLIENT_ID: z.string().default("rezible-app"),
    PUBLIC_AUTH_OIDC_CLIENT_SCOPES: z.string().default("openid profile email"),
    PUBLIC_AUTH_OIDC_CLIENT_REDIRECT_PATH: z.string().default("/login/callback"),
}).parse(rawEnv);

export const API_PATH_BASE = env.PUBLIC_API_PATH_BASE;
export const AUTH_OIDC_ISSUER_PATH = env.PUBLIC_AUTH_OIDC_ISSUER_PATH;
export const AUTH_OIDC_CLIENT_ID = env.PUBLIC_AUTH_OIDC_CLIENT_ID;
export const AUTH_OIDC_CLIENT_SCOPES = env.PUBLIC_AUTH_OIDC_CLIENT_SCOPES;
export const AUTH_OIDC_CLIENT_REDIRECT_PATH = env.PUBLIC_AUTH_OIDC_CLIENT_REDIRECT_PATH;
