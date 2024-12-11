import { mdiAccountGroup, mdiFire, mdiPhoneLog } from "@mdi/js";
import type { MenuOption } from "svelte-ux";

type SearchState = {

}

type SearchType = "general" | "oncall";
type SearchOption = MenuOption<string>;

const generalOptions: MenuOption<string>[] = [
	{label: "Incident Foo Bar", value: "incident-foo-bar", icon: mdiFire},
	{label: "Team Something", value: "team-id", icon: mdiAccountGroup},
	{label: "Demo Roster", value: "roster-id", icon: mdiPhoneLog},
]

const oncallOptions: MenuOption<string>[] = [
	{label: "Recommendations Service", value: "foo-id"},
]

const generalKey = "k";
const oncallKey = "o";

const createSearchState = () => {
	let searchInput = $state("");
	let searchType = $state<SearchType>("general");
	let options = $state<SearchOption[]>([]);

	const isSearchKeyPress = (e: KeyboardEvent): SearchType | false => {
		if (!e.metaKey && !e.ctrlKey) return false;
		
		const key = e.key;
		if (key === generalKey) return "general";
		if (key === oncallKey) return "oncall";
		return false;
	}

	const startSearch = (t: SearchType) => {
		searchType = t;
		updateInput("");
	}

	const updateInput = (v: string) => {
		console.log("input", v);
		searchInput = v;
		updateOptions();
	}

	const updateOptions = () => {
		const all = searchType === "general" ? generalOptions : oncallOptions;
		const input = searchInput.toLowerCase();
		if (input === "") {
			options = [];
			return;
		}
		options = all.filter(o => (o.label.toLowerCase().includes(input)));
	}

	const clear = () => {
		searchInput = ""
	}

	return {
		clear,
		isSearchKeyPress,
		startSearch,
		updateInput,
		get searchType() { return searchType },
		get options() { return options },
   };
}
export const searchState = createSearchState();