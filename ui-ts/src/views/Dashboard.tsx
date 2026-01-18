/**
 * Dashboard view - Overview of system status
 */

import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import {
  Box,
  Grid,
  Card,
  CardContent,
  Typography,
  Button,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  IconButton,
  Chip,
  Alert,
  Snackbar,
} from "@mui/material";
import {
  Timeline,
  TimelineItem,
  TimelineSeparator,
  TimelineConnector,
  TimelineContent,
  TimelineDot,
  TimelineOppositeContent,
} from "@mui/lab";
import {
  PlayArrow as PlayIcon,
  Add as AddIcon,
  Storage as StorageIcon,
  Schedule as ScheduleIcon,
  Search as SearchIcon,
  Error as ErrorIcon,
  CheckCircle as SuccessIcon,
  Edit as EditIcon,
  Science as TestIcon,
} from "@mui/icons-material";
import { formatDistanceToNow } from "date-fns";

import {
  getStatsSummary,
  getServers,
  getLogs,
  triggerAutomation,
} from "../services/api";
import type {
  StatsSummaryResponse,
  ServerConfig,
  LogEntry,
  LogEntryType,
} from "../types";
import LoadingSpinner from "../components/common/LoadingSpinner";
import StatusBadge from "../components/common/StatusBadge";

export default function Dashboard() {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(true);
  const [stats, setStats] = useState<StatsSummaryResponse | null>(null);
  const [servers, setServers] = useState<ServerConfig[]>([]);
  const [recentLogs, setRecentLogs] = useState<LogEntry[]>([]);
  const [triggering, setTriggering] = useState(false);
  const [snackbar, setSnackbar] = useState<{
    open: boolean;
    message: string;
    severity: "success" | "error";
  }>({ open: false, message: "", severity: "success" });

  useEffect(() => {
    loadDashboardData();
    // Refresh every 60 seconds
    const interval = setInterval(loadDashboardData, 60000);
    return () => clearInterval(interval);
  }, []);

  const loadDashboardData = async () => {
    setLoading(true);
    const [statsRes, serversRes, logsRes] = await Promise.all([
      getStatsSummary(),
      getServers(),
      getLogs({ limit: 10 }),
    ]);

    if (statsRes.success && statsRes.data) {
      setStats(statsRes.data);
    }

    if (serversRes.success && serversRes.data) {
      setServers(serversRes.data);
    }

    if (logsRes.success && logsRes.data) {
      setRecentLogs(logsRes.data);
    }

    setLoading(false);
  };

  const handleTriggerAutomation = async () => {
    setTriggering(true);
    const result = await triggerAutomation();

    if (result.success) {
      setSnackbar({
        open: true,
        message: "Automation triggered successfully",
        severity: "success",
      });
      // Refresh data after a short delay
      setTimeout(loadDashboardData, 2000);
    } else {
      setSnackbar({
        open: true,
        message: result.error || "Failed to trigger automation",
        severity: "error",
      });
    }

    setTriggering(false);
  };

  const getLogIcon = (type: LogEntryType) => {
    switch (type) {
      case "cycle_start":
        return <PlayIcon />;
      case "cycle_end":
        return <SuccessIcon />;
      case "search":
        return <SearchIcon />;
      case "error":
        return <ErrorIcon />;
      default:
        return <SearchIcon />;
    }
  };

  const getLogColor = (
    type: LogEntryType,
  ): "primary" | "success" | "info" | "error" => {
    switch (type) {
      case "cycle_start":
        return "primary";
      case "cycle_end":
        return "success";
      case "search":
        return "info";
      case "error":
        return "error";
      default:
        return "info";
    }
  };

  if (loading && !stats) {
    return <LoadingSpinner message="Loading dashboard..." />;
  }

  return (
    <Box>
      <Box
        sx={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "flex-start",
          mb: 3,
          flexWrap: "wrap",
          gap: 2,
        }}
      >
        <Typography variant="h4" gutterBottom sx={{ mb: 0 }}>
          Dashboard
        </Typography>
        <Box sx={{ display: "flex", gap: 1, flexWrap: "wrap" }}>
          <Button
            variant="contained"
            startIcon={<PlayIcon />}
            onClick={handleTriggerAutomation}
            disabled={triggering}
            sx={{ minWidth: { xs: "110px", sm: "auto" } }}
          >
            {triggering ? "Running..." : "Run Now"}
          </Button>
          <Button
            variant="outlined"
            startIcon={<AddIcon />}
            onClick={() => navigate("/servers")}
            sx={{ minWidth: { xs: "120px", sm: "auto" } }}
          >
            Add Server
          </Button>
        </Box>
      </Box>

      {/* Status Cards */}
      <Grid
        container
        spacing={3}
        sx={{ mb: 4 }}
        role="region"
        aria-label="System statistics"
      >
        <Grid size={{ xs: 12, sm: 6, md: 3 }}>
          <Card>
            <CardContent>
              <Box sx={{ display: "flex", alignItems: "center", mb: 1 }}>
                <StorageIcon
                  color="primary"
                  sx={{ mr: 1 }}
                  aria-hidden="true"
                />
                <Typography color="text.secondary" variant="body2">
                  Total Servers
                </Typography>
              </Box>
              <Typography variant="h4">{stats?.totalServers || 0}</Typography>
              <Typography variant="body2" color="text.secondary">
                {stats?.activeServers || 0} active
              </Typography>
            </CardContent>
          </Card>
        </Grid>

        <Grid size={{ xs: 12, sm: 6, md: 3 }}>
          <Card>
            <CardContent>
              <Box sx={{ display: "flex", alignItems: "center", mb: 1 }}>
                <ScheduleIcon
                  color="primary"
                  sx={{ mr: 1 }}
                  aria-hidden="true"
                />
                <Typography color="text.secondary" variant="body2">
                  Last Cycle
                </Typography>
              </Box>
              <Typography variant="h6">
                {stats?.lastCycleTime
                  ? formatDistanceToNow(new Date(stats.lastCycleTime), {
                      addSuffix: true,
                    })
                  : "Never"}
              </Typography>
              <Typography variant="body2" color="text.secondary">
                Next:{" "}
                {stats?.nextScheduledTime
                  ? formatDistanceToNow(new Date(stats.nextScheduledTime), {
                      addSuffix: true,
                    })
                  : "N/A"}
              </Typography>
            </CardContent>
          </Card>
        </Grid>

        <Grid size={{ xs: 12, sm: 6, md: 3 }}>
          <Card>
            <CardContent>
              <Box sx={{ display: "flex", alignItems: "center", mb: 1 }}>
                <SearchIcon color="primary" sx={{ mr: 1 }} aria-hidden="true" />
                <Typography color="text.secondary" variant="body2">
                  Recent Searches
                </Typography>
              </Box>
              <Typography variant="h4">
                {stats?.searchesLast24h || 0}
              </Typography>
              <Typography variant="body2" color="text.secondary">
                Last 24 hours
              </Typography>
            </CardContent>
          </Card>
        </Grid>

        <Grid size={{ xs: 12, sm: 6, md: 3 }}>
          <Card>
            <CardContent>
              <Box sx={{ display: "flex", alignItems: "center", mb: 1 }}>
                <ErrorIcon color="error" sx={{ mr: 1 }} aria-hidden="true" />
                <Typography color="text.secondary" variant="body2">
                  Errors
                </Typography>
              </Box>
              <Typography
                variant="h4"
                color={stats?.errorsLast24h ? "error" : "inherit"}
              >
                {stats?.errorsLast24h || 0}
              </Typography>
              <Typography variant="body2" color="text.secondary">
                Last 24 hours
              </Typography>
            </CardContent>
          </Card>
        </Grid>
      </Grid>

      {/* Server Status List */}
      <Card sx={{ mb: 4 }}>
        <CardContent>
          <Box sx={{ display: "flex", justifyContent: "space-between", mb: 2 }}>
            <Typography variant="h6" component="h2">
              Servers
            </Typography>
            <Button size="small" onClick={() => navigate("/servers")}>
              View All
            </Button>
          </Box>
          {servers.length === 0 ? (
            <Alert severity="info">
              No servers configured. Click "Add Server" to get started.
            </Alert>
          ) : (
            <TableContainer>
              <Table size="small" aria-label="Server status">
                <TableHead
                  sx={{ display: { xs: "none", sm: "table-header-group" } }}
                >
                  <TableRow>
                    <TableCell>Name</TableCell>
                    <TableCell>Type</TableCell>
                    <TableCell>Status</TableCell>
                    <TableCell align="right">Actions</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {servers.map((server) => (
                    <TableRow key={server.id}>
                      <TableCell>
                        <Box>
                          <Typography variant="body2" sx={{ fontWeight: 500 }}>
                            {server.name}
                          </Typography>
                          <Box
                            sx={{
                              display: { xs: "flex", sm: "none" },
                              gap: 1,
                              mt: 0.5,
                              alignItems: "center",
                              flexWrap: "wrap",
                            }}
                          >
                            <Chip
                              label={server.type.toUpperCase()}
                              size="small"
                              color={
                                server.type === "radarr"
                                  ? "primary"
                                  : "secondary"
                              }
                            />
                            <StatusBadge
                              status={
                                server.enabled === false
                                  ? "disabled"
                                  : "connected"
                              }
                            />
                          </Box>
                        </Box>
                      </TableCell>
                      <TableCell
                        sx={{ display: { xs: "none", sm: "table-cell" } }}
                      >
                        <Chip
                          label={server.type.toUpperCase()}
                          size="small"
                          color={
                            server.type === "radarr" ? "primary" : "secondary"
                          }
                        />
                      </TableCell>
                      <TableCell
                        sx={{ display: { xs: "none", sm: "table-cell" } }}
                      >
                        <StatusBadge
                          status={
                            server.enabled === false ? "disabled" : "connected"
                          }
                        />
                      </TableCell>
                      <TableCell align="right">
                        <IconButton
                          size="small"
                          onClick={() => navigate("/servers")}
                          title="Edit"
                          aria-label={`Edit ${server.name}`}
                          sx={{ minWidth: 44, minHeight: 44 }}
                        >
                          <EditIcon />
                        </IconButton>
                        <IconButton
                          size="small"
                          onClick={() => navigate("/servers")}
                          title="Test"
                          aria-label={`Test ${server.name}`}
                          sx={{ minWidth: 44, minHeight: 44 }}
                        >
                          <TestIcon />
                        </IconButton>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </TableContainer>
          )}
        </CardContent>
      </Card>

      {/* Recent Activity Timeline */}
      <Card>
        <CardContent>
          <Box sx={{ display: "flex", justifyContent: "space-between", mb: 2 }}>
            <Typography variant="h6" component="h2">
              Recent Activity
            </Typography>
            <Button size="small" onClick={() => navigate("/logs")}>
              View All Logs
            </Button>
          </Box>
          {recentLogs.length === 0 ? (
            <Alert severity="info">No recent activity</Alert>
          ) : (
            <Timeline>
              {recentLogs.map((log, index) => (
                <TimelineItem key={log.id}>
                  <TimelineOppositeContent
                    color="text.secondary"
                    sx={{ flex: 0.3 }}
                  >
                    <Typography variant="caption">
                      {formatDistanceToNow(new Date(log.timestamp), {
                        addSuffix: true,
                      })}
                    </Typography>
                  </TimelineOppositeContent>
                  <TimelineSeparator>
                    <TimelineDot color={getLogColor(log.type)}>
                      {getLogIcon(log.type)}
                    </TimelineDot>
                    {index < recentLogs.length - 1 && <TimelineConnector />}
                  </TimelineSeparator>
                  <TimelineContent>
                    <Typography variant="body2">{log.message}</Typography>
                    {log.serverName && (
                      <Typography variant="caption" color="text.secondary">
                        {log.serverName}
                        {log.count && ` • ${log.count} items`}
                      </Typography>
                    )}
                  </TimelineContent>
                </TimelineItem>
              ))}
            </Timeline>
          )}
        </CardContent>
      </Card>

      {/* Snackbar for notifications */}
      <Snackbar
        open={snackbar.open}
        autoHideDuration={6000}
        onClose={() => setSnackbar({ ...snackbar, open: false })}
        anchorOrigin={{ vertical: "bottom", horizontal: "right" }}
      >
        <Alert
          severity={snackbar.severity}
          onClose={() => setSnackbar({ ...snackbar, open: false })}
        >
          {snackbar.message}
        </Alert>
      </Snackbar>
    </Box>
  );
}
