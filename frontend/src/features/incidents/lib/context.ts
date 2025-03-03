import type { Incident, Retrospective, SystemAnalysis } from "$lib/api";
import { Context } from "runed";

export const incidentCtx = new Context<Incident>("incident");
export const retrospectiveCtx = new Context<Retrospective>("retrospective");
export const systemAnalysisIdCtx = new Context<string>("systemAnalysisId");
