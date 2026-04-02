import * as rawEnv from '$env/dynamic/public';
import { z } from 'zod';

const env = z.object({
    PUBLIC_API_PATH_BASE: z.string().default("/api/v1"),
    PUBLIC_API_PATH_DOCUMENTS: z.string().default("/api/documents"),
    PUBLIC_AUTH_OIDC_ISSUER_PATH: z.string().default("/dex"),
    PUBLIC_AUTH_OIDC_CLIENT_ID: z.string().default("rezible-app"),
    PUBLIC_AUTH_OIDC_CLIENT_SCOPES: z.string().default("openid profile email"),
}).parse(rawEnv);

export const APP_AUTH_ROUTE_BASE = "/auth";
export const API_PATH_BASE = env.PUBLIC_API_PATH_BASE;
export const API_PATH_DOCUMENTS = env.PUBLIC_API_PATH_DOCUMENTS;
export const AUTH_OIDC_ISSUER_PATH = env.PUBLIC_AUTH_OIDC_ISSUER_PATH;
export const AUTH_OIDC_CLIENT_ID = env.PUBLIC_AUTH_OIDC_CLIENT_ID;
export const AUTH_OIDC_CLIENT_SCOPES = env.PUBLIC_AUTH_OIDC_CLIENT_SCOPES;
