/**
 * Dashboard view - Overview of system status
 */

import { Box, Typography } from '@mui/material';

export default function Dashboard() {
  return (
    <Box>
      <Typography variant="h4" gutterBottom>
        Dashboard
      </Typography>
      <Typography variant="body1" color="text.secondary">
        Status cards, server list, and recent activity will appear here.
      </Typography>
    </Box>
  );
}
