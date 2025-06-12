import uxThemesCfg from "$src/svelteux-themes.json";
import { settings as setUxSettings, getSettings as getUxSettings } from "svelte-ux";
import { fromStore } from "svelte/store";

const createSettingsState = () => {
	let settings = $state.raw(getUxSettings());

	const currentTheme = $derived(fromStore(settings.currentTheme));
	const locale = $derived(fromStore(settings.locale));
	const format = $derived(fromStore(settings.format));

	const setup = () => {
		const uxThemes = uxThemesCfg as Record<string, Record<string, string>>;
		const themeNames = Object.keys(uxThemes);
		const lightThemes = themeNames.filter(name => uxThemes[name]["color-scheme"] === "light");
		const darkThemes = themeNames.filter(name => uxThemes[name]["color-scheme"] === "dark");

		settings = setUxSettings({
			themes: { light: lightThemes, dark: darkThemes },
			components: {},
		});
	}

	return {
		setup,
		get theme() { return currentTheme.current },
		get locale() { return locale.current },
		get format() { return format.current },
	}
}

export const settings = createSettingsState();