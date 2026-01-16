/**
 * Settings view - Application configuration
 */

import { useEffect, useState } from 'react';
import {
  Box,
  Card,
  CardContent,
  Typography,
  TextField,
  Switch,
  FormControlLabel,
  Button,
  Alert,
  Snackbar,
  Divider,
  Stack,
  Accordion,
  AccordionSummary,
  AccordionDetails,
  IconButton,
  Tooltip,
} from '@mui/material';
import {
  Save as SaveIcon,
  RestartAlt as ResetIcon,
  ExpandMore as ExpandMoreIcon,
  ContentCopy as CopyIcon,
} from '@mui/icons-material';

import { getConfig, updateConfig, resetConfig } from '../services/api';
import type { AppConfig } from '../types';
import LoadingSpinner from '../components/common/LoadingSpinner';
import ConfirmDialog from '../components/common/ConfirmDialog';

export default function Settings() {
  const [loading, setLoading] = useState(true);
  const [config, setConfig] = useState<AppConfig | null>(null);
  const [saving, setSaving] = useState(false);
  const [resetConfirmOpen, setResetConfirmOpen] = useState(false);
  const [snackbar, setSnackbar] = useState<{
    open: boolean;
    message: string;
    severity: 'success' | 'error';
  }>({ open: false, message: '', severity: 'success' });

  useEffect(() => {
    loadConfig();
  }, []);

  const loadConfig = async () => {
    setLoading(true);
    const result = await getConfig();
    if (result.success && result.data) {
      setConfig(result.data);
    }
    setLoading(false);
  };

  const handleSave = async () => {
    if (!config) return;

    setSaving(true);
    const result = await updateConfig(config);

    if (result.success) {
      setSnackbar({
        open: true,
        message: 'Settings saved successfully',
        severity: 'success',
      });
    } else {
      setSnackbar({
        open: true,
        message: result.error || 'Failed to save settings',
        severity: 'error',
      });
    }

    setSaving(false);
  };

  const handleReset = async () => {
    const result = await resetConfig();

    if (result.success && result.data) {
      setConfig(result.data);
      setSnackbar({
        open: true,
        message: 'Settings reset to defaults',
        severity: 'success',
      });
    } else {
      setSnackbar({
        open: true,
        message: result.error || 'Failed to reset settings',
        severity: 'error',
      });
    }

    setResetConfirmOpen(false);
  };

  const handleCopy = (text: string) => {
    navigator.clipboard.writeText(text);
    setSnackbar({
      open: true,
      message: 'Copied to clipboard',
      severity: 'success',
    });
  };

  if (loading || !config) {
    return <LoadingSpinner message="Loading settings..." />;
  }

  return (
    <Box>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 3 }}>
        <Typography variant="h4">Settings</Typography>
        <Box sx={{ display: 'flex', gap: 1 }}>
          <Button
            variant="outlined"
            startIcon={<ResetIcon />}
            onClick={() => setResetConfirmOpen(true)}
          >
            Reset
          </Button>
          <Button
            variant="contained"
            startIcon={<SaveIcon />}
            onClick={handleSave}
            disabled={saving}
          >
            {saving ? 'Saving...' : 'Save Changes'}
          </Button>
        </Box>
      </Box>

      <Stack spacing={3}>
        {/* Automation Schedule Section */}
        <Card>
          <CardContent>
            <Typography variant="h6" gutterBottom>
              Automation Schedule
            </Typography>
            <Divider sx={{ mb: 2 }} />

            <Stack spacing={3}>
              <FormControlLabel
                control={
                  <Switch
                    checked={config.schedule.enabled}
                    onChange={(e) =>
                      setConfig({
                        ...config,
                        schedule: {
                          ...config.schedule,
                          enabled: e.target.checked,
                        },
                      })
                    }
                  />
                }
                label="Enable Automation"
              />
              <Typography variant="body2" color="text.secondary" sx={{ mt: -2 }}>
                Automatically run detection and search cycles on schedule
              </Typography>

              <TextField
                label="Interval (hours)"
                type="number"
                value={config.schedule.intervalHours}
                onChange={(e) =>
                  setConfig({
                    ...config,
                    schedule: {
                      ...config.schedule,
                      intervalHours: parseInt(e.target.value) || 1,
                    },
                  })
                }
                inputProps={{ min: 1, max: 168 }}
                helperText="Time between automatic cycles (1-168 hours)"
                fullWidth
              />
            </Stack>
          </CardContent>
        </Card>

        {/* Search Limits Section */}
        <Card>
          <CardContent>
            <Typography variant="h6" gutterBottom>
              Search Limits
            </Typography>
            <Divider sx={{ mb: 2 }} />

            <Stack spacing={3}>
              <Box>
                <Typography variant="subtitle2" gutterBottom>
                  Missing Content
                </Typography>
                <Stack direction={{ xs: 'column', sm: 'row' }} spacing={2}>
                  <TextField
                    label="Movies"
                    type="number"
                    value={config.searchLimits.missingMoviesLimit}
                    onChange={(e) =>
                      setConfig({
                        ...config,
                        searchLimits: {
                          ...config.searchLimits,
                          missingMoviesLimit: parseInt(e.target.value) || 0,
                        },
                      })
                    }
                    inputProps={{ min: 0, max: 1000 }}
                    helperText="Max missing movies to search per cycle"
                    fullWidth
                  />
                  <TextField
                    label="Episodes"
                    type="number"
                    value={config.searchLimits.missingEpisodesLimit}
                    onChange={(e) =>
                      setConfig({
                        ...config,
                        searchLimits: {
                          ...config.searchLimits,
                          missingEpisodesLimit: parseInt(e.target.value) || 0,
                        },
                      })
                    }
                    inputProps={{ min: 0, max: 1000 }}
                    helperText="Max missing episodes to search per cycle"
                    fullWidth
                  />
                </Stack>
              </Box>

              <Box>
                <Typography variant="subtitle2" gutterBottom>
                  Quality Cutoff
                </Typography>
                <Stack direction={{ xs: 'column', sm: 'row' }} spacing={2}>
                  <TextField
                    label="Movies"
                    type="number"
                    value={config.searchLimits.cutoffMoviesLimit}
                    onChange={(e) =>
                      setConfig({
                        ...config,
                        searchLimits: {
                          ...config.searchLimits,
                          cutoffMoviesLimit: parseInt(e.target.value) || 0,
                        },
                      })
                    }
                    inputProps={{ min: 0, max: 1000 }}
                    helperText="Max cutoff movies to search per cycle"
                    fullWidth
                  />
                  <TextField
                    label="Episodes"
                    type="number"
                    value={config.searchLimits.cutoffEpisodesLimit}
                    onChange={(e) =>
                      setConfig({
                        ...config,
                        searchLimits: {
                          ...config.searchLimits,
                          cutoffEpisodesLimit: parseInt(e.target.value) || 0,
                        },
                      })
                    }
                    inputProps={{ min: 0, max: 1000 }}
                    helperText="Max cutoff episodes to search per cycle"
                    fullWidth
                  />
                </Stack>
              </Box>
            </Stack>
          </CardContent>
        </Card>

        {/* Advanced Section */}
        <Accordion>
          <AccordionSummary expandIcon={<ExpandMoreIcon />}>
            <Typography variant="h6">Advanced</Typography>
          </AccordionSummary>
          <AccordionDetails>
            <Stack spacing={2}>
              <Box>
                <Typography variant="subtitle2" gutterBottom>
                  API Base URL
                </Typography>
                <Box sx={{ display: 'flex', gap: 1 }}>
                  <TextField
                    value={`${window.location.origin}/api`}
                    fullWidth
                    InputProps={{
                      readOnly: true,
                    }}
                    size="small"
                  />
                  <Tooltip title="Copy to clipboard">
                    <IconButton
                      onClick={() => handleCopy(`${window.location.origin}/api`)}
                    >
                      <CopyIcon />
                    </IconButton>
                  </Tooltip>
                </Box>
                <Typography variant="caption" color="text.secondary">
                  For external integrations
                </Typography>
              </Box>

              <Alert severity="info">
                Database path and other advanced settings can be configured via
                environment variables. See documentation for details.
              </Alert>
            </Stack>
          </AccordionDetails>
        </Accordion>
      </Stack>

      {/* Reset Confirmation Dialog */}
      <ConfirmDialog
        open={resetConfirmOpen}
        title="Reset to Defaults"
        message="Are you sure you want to reset all settings to their default values? This action cannot be undone."
        confirmLabel="Reset"
        onConfirm={handleReset}
        onCancel={() => setResetConfirmOpen(false)}
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
