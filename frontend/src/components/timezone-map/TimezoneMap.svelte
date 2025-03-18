<script lang="ts">
	import { geoEquirectangular } from "d3-geo";
	import { extent } from "d3-array";
	import { scaleSequential } from "d3-scale";
	import { interpolateRdBu } from "d3-scale-chromatic";
	import { feature } from "topojson-client";
	import type { GeometryCollection, Topology } from "topojson-specification";
	// @ts-expect-error
	import { century, equationOfTime, declination } from "solar-calculator";
	import { Blur, Chart, ClipPath, GeoCircle, GeoPath, Graticule, Svg, Tooltip, antipode } from "layerchart";
	import { timerStore } from "@layerstack/svelte-stores";

	// TODO: embed this data in app
	import { createQuery, queryOptions } from "@tanstack/svelte-query";

	const countriesQuery = createQuery(() => queryOptions({
		queryKey: ["countries110mJson"],
		queryFn: async () => {
			const res = await fetch("https://cdn.jsdelivr.net/npm/world-atlas@2/countries-110m.json");
			const data = await res.json()
			return data as Topology<{countries: GeometryCollection<{ name: string }>; land: GeometryCollection}>;
		}
	}));
	const countriesGeojson = $derived(countriesQuery.data && feature(countriesQuery.data, countriesQuery.data.objects.countries));

	const timezonesQuery = createQuery(() => queryOptions({
		queryKey: ["timezonesJson"],
		queryFn: async () => {
			const res = await fetch("https://www.layerchart.com/data/examples/geo/timezones.json");
			const data = await res.json()
			return data as Topology<{
				timezones: GeometryCollection<{
					objectid: number;
					scalerank: number;
					featurecla: string;
					name: string;
					map_color6: number;
					map_color8: number;
					note: any;
					zone: number;
					utc_format: string;
					time_zone: string;
					iso_8601: string;
					places: string;
					dst_places: any;
					tz_name1st: any;
					tz_namesum: number;
				}>;
			}>;
		}
	}));
	const timezonesGeojson = $derived(timezonesQuery.data && feature(timezonesQuery.data, timezonesQuery.data.objects.timezones));

	const colorScale = $derived(timezonesGeojson && scaleSequential(
		// @ts-expect-error
		extent(timezonesGeojson.features, (d) => d.properties.zone),
		interpolateRdBu
	));

	const dateTimer = timerStore();

	function formatDate(date: Date, timeZone: string | null) {
		let result = "-";
		if (timeZone) {
			try {
				result = new Intl.DateTimeFormat(undefined, {
					timeStyle: "medium",
					dateStyle: "short",
					timeZone,
				}).format(date);
			} catch {}
		}

		return result;
	}

	const now = new Date();
	const day = new Date(+now).setUTCHours(0, 0, 0, 0);
	const t = century(now);
	const longitude = ((day - now.valueOf()) / 864e5) * 360 - 180;
	const sun = [longitude - equationOfTime(t) / 4, declination(t)] as [number, number];
</script>

<div class="h-[480px]">
	{#if countriesGeojson && timezonesGeojson && colorScale}
	<Chart
		geo={{
			projection: geoEquirectangular,
			fitGeojson: countriesGeojson,
		}}
		let:tooltip
	>
		<Svg>
			<GeoPath geojson={{ type: "Sphere" }} class="stroke-surface-content/30" id="globe" />
			<Graticule class="stroke-surface-content/20" />

			<GeoPath geojson={countriesGeojson} id="clip" />
			<ClipPath useId="clip" disabled={true}>
				{#each timezonesGeojson.features as feature}
					<GeoPath
						geojson={feature}
						{tooltip}
						fill={colorScale(feature.properties.zone)}
						class="stroke-gray-900/50 hover:brightness-110"
					/>
				{/each}
			</ClipPath>

			{#each countriesGeojson.features as feature}
				<GeoPath geojson={feature} class="stroke-gray-900/10 fill-gray-900/20 pointer-events-none" />
			{/each}

			<ClipPath useId="globe">
				<Blur>
					<GeoCircle
						center={antipode(sun)}
						class="stroke-none fill-black/50 pointer-events-none"
					/>
				</Blur>
			</ClipPath>
		</Svg>

		<Tooltip.Root let:data>
			{@const { tz_name1st, time_zone } = data.properties}
			<Tooltip.List>
				<Tooltip.Item label="Name" value={tz_name1st} />
				<Tooltip.Item label="Timezone" value={time_zone} />
				<Tooltip.Item
					label="Current time"
					value={formatDate($dateTimer, time_zone.replace("UTC", "").replace("±", "+"))}
				/>
			</Tooltip.List>
		</Tooltip.Root>
	</Chart>
	{/if}
</div>
