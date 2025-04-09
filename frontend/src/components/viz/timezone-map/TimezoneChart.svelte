<script lang="ts" module>
	export type CountriesTopology = Topology<{
		countries: GeometryCollection<{ name: string }>;
		land: GeometryCollection;
	}>;

	export type TimezonesTopology = Topology<{
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
</script>

<script lang="ts">
	import { geoEquirectangular } from "d3-geo";
	import { extent } from "d3-array";
	import { scaleSequential } from "d3-scale";
	import { interpolateRdBu } from "d3-scale-chromatic";
	import { feature } from "topojson-client";
	import type { GeometryCollection, Topology } from "topojson-specification";
	// @ts-expect-error
	import { century, equationOfTime, declination } from "solar-calculator";
	import {
		Blur,
		Canvas,
		Chart,
		ClipPath,
		GeoCircle,
		GeoPath,
		Graticule,
		antipode,
	} from "layerchart";

	type Props = {
		countriesData: CountriesTopology;
		timezonesData: TimezonesTopology;
	};
	const { timezonesData, countriesData }: Props = $props();

	const countriesGeojson = $derived(feature(countriesData, countriesData.objects.countries));
	const timezonesGeojson = $derived(feature(timezonesData, timezonesData.objects.timezones));

	const colorExtent = $derived(extent(timezonesGeojson.features, (d) => d.properties.zone));

	// @ts-expect-error
	const colorScale = $derived(scaleSequential(colorExtent, interpolateRdBu));
	// const dateTimer = timerStore();

	// function formatDate(date: Date, timeZone: string | null) {
	// 	let result = "-";
	// 	if (timeZone) {
	// 		try {
	// 			result = new Intl.DateTimeFormat(undefined, {
	// 				timeStyle: "medium",
	// 				dateStyle: "short",
	// 				timeZone,
	// 			}).format(date);
	// 		} catch {}
	// 	}

	// 	return result;
	// }

	const getSunAntipode = () => {
		const now = new Date();
		const day = new Date(+now).setUTCHours(0, 0, 0, 0);
		const t = century(now);
		const longitude = ((day - now.valueOf()) / 864e5) * 360 - 180;
		const sunPos = [longitude - equationOfTime(t) / 4, declination(t)] as [number, number];
		return antipode(sunPos);
	};
	const sunAntipode = getSunAntipode();

</script>

<Chart geo={{ projection: geoEquirectangular, fitGeojson: countriesGeojson }}>
	<Canvas>
		<GeoPath geojson={{ type: "Sphere" }} class="stroke-surface-content/30" id="globe" />
		<Graticule class="stroke-surface-content/20" />

		<GeoPath geojson={countriesGeojson} id="clip" />
		<ClipPath useId="clip" disabled={true}>
			{#each timezonesGeojson.features as feature}
				<GeoPath
					geojson={feature}
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
				<GeoCircle center={sunAntipode} class="stroke-none fill-black/50 pointer-events-none" />
			</Blur>
		</ClipPath>
	</Canvas>

	<!--Tooltip.Root let:data>
		{@const { tz_name1st, time_zone } = data.properties}
		<Tooltip.List>
			<Tooltip.Item label="Name" value={tz_name1st} />
			<Tooltip.Item label="Timezone" value={time_zone} />
			<Tooltip.Item
				label="Current time"
				value={formatDate($dateTimer, time_zone.replace("UTC", "").replace("±", "+"))}
			/>
		</Tooltip.List>
	</Tooltip.Root-->
</Chart>
