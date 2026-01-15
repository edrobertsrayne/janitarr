/**
 * Scheduler
 *
 * Manages periodic execution of automation cycles with configurable intervals.
 * Prevents concurrent execution and tracks next run time.
 */

import { getDatabase } from "../storage/database";

/** Minimum interval in hours */
const MIN_INTERVAL_HOURS = 1;

/** Scheduler state */
interface SchedulerState {
  isRunning: boolean;
  isCycleActive: boolean;
  nextRunTime: Date | null;
  timeoutId: Timer | null;
}

/** Singleton state */
const state: SchedulerState = {
  isRunning: false,
  isCycleActive: false,
  nextRunTime: null,
  timeoutId: null,
};

/** Callback function type for automation cycle */
type CycleCallback = (isManual: boolean) => Promise<void>;

/** Registered callback for cycle execution */
let cycleCallback: CycleCallback | null = null;

/**
 * Register the automation cycle callback
 */
export function registerCycleCallback(callback: CycleCallback): void {
  cycleCallback = callback;
}

/**
 * Get current scheduler configuration
 */
export function getScheduleConfig(): {
  intervalHours: number;
  enabled: boolean;
} {
  const db = getDatabase();
  const config = db.getAppConfig();
  return config.schedule;
}

/**
 * Update scheduler configuration
 */
export function setScheduleConfig(
  intervalHours?: number,
  enabled?: boolean
): void {
  const db = getDatabase();

  // Validate interval if provided
  if (intervalHours !== undefined && intervalHours < MIN_INTERVAL_HOURS) {
    throw new Error(
      `Interval must be at least ${MIN_INTERVAL_HOURS} hour(s)`
    );
  }

  db.setAppConfig({
    schedule: {
      intervalHours,
      enabled,
    },
  });

  // If scheduler is running, restart it to apply new interval
  if (state.isRunning && enabled !== false) {
    stop();
    start();
  } else if (state.isRunning && enabled === false) {
    stop();
  } else if (!state.isRunning && enabled === true) {
    start();
  }
}

/**
 * Calculate next run time based on current config
 */
function calculateNextRunTime(): Date {
  const config = getScheduleConfig();
  const now = new Date();
  const nextRun = new Date(now.getTime() + config.intervalHours * 60 * 60 * 1000);
  return nextRun;
}

/**
 * Schedule the next cycle execution
 */
function scheduleNext(): void {
  if (state.timeoutId) {
    clearTimeout(state.timeoutId);
    state.timeoutId = null;
  }

  const config = getScheduleConfig();
  if (!config.enabled) {
    state.nextRunTime = null;
    return;
  }

  const nextRun = calculateNextRunTime();
  state.nextRunTime = nextRun;

  const delay = nextRun.getTime() - Date.now();

  state.timeoutId = setTimeout(() => {
    executeCycle(false);
  }, delay);
}

/**
 * Execute the automation cycle (internal)
 */
async function executeCycle(isManual: boolean): Promise<void> {
  if (!cycleCallback) {
    console.error("No cycle callback registered");
    return;
  }

  // Prevent concurrent execution
  if (state.isCycleActive) {
    console.warn("Cycle already in progress, skipping");
    return;
  }

  state.isCycleActive = true;

  try {
    await cycleCallback(isManual);
  } catch (error) {
    console.error("Cycle execution failed:", error);
  } finally {
    state.isCycleActive = false;

    // Schedule next run if this was a scheduled (not manual) cycle
    if (!isManual && state.isRunning) {
      scheduleNext();
    }
  }
}

/**
 * Start the scheduler
 * Runs the first cycle immediately, then schedules subsequent cycles
 */
export async function start(): Promise<void> {
  if (state.isRunning) {
    console.warn("Scheduler already running");
    return;
  }

  const config = getScheduleConfig();
  if (!config.enabled) {
    console.log("Scheduler is disabled in configuration");
    return;
  }

  state.isRunning = true;

  // Run first cycle immediately
  await executeCycle(false);

  // Schedule next cycle (if still running after first cycle completes)
  if (state.isRunning) {
    scheduleNext();
  }
}

/**
 * Stop the scheduler
 * Does not interrupt a running cycle
 */
export function stop(): void {
  if (!state.isRunning) {
    return;
  }

  state.isRunning = false;

  if (state.timeoutId) {
    clearTimeout(state.timeoutId);
    state.timeoutId = null;
  }

  state.nextRunTime = null;
}

/**
 * Manually trigger a cycle
 * Does not affect the regular schedule
 */
export async function triggerManual(): Promise<void> {
  if (!cycleCallback) {
    throw new Error("No cycle callback registered");
  }

  // Check if a cycle is already running
  if (state.isCycleActive) {
    throw new Error("A cycle is already in progress");
  }

  await executeCycle(true);
}

/**
 * Get current scheduler status
 */
export function getStatus(): {
  isRunning: boolean;
  isCycleActive: boolean;
  nextRunTime: Date | null;
  config: {
    intervalHours: number;
    enabled: boolean;
  };
} {
  return {
    isRunning: state.isRunning,
    isCycleActive: state.isCycleActive,
    nextRunTime: state.nextRunTime,
    config: getScheduleConfig(),
  };
}

/**
 * Get time remaining until next scheduled run (in milliseconds)
 * Returns null if scheduler is not running or no run is scheduled
 */
export function getTimeUntilNextRun(): number | null {
  if (!state.isRunning || !state.nextRunTime) {
    return null;
  }

  const remaining = state.nextRunTime.getTime() - Date.now();
  return Math.max(0, remaining);
}

/**
 * Check if scheduler is currently running
 */
export function isRunning(): boolean {
  return state.isRunning;
}

/**
 * Check if a cycle is currently active
 */
export function isCycleActive(): boolean {
  return state.isCycleActive;
}
