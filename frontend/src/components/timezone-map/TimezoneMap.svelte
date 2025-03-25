<script lang="ts">
	import { base } from "$app/paths";
	// TODO: embed this data in app
	import { createQuery, queryOptions } from "@tanstack/svelte-query";
	import type { CountriesTopology, TimezonesTopology } from "./TimezoneChart.svelte";
	import TimezoneChart from "./TimezoneChart.svelte";

	const countriesQuery = createQuery(() => queryOptions({
		queryKey: ["countries110mJson"],
		queryFn: async () => {
			const res = await fetch(base + "/countries-110m.json");
			const data = await res.json();
			return data as CountriesTopology;
		}
	}));
	const countriesData = $derived(countriesQuery.data);


	const timezonesQuery = createQuery(() => queryOptions({
		queryKey: ["timezonesJson"],
		queryFn: async () => {
			const res = await fetch(base + "/timezones.json");
			const data = await res.json();
			return data as TimezonesTopology;
		}
	}));
	const timezonesData = $derived(timezonesQuery.data);

</script>

<div class="h-full">
	{#if countriesData && timezonesData}
		<TimezoneChart {countriesData} {timezonesData} />
	{/if}
</div>
