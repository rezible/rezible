import uxThemesCfg from "$src/svelteux-themes.json";
import { Context } from "runed";
import { settings as setUxSettings, getSettings as getUxSettings } from "svelte-ux";
import { fromStore } from "svelte/store";

export type UxSettings = ReturnType<typeof getUxSettings>;

export class SettingsState {
	uxSettings: UxSettings = $state.raw(getUxSettings());

	currentTheme = $derived(fromStore(this.uxSettings.currentTheme));
	locale = $derived(fromStore(this.uxSettings.locale));
	format = $derived(fromStore(this.uxSettings.format).current);

	setup() {
		const uxThemes = uxThemesCfg as Record<string, Record<string, string>>;
		const themeNames = Object.keys(uxThemes);
		const lightThemes = themeNames.filter(name => uxThemes[name]["color-scheme"] === "light");
		const darkThemes = themeNames.filter(name => uxThemes[name]["color-scheme"] === "dark");

		this.uxSettings = setUxSettings({
			themes: { light: lightThemes, dark: darkThemes },
			components: {},
		});
	}
}

export const settings = new SettingsState();