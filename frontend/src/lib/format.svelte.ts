
export const formatPercentage = (value: number) => (`${Math.round(value)}%`);

export const formatDelta = (value: number) => (`${value > 0 ? '+' : ''}${Math.round(value)}%`);