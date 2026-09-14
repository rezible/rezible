import type { ParamMatcher } from "@sveltejs/kit";

export type DesignSystemSection = "controls" | "content" | "overlays";

const sections = new Set<DesignSystemSection>(["controls", "content", "overlays"]);

export const match = ((param?: string): param is DesignSystemSection => {
	return param !== undefined && sections.has(param as DesignSystemSection);
}) satisfies ParamMatcher;
