import React from 'react';
import { Typography, Box } from '@mui/material';

const DeviceManagement: React.FC = () => {
  return (
    <Box>
      <Typography variant="h4" gutterBottom>
        Device Management
      </Typography>
      <Typography variant="body1">
        IoT device registration, monitoring, and control interface.
      </Typography>
    </Box>
  );
};

export default DeviceManagement;