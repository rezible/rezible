import themesCfg from "$src/themes.json";
import { settings as uxSettings, ThemeInit } from "svelte-ux";

const createSettingsState = () => {

	const setup = () => {
		const themes = themesCfg as Record<string, Record<string, string>>;
		const themeNames = Object.keys(themes);
		const lightThemes = themeNames.filter(name => themes[name]["color-scheme"] === "light");
		const darkThemes = themeNames.filter(name => themes[name]["color-scheme"] === "dark");
	
		console.log({ lightThemes, darkThemes });
		uxSettings({
			themes: { light: lightThemes, dark: darkThemes },
			components: {},
			localeFormats: {},
		});
	}

	return {
		setup,
	}
}

export const settings = createSettingsState();