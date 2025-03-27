export type MetricComparison = {
	value: number;
	positive?: boolean;
	icon?: string;
	deltaLabel?: string;
	hint?: string;
}

export const comparisonClass = (c: MetricComparison) => {
	const low = (c.positive ? 1 : c.value);
	const high = (c.positive ? c.value : 1);
	if (low > high) return "text-red-500";
	if (low < high) return "text-green-500";
	return "text-gray-500";
};