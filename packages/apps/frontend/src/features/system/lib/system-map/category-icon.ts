import type { Component } from "svelte";
import RiBox3Line from "remixicon-svelte/icons/box-3-line";
import RiServerLine from "remixicon-svelte/icons/server-line";
import { MapCategory } from "./category";

/** The icon an entity is drawn with, by category: platform resources apart from the units they host. */
export function categoryIcon(category: string): Component {
	if (category === MapCategory.Infrastructure) {
		return RiServerLine;
	}
	return RiBox3Line;
}
