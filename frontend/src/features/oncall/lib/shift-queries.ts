import { createQuery } from "@tanstack/svelte-query";
import type { OncallShift } from "$lib/api";
import type { ComparisonMetrics, ShiftMetrics } from "./utils";

export const useShiftMetricsQuery = (shift: OncallShift) => {
  return createQuery({
    queryKey: ["shiftMetrics", shift.id],
    queryFn: async (): Promise<ShiftMetrics> => {
      // Simulate API delay
      await new Promise(resolve => setTimeout(resolve, 500));
      
      // Return mock data
      return {
        totalAlerts: Math.floor(Math.random() * 15) + 5,
        totalIncidents: Math.floor(Math.random() * 5) + 1,
        nightAlerts: Math.floor(Math.random() * 6) + 1,
        avgResponseTime: Math.floor(Math.random() * 20) + 5,
        escalationRate: Math.floor(Math.random() * 30) + 10,
        totalIncidentTime: Math.floor(Math.random() * 240) + 60,
        longestIncident: Math.floor(Math.random() * 120) + 30,
        businessHoursAlerts: Math.floor(Math.random() * 10) + 3,
        offHoursAlerts: Math.floor(Math.random() * 8) + 2,
        peakAlertHour: Math.floor(Math.random() * 24),
        totalOncallTime: 24 * 60, // 24 hours in minutes
        severityBreakdown: {
          critical: Math.floor(Math.random() * 2),
          high: Math.floor(Math.random() * 3) + 1,
          medium: Math.floor(Math.random() * 4) + 1,
          low: Math.floor(Math.random() * 3)
        },
        sleepDisruptionScore: Math.floor(Math.random() * 70) + 10,
        workloadScore: Math.floor(Math.random() * 80) + 20,
        burdenScore: Math.floor(Math.random() * 75) + 15
      };
    },
    staleTime: 5 * 60 * 1000, // 5 minutes
  });
};

export const useShiftComparisonQuery = (shift: OncallShift) => {
  return createQuery({
    queryKey: ["shiftComparison", shift.id],
    queryFn: async (): Promise<ComparisonMetrics> => {
      // Simulate API delay
      await new Promise(resolve => setTimeout(resolve, 700));
      
      // Return mock comparison data (percentage difference from average)
      return {
        alertsComparison: Math.floor(Math.random() * 60) - 30, // -30% to +30%
        incidentsComparison: Math.floor(Math.random() * 70) - 35,
        responseTimeComparison: Math.floor(Math.random() * 50) - 25,
        escalationRateComparison: Math.floor(Math.random() * 40) - 20,
        nightAlertsComparison: Math.floor(Math.random() * 80) - 40,
        severityComparison: {
          critical: Math.floor(Math.random() * 60) - 30,
          high: Math.floor(Math.random() * 50) - 25,
          medium: Math.floor(Math.random() * 40) - 20,
          low: Math.floor(Math.random() * 30) - 15
        }
      };
    },
    staleTime: 5 * 60 * 1000, // 5 minutes
  });
};

export const useHourlyDistributionQuery = (shift: OncallShift) => {
  return createQuery({
    queryKey: ["hourlyDistribution", shift.id],
    queryFn: async () => {
      // Simulate API delay
      await new Promise(resolve => setTimeout(resolve, 600));
      
      // Generate mock hourly distribution data
      const hours = Array.from({ length: 24 }, (_, i) => i);
      return hours.map(hour => ({
        hour,
        alerts: Math.floor(Math.random() * (isBusinessHours(hour) ? 3 : 1.5)),
        incidents: Math.random() > 0.8 ? Math.floor(Math.random() * 2) : 0
      }));
    },
    staleTime: 5 * 60 * 1000, // 5 minutes
  });
};

function isBusinessHours(hour: number): boolean {
  return hour >= 9 && hour < 17;
}
