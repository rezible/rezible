import type { Incident, Retrospective } from "$lib/api";
import { Context } from "runed";
 
export const incidentCtx = new Context<Incident>("incident");
export const retrospectiveCtx = new Context<Retrospective>("retrospective");