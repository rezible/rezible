import themesCfg from "$src/themes.json";
import { settings as setUxSettings, getSettings as getUxSettings } from "svelte-ux";
import { fromStore } from "svelte/store";

const createSettingsState = () => {
	let settings = $state(getUxSettings());

	const currentTheme = $derived(fromStore(settings.currentTheme));
	const locale = $derived(fromStore(settings.locale));
	const format = $derived(fromStore(settings.format));

	const setup = () => {
		const themes = themesCfg as Record<string, Record<string, string>>;
		const themeNames = Object.keys(themes);
		const lightThemes = themeNames.filter(name => themes[name]["color-scheme"] === "light");
		const darkThemes = themeNames.filter(name => themes[name]["color-scheme"] === "dark");

		settings = setUxSettings({
			themes: { light: lightThemes, dark: darkThemes },
			components: {},
			localeFormats: {},
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