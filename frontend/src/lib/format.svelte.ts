
export const formatPercentage = (value: number) => (`${Math.round(value)}%`);

export const formatDelta = (value: number) => (`${value > 1 ? '+' : ''}${Math.round((value - 1) * 100)}%`);