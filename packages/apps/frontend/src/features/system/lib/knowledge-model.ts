import { getKnowledgeGraphModelOptions } from "$lib/api";

/** The model changes only with a deploy, so one load serves the session. */
export const knowledgeModelOptions = () => ({ ...getKnowledgeGraphModelOptions(), staleTime: Infinity });
