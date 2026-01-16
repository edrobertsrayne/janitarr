/**
 * Logs view - Activity logs with real-time streaming
 */

import { useEffect, useState, useRef } from 'react';
import {
  Box,
  Card,
  CardContent,
  Typography,
  TextField,
  Select,
  MenuItem,
  FormControl,
  InputLabel,
  Button,
  Chip,
  Alert,
  Snackbar,
  IconButton,
  Stack,
  Paper,
  List,
  ListItem,
  ListItemText,
  Divider,
} from '@mui/material';
import {
  Search as SearchIcon,
  Download as ExportIcon,
  Delete as ClearIcon,
  Refresh as RefreshIcon,
  Circle as ConnectionIcon,
} from '@mui/icons-material';
import { formatDistanceToNow } from 'date-fns';

import { getLogs, deleteLogs, exportLogs } from '../services/api';
import { WebSocketClient } from '../services/websocket';
import type { LogEntry, LogEntryType } from '../types';
import LoadingSpinner from '../components/common/LoadingSpinner';
import ConfirmDialog from '../components/common/ConfirmDialog';

export default function Logs() {
  const [logs, setLogs] = useState<LogEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const [wsClient] = useState(() => new WebSocketClient({
    onLog: (log) => {
      setLogs((prev) => [log, ...prev]);
      if (autoScroll && listRef.current) {
        listRef.current.scrollTop = 0;
      }
    },
    onStatusChange: (status) => setWsStatus(status),
  }));
  const [wsStatus, setWsStatus] = useState<'connecting' | 'connected' | 'disconnected' | 'error'>('disconnected');
  const [autoScroll] = useState(true);
  const [searchQuery, setSearchQuery] = useState('');
  const [typeFilter, setTypeFilter] = useState<LogEntryType | 'all'>('all');
  const [clearConfirmOpen, setClearConfirmOpen] = useState(false);
  const [snackbar, setSnackbar] = useState<{
    open: boolean;
    message: string;
    severity: 'success' | 'error';
  }>({ open: false, message: '', severity: 'success' });

  const listRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    loadLogs();
    wsClient.connect();
    wsClient.subscribe();

    return () => {
      wsClient.disconnect();
    };
  }, [wsClient]);

  const loadLogs = async () => {
    setLoading(true);
    const result = await getLogs({ limit: 100 });
    if (result.success && result.data) {
      setLogs(result.data);
    }
    setLoading(false);
  };

  const handleClearLogs = async () => {
    const result = await deleteLogs();

    if (result.success) {
      setLogs([]);
      setSnackbar({
        open: true,
        message: 'Logs cleared successfully',
        severity: 'success',
      });
    } else {
      setSnackbar({
        open: true,
        message: result.error || 'Failed to clear logs',
        severity: 'error',
      });
    }

    setClearConfirmOpen(false);
  };

  const handleExport = async (format: 'json' | 'csv') => {
    const blob = await exportLogs(undefined, format);

    if (blob) {
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `janitarr-logs-${Date.now()}.${format}`;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);

      setSnackbar({
        open: true,
        message: `Logs exported as ${format.toUpperCase()}`,
        severity: 'success',
      });
    } else {
      setSnackbar({
        open: true,
        message: 'Failed to export logs',
        severity: 'error',
      });
    }
  };

  const getLogIcon = (type: LogEntryType): string => {
    switch (type) {
      case 'cycle_start':
        return '▶️';
      case 'cycle_end':
        return '✅';
      case 'search':
        return '🔍';
      case 'error':
        return '❌';
      default:
        return '📝';
    }
  };

  const getLogColor = (type: LogEntryType): string => {
    switch (type) {
      case 'cycle_start':
        return '#2196f3';
      case 'cycle_end':
        return '#4caf50';
      case 'search':
        return '#00bcd4';
      case 'error':
        return '#f44336';
      default:
        return '#9e9e9e';
    }
  };

  const filteredLogs = logs.filter((log) => {
    if (typeFilter !== 'all' && log.type !== typeFilter) {
      return false;
    }

    if (searchQuery && !log.message.toLowerCase().includes(searchQuery.toLowerCase())) {
      return false;
    }

    return true;
  });

  if (loading) {
    return <LoadingSpinner message="Loading logs..." />;
  }

  return (
    <Box>
      {/* Toolbar */}
      <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 3 }}>
        <Typography variant="h4">Logs</Typography>
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
          <Chip
            icon={<ConnectionIcon />}
            label={wsStatus}
            size="small"
            color={
              wsStatus === 'connected'
                ? 'success'
                : wsStatus === 'connecting'
                ? 'warning'
                : 'default'
            }
          />
          <Button
            size="small"
            startIcon={<ExportIcon />}
            onClick={() => handleExport('json')}
          >
            JSON
          </Button>
          <Button
            size="small"
            startIcon={<ExportIcon />}
            onClick={() => handleExport('csv')}
          >
            CSV
          </Button>
          <IconButton onClick={loadLogs} title="Refresh">
            <RefreshIcon />
          </IconButton>
          <IconButton
            onClick={() => setClearConfirmOpen(true)}
            title="Clear All Logs"
            color="error"
          >
            <ClearIcon />
          </IconButton>
        </Box>
      </Box>

      {/* Filters */}
      <Card sx={{ mb: 3 }}>
        <CardContent>
          <Stack direction={{ xs: 'column', sm: 'row' }} spacing={2}>
            <TextField
              placeholder="Search logs..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              InputProps={{
                startAdornment: <SearchIcon sx={{ mr: 1, color: 'text.secondary' }} />,
              }}
              size="small"
              fullWidth
            />
            <FormControl size="small" sx={{ minWidth: 200 }}>
              <InputLabel>Type Filter</InputLabel>
              <Select
                value={typeFilter}
                label="Type Filter"
                onChange={(e) => setTypeFilter(e.target.value as LogEntryType | 'all')}
              >
                <MenuItem value="all">All Types</MenuItem>
                <MenuItem value="cycle_start">Cycle Start</MenuItem>
                <MenuItem value="cycle_end">Cycle End</MenuItem>
                <MenuItem value="search">Search</MenuItem>
                <MenuItem value="error">Error</MenuItem>
              </Select>
            </FormControl>
          </Stack>
          <Typography variant="caption" color="text.secondary" sx={{ mt: 1, display: 'block' }}>
            {filteredLogs.length} of {logs.length} logs
          </Typography>
        </CardContent>
      </Card>

      {/* Log List */}
      {filteredLogs.length === 0 ? (
        <Alert severity="info">No logs found</Alert>
      ) : (
        <Paper
          ref={listRef}
          sx={{
            maxHeight: '70vh',
            overflow: 'auto',
          }}
        >
          <List>
            {filteredLogs.map((log, index) => (
              <Box key={log.id}>
                <ListItem
                  sx={{
                    borderLeft: `4px solid ${getLogColor(log.type)}`,
                    '&:hover': {
                      bgcolor: 'action.hover',
                    },
                  }}
                >
                  <ListItemText
                    primary={
                      <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                        <span>{getLogIcon(log.type)}</span>
                        <Chip label={log.type} size="small" />
                        {log.serverName && (
                          <Chip
                            label={log.serverName}
                            size="small"
                            variant="outlined"
                          />
                        )}
                        {log.isManual && (
                          <Chip label="Manual" size="small" color="warning" />
                        )}
                        <Typography variant="caption" color="text.secondary">
                          {formatDistanceToNow(new Date(log.timestamp), {
                            addSuffix: true,
                          })}
                        </Typography>
                      </Box>
                    }
                    secondary={
                      <Box>
                        <Typography variant="body2">{log.message}</Typography>
                        {log.count && (
                          <Typography variant="caption" color="text.secondary">
                            {log.count} items • {log.category}
                          </Typography>
                        )}
                      </Box>
                    }
                  />
                </ListItem>
                {index < filteredLogs.length - 1 && <Divider />}
              </Box>
            ))}
          </List>
        </Paper>
      )}

      {/* Clear Confirmation Dialog */}
      <ConfirmDialog
        open={clearConfirmOpen}
        title="Clear All Logs"
        message="Are you sure you want to clear all logs? This action cannot be undone."
        confirmLabel="Clear"
        onConfirm={handleClearLogs}
        onCancel={() => setClearConfirmOpen(false)}
        destructive
      />

      {/* Snackbar for notifications */}
      <Snackbar
        open={snackbar.open}
        autoHideDuration={6000}
        onClose={() => setSnackbar({ ...snackbar, open: false })}
        anchorOrigin={{ vertical: 'bottom', horizontal: 'right' }}
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
