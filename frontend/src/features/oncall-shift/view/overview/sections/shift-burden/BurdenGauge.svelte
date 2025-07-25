<script lang="ts">
	import * as echarts from "echarts";
	import EChart, { type ChartProps } from "$components/viz/echart/EChart.svelte";

	const burdenGaugeValues = [10, 10, 10, 10, 10];

	const burdenGaugeData: echarts.GaugeSeriesOption["data"] = [
		{
			name: "Burden\nScore",
			value: burdenGaugeValues.reduce((prev, curr) => (prev + curr)),
			title: {
				offsetCenter: ["0%", "-25%"],
			},
			detail: {
				offsetCenter: ["0%", "-5%"],
			},
			progress: {
				show: false,
				width: 0,
			}
		},
		{
			name: "Event\nFrequency",
			value: 10,
			title: {
				offsetCenter: ["-120%", "50%"],
			},
			detail: {
				offsetCenter: ["-120%", "70%"],
			},
		},
		{
			name: "Life\nImpact",
			value: 20,
			title: {
				offsetCenter: ["-60%", "40%"],
			},
			detail: {
				offsetCenter: ["-60%", "60%"],
			},
		},
		{
			name: "Response\nRequirements",
			value: 30,
			title: {
				offsetCenter: ["0%", "50%"],
			},
			detail: {
				offsetCenter: ["0%", "70%"],
			},
		},
		{
			name: "Time\nImpact",
			value: 40,
			title: {
				offsetCenter: ["60%", "40%"],
			},
			detail: {
				offsetCenter: ["60%", "60%"],
				formatter: "10",
			},
		},
		{
			name: "Support",
			value: 50,
			title: {
				offsetCenter: ["110%", "55%"],
			},
			detail: {
				offsetCenter: ["110%", "70%"],
			},
		},
	];

	const burdenChartOptions = $derived<ChartProps["options"]>({
		series: [
			{
				type: "gauge",
				name: "Burden Score",
				data: burdenGaugeData,
				startAngle: 180,
      			endAngle: 0,
				min: 0,
				max: 100,
				pointer: {
					show: false,
				},
				progress: {
					show: true,
					clip: true,
					width: 18,
					itemStyle: {
						borderWidth: 0,
					}
				},
				axisLine: {
					show: true,
					lineStyle: {
						shadowBlur: 0,
						opacity: .10,
						width: 18,
					}
				},
				tooltip: {
					formatter: (p) => {
						const val = burdenGaugeValues.at(p.dataIndex - 1);
						return `${p.name}: ${val}`;
					}
				},
				axisLabel: { color: "inherit" },
				title: {
					fontSize: 14,
					color: "inherit",
					show: true,
				},
				detail: {
					width: 30,
					height: 10,
					fontSize: 14,
					borderRadius: 3,
					backgroundColor: "inherit",
					color: "black",
					formatter: (val: number) => `${val}`
				},
			},
		],
		tooltip: {
			formatter: "{b} : {c}%",
		},
	});
</script>

<EChart init={echarts.init} options={burdenChartOptions} />