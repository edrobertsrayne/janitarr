/**
 * Servers view - Manage Radarr/Sonarr servers
 */

import { useEffect, useState } from "react";
import {
  Box,
  Button,
  Card,
  CardContent,
  CardActions,
  Grid,
  Typography,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Paper,
  IconButton,
  Chip,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  FormControl,
  FormLabel,
  RadioGroup,
  FormControlLabel,
  Radio,
  Switch,
  Alert,
  Snackbar,
  ToggleButton,
  ToggleButtonGroup,
  CircularProgress,
} from "@mui/material";
import {
  Add as AddIcon,
  Edit as EditIcon,
  Delete as DeleteIcon,
  ViewList as ListIcon,
  ViewModule as CardIcon,
  Science as TestIcon,
  Storage as StorageIcon,
} from "@mui/icons-material";

import {
  getServers,
  createServer,
  updateServer,
  deleteServer,
  testServer, // For testing new server configurations (CreateServerRequest)
  testServerConnectionById, // For testing existing servers by ID
} from "../services/api";
import type { ServerConfig, CreateServerRequest, ServerType } from "../types";
import LoadingSpinner from "../components/common/LoadingSpinner";
import StatusBadge from "../components/common/StatusBadge";
import ConfirmDialog from "../components/common/ConfirmDialog";

type ViewMode = "list" | "card";

export default function Servers() {
  const [servers, setServers] = useState<ServerConfig[]>([]);
  const [loading, setLoading] = useState(true);
  const [viewMode, setViewMode] = useState<ViewMode>("list");
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingServer, setEditingServer] = useState<ServerConfig | null>(null);
  const [deleteConfirmOpen, setDeleteConfirmOpen] = useState(false);
  const [serverToDelete, setServerToDelete] = useState<ServerConfig | null>(
    null,
  );
  const [testing, setTesting] = useState<string | null>(null);
  const [snackbar, setSnackbar] = useState<{
    open: boolean;
    message: string;
    severity: "success" | "error";
  }>({ open: false, message: "", severity: "success" });

  // Form state
  const [formData, setFormData] = useState<CreateServerRequest>({
    name: "",
    type: "radarr",
    url: "",
    apiKey: "",
    enabled: true,
  });
  const [formErrors, setFormErrors] = useState<Record<string, string>>({});

  useEffect(() => {
    loadServers();
  }, []);

  const loadServers = async () => {
    setLoading(true);
    const result = await getServers();
    if (result.success && result.data) {
      setServers(result.data);
    }
    setLoading(false);
  };

  const handleOpenDialog = (server?: ServerConfig) => {
    if (server) {
      setEditingServer(server);
      setFormData({
        name: server.name,
        type: server.type,
        url: server.url,
        apiKey: server.apiKey,
        enabled: server.enabled !== false,
      });
    } else {
      setEditingServer(null);
      setFormData({
        name: "",
        type: "radarr",
        url: "",
        apiKey: "",
        enabled: true,
      });
    }
    setFormErrors({});
    setDialogOpen(true);
  };

  const handleCloseDialog = () => {
    setDialogOpen(false);
    setEditingServer(null);
    setFormData({
      name: "",
      type: "radarr",
      url: "",
      apiKey: "",
      enabled: true,
    });
    setFormErrors({});
  };

  const validateForm = (): boolean => {
    const errors: Record<string, string> = {};

    if (!formData.name.trim()) {
      errors.name = "Name is required";
    }

    if (!formData.url.trim()) {
      errors.url = "URL is required";
    } else if (!/^https?:\/\/.+/.test(formData.url)) {
      errors.url = "URL must start with http:// or https://";
    }

    if (!formData.apiKey.trim()) {
      errors.apiKey = "API Key is required";
    }

    setFormErrors(errors);
    return Object.keys(errors).length === 0;
  };

  const handleSubmit = async () => {
    if (!validateForm()) {
      return;
    }

    let result;
    if (editingServer) {
      result = await updateServer(editingServer.id, formData);
    } else {
      result = await createServer(formData);
    }

    if (result.success) {
      setSnackbar({
        open: true,
        message: editingServer
          ? "Server updated successfully"
          : "Server created successfully",
        severity: "success",
      });
      handleCloseDialog();
      loadServers();
    } else {
      setSnackbar({
        open: true,
        message: result.error || "Failed to save server",
        severity: "error",
      });
    }
  };

  const handleTestConnection = async (server?: ServerConfig) => {
    setTesting(server ? server.id : "temp"); // Set testing state

    let testResult;
    if (server) {
      // Testing an existing server from the list
      testResult = await testServerConnectionById(server.id);
    } else {
      // Testing a new server from the dialog
      if (!validateForm()) {
        setTesting(null);
        return;
      }
      testResult = await testServer(formData);
    }

    setTesting(null); // Clear testing state

    if (testResult.success) {
      setSnackbar({
        open: true,
        message: testResult.data?.message || "Connection successful",
        severity: "success",
      });
    } else {
      setSnackbar({
        open: true,
        message: testResult.error || "Connection failed",
        severity: "error",
      });
    }
  };

  const handleDelete = async () => {
    if (!serverToDelete) return;

    const result = await deleteServer(serverToDelete.id);

    if (result.success) {
      setSnackbar({
        open: true,
        message: "Server deleted successfully",
        severity: "success",
      });
      loadServers();
    } else {
      setSnackbar({
        open: true,
        message: result.error || "Failed to delete server",
        severity: "error",
      });
    }

    setDeleteConfirmOpen(false);
    setServerToDelete(null);
  };

  if (loading) {
    return <LoadingSpinner message="Loading servers..." />;
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
        <Typography variant="h4" sx={{ mb: 0 }}>
          Servers
        </Typography>
        <Box sx={{ display: "flex", gap: 1, flexWrap: "wrap" }}>
          <ToggleButtonGroup
            value={viewMode}
            exclusive
            onChange={(_, value) => value && setViewMode(value)}
            size="small"
            aria-label="View mode"
          >
            <ToggleButton
              value="list"
              aria-label="List view"
              sx={{ minWidth: 44, minHeight: 44 }}
            >
              <ListIcon />
            </ToggleButton>
            <ToggleButton
              value="card"
              aria-label="Card view"
              sx={{ minWidth: 44, minHeight: 44 }}
            >
              <CardIcon />
            </ToggleButton>
          </ToggleButtonGroup>
          <Button
            variant="contained"
            startIcon={
              <AddIcon sx={{ display: { xs: "none", sm: "inline-flex" } }} />
            }
            onClick={() => handleOpenDialog()}
            sx={{ minWidth: { xs: "100px", sm: "auto" } }}
          >
            Add Server
          </Button>
        </Box>
      </Box>

      {servers.length === 0 ? (
        <Alert severity="info">
          No servers configured. Click "Add Server" to get started.
        </Alert>
      ) : viewMode === "list" ? (
        <TableContainer component={Paper}>
          <Table size="small" aria-label="Configured servers">
            <TableHead
              sx={{ display: { xs: "none", md: "table-header-group" } }}
            >
              <TableRow>
                <TableCell>Name</TableCell>
                <TableCell>Type</TableCell>
                <TableCell>URL</TableCell>
                <TableCell>Status</TableCell>
                <TableCell align="right">Actions</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {servers.map((server) => (
                <TableRow key={server.id}>
                  <TableCell>
                    <Box>
                      <Typography variant="body1" sx={{ fontWeight: 500 }}>
                        {server.name}
                      </Typography>
                      <Box
                        sx={{ display: { xs: "block", md: "none" }, mt: 0.5 }}
                      >
                        <Box
                          sx={{
                            display: "flex",
                            gap: 1,
                            alignItems: "center",
                            flexWrap: "wrap",
                            mb: 0.5,
                          }}
                        >
                          <Chip
                            label={server.type.toUpperCase()}
                            size="small"
                            color={
                              server.type === "radarr" ? "primary" : "secondary"
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
                        <Typography
                          variant="caption"
                          color="text.secondary"
                          sx={{ wordBreak: "break-all" }}
                        >
                          {server.url}
                        </Typography>
                      </Box>
                    </Box>
                  </TableCell>
                  <TableCell sx={{ display: { xs: "none", md: "table-cell" } }}>
                    <Chip
                      label={server.type.toUpperCase()}
                      size="small"
                      color={server.type === "radarr" ? "primary" : "secondary"}
                    />
                  </TableCell>
                  <TableCell sx={{ display: { xs: "none", md: "table-cell" } }}>
                    <Typography
                      variant="body2"
                      color="text.secondary"
                      sx={{
                        maxWidth: 200,
                        overflow: "hidden",
                        textOverflow: "ellipsis",
                      }}
                    >
                      {server.url}
                    </Typography>
                  </TableCell>
                  <TableCell sx={{ display: { xs: "none", md: "table-cell" } }}>
                    <StatusBadge
                      status={
                        server.enabled === false ? "disabled" : "connected"
                      }
                    />
                  </TableCell>
                  <TableCell align="right">
                    <Box
                      sx={{
                        display: "flex",
                        justifyContent: "flex-end",
                        gap: 0.5,
                      }}
                    >
                      <IconButton
                        size="small"
                        onClick={() => handleTestConnection(server)}
                        disabled={testing === server.id}
                        title="Test Connection"
                        aria-label={`Test connection to ${server.name}`}
                        sx={{ minWidth: 44, minHeight: 44 }}
                      >
                        {testing === server.id ? (
                          <CircularProgress size={20} />
                        ) : (
                          <TestIcon />
                        )}
                      </IconButton>
                      <IconButton
                        size="small"
                        onClick={() => handleOpenDialog(server)}
                        title="Edit"
                        aria-label={`Edit ${server.name}`}
                        sx={{ minWidth: 44, minHeight: 44 }}
                      >
                        <EditIcon />
                      </IconButton>
                      <IconButton
                        size="small"
                        onClick={() => {
                          setServerToDelete(server);
                          setDeleteConfirmOpen(true);
                        }}
                        title="Delete"
                        aria-label={`Delete ${server.name}`}
                        color="error"
                        sx={{ minWidth: 44, minHeight: 44 }}
                      >
                        <DeleteIcon />
                      </IconButton>
                    </Box>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </TableContainer>
      ) : (
        <Grid container spacing={3}>
          {servers.map((server) => (
            <Grid size={{ xs: 12, sm: 6, md: 4 }} key={server.id}>
              <Card>
                <CardContent>
                  <Box
                    sx={{
                      display: "flex",
                      justifyContent: "space-between",
                      mb: 2,
                    }}
                  >
                    <Typography variant="h6">{server.name}</Typography>
                    <StatusBadge
                      status={
                        server.enabled === false ? "disabled" : "connected"
                      }
                    />
                  </Box>
                  <Box sx={{ display: "flex", alignItems: "center", mb: 1 }}>
                    <StorageIcon
                      sx={{ mr: 1, color: "text.secondary" }}
                      aria-hidden="true"
                    />
                    <Chip
                      label={server.type.toUpperCase()}
                      size="small"
                      color={server.type === "radarr" ? "primary" : "secondary"}
                    />
                  </Box>
                  <Typography variant="body2" color="text.secondary" noWrap>
                    {server.url}
                  </Typography>
                </CardContent>
                <CardActions sx={{ justifyContent: "space-between" }}>
                  <Box sx={{ display: "flex", gap: 0.5 }}>
                    <IconButton
                      size="small"
                      onClick={() => handleTestConnection(server)}
                      disabled={testing === server.id}
                      title="Test"
                      aria-label={`Test connection to ${server.name}`}
                      sx={{ minWidth: 44, minHeight: 44 }}
                    >
                      {testing === server.id ? (
                        <CircularProgress size={20} />
                      ) : (
                        <TestIcon />
                      )}
                    </IconButton>
                    <IconButton
                      size="small"
                      onClick={() => handleOpenDialog(server)}
                      title="Edit"
                      aria-label={`Edit ${server.name}`}
                      sx={{ minWidth: 44, minHeight: 44 }}
                    >
                      <EditIcon />
                    </IconButton>
                  </Box>
                  <IconButton
                    size="small"
                    onClick={() => {
                      setServerToDelete(server);
                      setDeleteConfirmOpen(true);
                    }}
                    title="Delete"
                    aria-label={`Delete ${server.name}`}
                    color="error"
                    sx={{ minWidth: 44, minHeight: 44 }}
                  >
                    <DeleteIcon />
                  </IconButton>
                </CardActions>
              </Card>
            </Grid>
          ))}
        </Grid>
      )}

      {/* Add/Edit Server Dialog */}
      <Dialog
        open={dialogOpen}
        onClose={handleCloseDialog}
        maxWidth="sm"
        fullWidth
      >
        <DialogTitle>
          {editingServer ? "Edit Server" : "Add Server"}
        </DialogTitle>
        <DialogContent>
          <Box sx={{ display: "flex", flexDirection: "column", gap: 2, mt: 1 }}>
            <TextField
              label="Name"
              value={formData.name}
              onChange={(e) =>
                setFormData({ ...formData, name: e.target.value })
              }
              error={!!formErrors.name}
              helperText={formErrors.name}
              fullWidth
              required
            />

            <FormControl>
              <FormLabel>Type</FormLabel>
              <RadioGroup
                row
                value={formData.type}
                onChange={(e) =>
                  setFormData({
                    ...formData,
                    type: e.target.value as ServerType,
                  })
                }
              >
                <FormControlLabel
                  value="radarr"
                  control={<Radio />}
                  label="Radarr"
                />
                <FormControlLabel
                  value="sonarr"
                  control={<Radio />}
                  label="Sonarr"
                />
              </RadioGroup>
            </FormControl>

            <TextField
              label="URL"
              value={formData.url}
              onChange={(e) =>
                setFormData({ ...formData, url: e.target.value })
              }
              error={!!formErrors.url}
              helperText={formErrors.url || "Example: http://localhost:7878"}
              fullWidth
              required
            />

            <TextField
              label="API Key"
              type="password"
              value={formData.apiKey}
              onChange={(e) =>
                setFormData({ ...formData, apiKey: e.target.value })
              }
              error={!!formErrors.apiKey}
              helperText={formErrors.apiKey}
              fullWidth
              required
            />

            <FormControlLabel
              control={
                <Switch
                  checked={formData.enabled}
                  onChange={(e) =>
                    setFormData({ ...formData, enabled: e.target.checked })
                  }
                />
              }
              label="Enabled"
            />

            <Button
              variant="outlined"
              startIcon={
                testing ? <CircularProgress size={20} /> : <TestIcon />
              }
              onClick={() => handleTestConnection()}
              disabled={testing !== null}
              fullWidth
            >
              {testing ? "Testing..." : "Test Connection"}
            </Button>
          </Box>
        </DialogContent>
        <DialogActions>
          <Button onClick={handleCloseDialog}>Cancel</Button>
          <Button onClick={handleSubmit} variant="contained">
            {editingServer ? "Update" : "Create"}
          </Button>
        </DialogActions>
      </Dialog>

      {/* Delete Confirmation Dialog */}
      <ConfirmDialog
        open={deleteConfirmOpen}
        title="Delete Server"
        message={`Are you sure you want to delete "${serverToDelete?.name}"? This action cannot be undone.`}
        confirmLabel="Delete"
        onConfirm={handleDelete}
        onCancel={() => {
          setDeleteConfirmOpen(false);
          setServerToDelete(null);
        }}
        destructive
      />

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
