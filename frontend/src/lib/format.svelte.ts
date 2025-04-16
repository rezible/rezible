
export const formatPercentage = (value: number) => (`${Math.round(value)}%`);

export const hour12 = (hour: number) => {
	if (hour > 12) return hour - 12;
	if (hour == 0) return 12;
	return hour;
};

export const hour12Label = (hour: number): string => {
	const ampm = hour >= 12 ? 'PM' : 'AM';
	return `${hour12(hour)}${ampm}`;
};