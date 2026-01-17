/**
 * Status badge component for servers and automation
 */

import { Chip } from '@mui/material';
import {
  CheckCircle as ConnectedIcon,
  Error as ErrorIcon,
  RemoveCircle as DisabledIcon,
} from '@mui/icons-material';

type StatusType = 'connected' | 'error' | 'disabled' | 'running' | 'stopped';

interface StatusBadgeProps {
  status: StatusType;
  label?: string;
}

const statusConfig = {
  connected: {
    color: 'success' as const,
    icon: <ConnectedIcon />,
    label: 'Connected',
  },
  error: {
    color: 'error' as const,
    icon: <ErrorIcon />,
    label: 'Error',
  },
  disabled: {
    color: 'default' as const,
    icon: <DisabledIcon />,
    label: 'Disabled',
  },
  running: {
    color: 'primary' as const,
    icon: <ConnectedIcon />,
    label: 'Running',
  },
  stopped: {
    color: 'default' as const,
    icon: <DisabledIcon />,
    label: 'Stopped',
  },
};

export default function StatusBadge({ status, label }: StatusBadgeProps) {
  const config = statusConfig[status];

  return (
    <Chip
      icon={config.icon}
      label={label || config.label}
      color={config.color}
      size="small"
      variant="outlined"
    />
  );
}
