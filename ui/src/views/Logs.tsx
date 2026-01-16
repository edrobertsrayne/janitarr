/**
 * Logs view - Activity logs with real-time streaming
 */

import { Box, Typography } from '@mui/material';

export default function Logs() {
  return (
    <Box>
      <Typography variant="h4" gutterBottom>
        Logs
      </Typography>
      <Typography variant="body1" color="text.secondary">
        Log viewer with real-time streaming will appear here.
      </Typography>
    </Box>
  );
}
