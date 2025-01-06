import type { Incident } from "$lib/api";
import { Context } from "runed";
 
export const incidentCtx = new Context<Incident>("incident");