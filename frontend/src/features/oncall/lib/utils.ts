import type { OncallShift } from "$lib/api";
import { settings } from "$lib/settings.svelte";
import type { ZonedDateTime } from "@internationalized/date";
import { DateToken, PeriodType } from "@layerstack/utils";
import { differenceInMinutes, isFuture, isPast } from "date-fns";

export type ShiftStatus = "active" | "upcoming" | "finished";
export type ShiftTimeDetails = {
	start: Date;
	end: Date;
	progress: number;
	minutesLeft: number;

	status: ShiftStatus;
};

export const buildShiftTimeDetails = (shift: OncallShift): ShiftTimeDetails => {
	const attr = shift.attributes;
	const start = new Date(attr.startAt);
	const end = new Date(attr.endAt);
	const progress = (Date.now() - start.valueOf()) / (end.valueOf() - start.valueOf());
	const minutesLeft = differenceInMinutes(end, Date.now());
	const status: ShiftStatus = isPast(end) ? "finished" : isFuture(start) ? "upcoming" : "active";
	return { start, end, progress, minutesLeft, status };
};

export const formatShiftDates = (shift: OncallShift) => {
	const startFmt = settings.format(new Date(shift.attributes.startAt), PeriodType.Day);
	const endFmt = settings.format(new Date(shift.attributes.endAt), PeriodType.Day);
	const rosterName = shift.attributes.roster.attributes.name;
	return `${rosterName} - ${startFmt} to ${endFmt}`;
};

export type ShiftEvent = {
	id: string;
	timestamp: ZonedDateTime;
	eventType: "alert" | "incident";
	title?: string;
	description?: string;
	severity?: "critical" | "high" | "medium" | "low";
	status?: "active" | "resolved" | "acknowledged";
	source?: string;
	annotation?: string;
};

export type ShiftEventFilterKind = "alerts" | "nightAlerts" | "incidents";

export const isBusinessHours = (hour: number) => {
	return hour >= 9 && hour < 17; // 9am to 5pm
};

export const isNightHours = (hour: number) => {
	return hour >= 22 || hour < 6; // 10pm to 6am
};

export const shiftEventMatchesFilter = (event: ShiftEvent, kind: ShiftEventFilterKind) => {
	if ((kind === "alerts" || kind === "nightAlerts") && event.eventType !== "alert") return false;
	if (kind === "nightAlerts" && (event.timestamp.hour < 18 && event.timestamp.hour > 6)) return false;
	if (kind === "incidents" && event.eventType !== "incident") return false;
	return true;
}

export type ShiftMetrics = {
	totalAlerts: number;
	totalIncidents: number;
	nightAlerts: number;
	avgResponseTime: number; // in minutes
	escalationRate: number; // percentage
	totalIncidentTime: number; // in minutes
	longestIncident: number; // in minutes
	businessHoursAlerts: number;
	offHoursAlerts: number;
	peakAlertHour: number;
	totalOncallTime: number; // in minutes
	severityBreakdown: {
		critical: number;
		high: number;
		medium: number;
		low: number;
	};
	sleepDisruptionScore: number; // 0-100
	workloadScore: number; // 0-100
	burdenScore: number; // 0-100
};

export type ComparisonMetrics = {
	alertsComparison: number; // percentage difference from average
	incidentsComparison: number;
	responseTimeComparison: number;
	escalationRateComparison: number;
	nightAlertsComparison: number;
	severityComparison: {
		critical: number;
		high: number;
		medium: number;
		low: number;
	};
};

export const calculateShiftMetrics = (events: ShiftEvent[], shiftDetails: ShiftTimeDetails): ShiftMetrics => {
	const alerts = events.filter(e => e.eventType === "alert");
	const incidents = events.filter(e => e.eventType === "incident");
	const nightAlerts = alerts.filter(e => isNightHours(e.timestamp.hour));
	const businessHoursAlerts = alerts.filter(e => isBusinessHours(e.timestamp.hour));
	const offHoursAlerts = alerts.filter(e => !isBusinessHours(e.timestamp.hour));
	
	// Calculate peak alert hour
	const hourCounts = new Array(24).fill(0);
	alerts.forEach(alert => {
		hourCounts[alert.timestamp.hour]++;
	});
	const peakAlertHour = hourCounts.indexOf(Math.max(...hourCounts));
	
	// Mock response times (in minutes)
	const responseTimes = alerts.map(() => Math.floor(Math.random() * 30) + 1);
	const avgResponseTime = responseTimes.length ? 
		responseTimes.reduce((sum, time) => sum + time, 0) / responseTimes.length : 0;
	
	// Mock incident durations (in minutes)
	const incidentDurations = incidents.map(() => Math.floor(Math.random() * 180) + 30);
	const totalIncidentTime = incidentDurations.reduce((sum, time) => sum + time, 0);
	const longestIncident = incidentDurations.length ? Math.max(...incidentDurations) : 0;
	
	// Severity breakdown
	const severityCounts = {
		critical: incidents.filter(i => i.severity === "critical").length,
		high: incidents.filter(i => i.severity === "high").length,
		medium: incidents.filter(i => i.severity === "medium").length,
		low: incidents.filter(i => i.severity === "low").length
	};
	
	// Calculate scores
	const sleepDisruptionScore = Math.min(100, (nightAlerts.length * 20));
	const workloadScore = Math.min(100, (alerts.length * 5) + (incidents.length * 15));
	const burdenScore = Math.round((sleepDisruptionScore + workloadScore) / 2);
	
	return {
		totalAlerts: alerts.length,
		totalIncidents: incidents.length,
		nightAlerts: nightAlerts.length,
		avgResponseTime,
		escalationRate: alerts.length ? (incidents.length / alerts.length) * 100 : 0,
		totalIncidentTime,
		longestIncident,
		businessHoursAlerts: businessHoursAlerts.length,
		offHoursAlerts: offHoursAlerts.length,
		peakAlertHour,
		totalOncallTime: differenceInMinutes(shiftDetails.end, shiftDetails.start),
		severityBreakdown: severityCounts,
		sleepDisruptionScore,
		workloadScore,
		burdenScore
	};
};

export const getHourLabel = (hour: number): string => {
	const ampm = hour >= 12 ? 'PM' : 'AM';
	const displayHour = hour % 12 || 12;
	return `${displayHour}${ampm}`;
};

export const formatDuration = (minutes: number): string => {
	if (minutes < 60) return `${minutes}m`;
	const hours = Math.floor(minutes / 60);
	const mins = minutes % 60;
	if (mins === 0) return `${hours}h`;
	return `${hours}h ${mins}m`;
};

export const formatPercentage = (value: number): string => {
	return `${Math.round(value)}%`;
};

export const formatComparisonValue = (value: number): string => {
	const sign = value > 0 ? '+' : '';
	return `${sign}${Math.round(value)}%`;
};

