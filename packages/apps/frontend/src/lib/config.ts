import * as rawEnv from '$env/dynamic/public';
import { z } from 'zod';

const env = z.object({
    PUBLIC_API_PATH_BASE: z.string().default("/api"),
    PUBLIC_AUTH_PATH_BASE: z.string().default("/auth"),
    PUBLIC_AUTH_OIDC_CLIENT_ID: z.string().default("rezible-app"),
    PUBLIC_AUTH_OIDC_CLIENT_SCOPES: z.string().default("openid profile email"),
}).parse(rawEnv);

export const APP_LOGIN_ROUTE_BASE = "/login";
export const API_PATH_BASE = env.PUBLIC_API_PATH_BASE;
export const AUTH_PATH_BASE = env.PUBLIC_AUTH_PATH_BASE;
export const AUTH_OIDC_CLIENT_ID = env.PUBLIC_AUTH_OIDC_CLIENT_ID;
export const AUTH_OIDC_CLIENT_SCOPES = env.PUBLIC_AUTH_OIDC_CLIENT_SCOPES;
