// import type { Incident, Retrospective, SystemAnalysis } from "$lib/api";
import type { OncallShift } from "$lib/api";
import { Context } from "runed";

export const shiftCtx = new Context<OncallShift>("shift");
