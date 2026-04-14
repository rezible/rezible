import { env as rawEnv } from '$env/dynamic/public';
import { z } from 'zod';

const env = z.object({
    PUBLIC_API_PATH_BASE: z.string().default("/api"),
}).parse(rawEnv);

export const API_PATH_BASE = env.PUBLIC_API_PATH_BASE;
