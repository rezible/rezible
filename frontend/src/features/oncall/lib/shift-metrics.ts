import { isBusinessHours, isNightHours, type ShiftEvent, type ShiftTimeDetails } from "./utils";

export type ShiftMetrics = {
	totalAlerts: number;
	totalIncidents: number;
	nightAlerts: number;
	alertIncidentRate: number; // percentage
	totalIncidentTime: number; // in minutes
	longestIncident: number; // in minutes
	businessHoursAlerts: number;
	offHoursAlerts: number;
	offHoursTime: number;
	sleepDisruptionScore: number; // 0-100
	workloadScore: number; // 0-100
	burdenScore: number; // 0-100
};

export const makeFakeShiftMetrics = (): ShiftMetrics => ({
	totalAlerts: Math.floor(Math.random() * 15) + 5,
	totalIncidents: Math.floor(Math.random() * 5) + 1,
	nightAlerts: Math.floor(Math.random() * 6) + 1,
	alertIncidentRate: Math.floor(Math.random() * 30) + 10,
	totalIncidentTime: Math.floor(Math.random() * 240) + 60,
	longestIncident: Math.floor(Math.random() * 120) + 30,
	businessHoursAlerts: Math.floor(Math.random() * 10) + 3,
	offHoursAlerts: Math.floor(Math.random() * 8) + 2,
	offHoursTime: Math.floor(Math.random() * 120) + 30,
	sleepDisruptionScore: Math.floor(Math.random() * 70) + 10,
	workloadScore: Math.floor(Math.random() * 80) + 20,
	burdenScore: Math.floor(Math.random() * 75) + 15,
});

export type ComparisonMetrics = {
	alertsComparison: number;
	incidentsComparison: number;
	escalationRateComparison: number;
	nightAlertsComparison: number;
};

export const makeFakeComparisonMetrics = (): ComparisonMetrics => ({
	alertsComparison: .7 + (Math.random() * .6),
	incidentsComparison: .7 + (Math.random() * .6),
	escalationRateComparison: .9 + (Math.random() * .4),
	nightAlertsComparison: .8 + (Math.random() * .8),
});

export const calculateShiftMetrics = (events: ShiftEvent[], shiftDetails: ShiftTimeDetails): ShiftMetrics => {
	const alerts = events.filter(e => e.eventType === "alert");
	const incidents = events.filter(e => e.eventType === "incident");
	const nightAlerts = alerts.filter(e => isNightHours(e.timestamp.hour));
	const businessHoursAlerts = alerts.filter(e => isBusinessHours(e.timestamp.hour));
	const offHoursAlerts = alerts.filter(e => !isBusinessHours(e.timestamp.hour));

	const incidentDurations = incidents.map(() => Math.floor(Math.random() * 180) + 30);
	const totalIncidentTime = incidentDurations.reduce((sum, time) => sum + time, 0);
	const longestIncident = incidentDurations.length ? Math.max(...incidentDurations) : 0;

	// Calculate scores
	const sleepDisruptionScore = Math.min(100, (nightAlerts.length * 20));
	const workloadScore = Math.min(100, (alerts.length * 5) + (incidents.length * 15));
	const burdenScore = Math.round((sleepDisruptionScore + workloadScore) / 2);

	return {
		totalAlerts: alerts.length,
		totalIncidents: incidents.length,
		nightAlerts: nightAlerts.length,
		alertIncidentRate: alerts.length ? (incidents.length / alerts.length) * Math.random() : 0,
		totalIncidentTime,
		longestIncident,
		businessHoursAlerts: businessHoursAlerts.length,
		offHoursAlerts: offHoursAlerts.length,
		offHoursTime: Math.floor(Math.random() * 120) + 30,
		sleepDisruptionScore,
		workloadScore,
		burdenScore
	};
};