// import type { Incident, Retrospective, SystemAnalysis } from "$lib/api";
import type { OncallShift } from "$lib/api";
import { Context } from "runed";

export const shiftIdCtx = new Context<string>("shiftId");
