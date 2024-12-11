<script lang="ts">
	import { scaleBand, scaleLinear } from 'd3-scale';
	import { max, range } from 'd3-array';
	import { getDay, getWeek } from 'date-fns';
  
	import { Axis, Chart, Circle, Highlight, Points, Svg, Tooltip } from 'layerchart';
	import { formatDate, PeriodType } from 'svelte-ux';
    import { createDateSeries } from './genData';
  
	const data = createDateSeries({ count: 60, min: 0, max: 20, value: 'integer' });
	const daysOfWeek = ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday'];

	const hack = (d: unknown, fn: Function) => {
		// @ts-ignore
		return fn(d.date)
	}
	// @ts-ignore
	const hack2 = (d: unknown) => d.value;
</script>

<div class="h-[300px] w-1/2 p-4 border rounded bg-surface-100 fill-danger">
    <Chart
      {data}
      x={(d) => hack(d, getWeek)}
      xScale={scaleBand()}
      y={(d) => hack(d, getDay)}
      yScale={scaleBand()}
      yDomain={range(7)}
      r={(d) => hack2(d)}
      padding={{ left: 48, bottom: 36 }}
      tooltip={{ mode: 'band' }}
      let:xScale
      let:yScale
    >
      {@const minBandwidth = Math.min(xScale.bandwidth(), yScale.bandwidth())}
      {@const maxValue = max(data, (d) => d.value) ?? 0}
      {@const rScale = scaleLinear()
        .domain([0, maxValue])
        .range([0, minBandwidth / 2 - 5])}
      <Svg>
        <Axis
          placement="left"
          format={(d) => daysOfWeek[d]}
          grid={{ style: 'stroke-dasharray: 2' }}
          rule
        />
        <Axis placement="bottom" format={(d) => 'Week ' + d} />
        <Points let:points>
          {#each points as point, index}
            <Circle
              cx={point.x}
              cy={point.y}
              r={rScale(point.data.value)}
              class="fill-primary/10 stroke-primary"
            />
          {/each}
        </Points>
        <Highlight area axis="x" />
        <Highlight area axis="y" />
      </Svg>
    </Chart>
  </div>