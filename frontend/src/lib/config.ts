import {
    PUBLIC_APP_URL,
    PUBLIC_AUTH_ISSUER_URL,
    PUBLIC_AUTH_CLIENT_ID,
    PUBLIC_API_URL,
} from '$env/static/public';

export const APP_URL = PUBLIC_APP_URL;
export const AUTH_ISSUER_URL = PUBLIC_AUTH_ISSUER_URL;
export const AUTH_CLIENT_ID = PUBLIC_AUTH_CLIENT_ID;
export const API_URL = PUBLIC_API_URL || "/api/v1";