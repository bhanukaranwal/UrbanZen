import React, { useState, useEffect } from 'react';
import {
  Grid,
  Card,
  CardContent,
  Typography,
  Box,
  Button,
  Alert,
  Chip,
} from '@mui/material';
import {
  WaterDrop,
  ElectricBolt,
  DirectionsCar,
  Warning,
  CheckCircle,
  Error,
  Info,
} from '@mui/icons-material';
import { MapContainer, TileLayer, Marker, Popup } from 'react-leaflet';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';
import { useQuery } from '@tanstack/react-query';
import { api } from '../services/api';
import 'leaflet/dist/leaflet.css';

interface DashboardData {
  totalDevices: number;
  activeDevices: number;
  offlineDevices: number;
  criticalAlerts: number;
  waterQuality: number;
  powerGrid: {
    status: string;
    load: number;
  };
  trafficIncidents: number;
  recentAlerts: Alert[];
  deviceLocations: DeviceLocation[];
  consumptionTrends: TrendData[];
}

interface Alert {
  id: string;
  type: string;
  severity: string;
  message: string;
  timestamp: string;
  deviceId?: string;
}

interface DeviceLocation {
  id: string;
  name: string;
  type: string;
  latitude: number;
  longitude: number;
  status: string;
}

interface TrendData {
  timestamp: string;
  water: number;
  electricity: number;
  traffic: number;
}

export const Dashboard: React.FC = () => {
  const [selectedTimeRange, setSelectedTimeRange] = useState('24h');

  const { data: dashboardData, isLoading, error } = useQuery({
    queryKey: ['dashboard', selectedTimeRange],
    queryFn: () => api.getDashboardData(selectedTimeRange),
    refetchInterval: 30000, // Refresh every 30 seconds
  });

  if (isLoading) {
    return (
      <Box display="flex" justifyContent="center" alignItems="center" height="400px">
        <Typography>Loading dashboard...</Typography>
      </Box>
    );
  }

  if (error) {
    return (
      <Alert severity="error">
        Failed to load dashboard data. Please try again.
      </Alert>
    );
  }

  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case 'critical': return 'error';
      case 'warning': return 'warning';
      case 'info': return 'info';
      default: return 'default';
    }
  };

  const getDeviceIcon = (type: string) => {
    switch (type) {
      case 'water_sensor': return <WaterDrop color="primary" />;
      case 'electricity_meter': return <ElectricBolt color="warning" />;
      case 'traffic_camera': return <DirectionsCar color="secondary" />;
      default: return <Info />;
    }
  };

  return (
    <Box p={3}>
      <Typography variant="h4" gutterBottom>
        Smart City Dashboard
      </Typography>

      <Box mb={3}>
        <Button
          variant={selectedTimeRange === '1h' ? 'contained' : 'outlined'}
          onClick={() => setSelectedTimeRange('1h')}
          sx={{ mr: 1 }}
        >
          1 Hour
        </Button>
        <Button
          variant={selectedTimeRange === '24h' ? 'contained' : 'outlined'}
          onClick={() => setSelectedTimeRange('24h')}
          sx={{ mr: 1 }}
        >
          24 Hours
        </Button>
        <Button
          variant={selectedTimeRange === '7d' ? 'contained' : 'outlined'}
          onClick={() => setSelectedTimeRange('7d')}
        >
          7 Days
        </Button>
      </Box>

      {/* Key Metrics */}
      <Grid container spacing={3} mb={3}>
        <Grid item xs={12} sm={6} md={3}>
          <Card>
            <CardContent>
              <Box display="flex" alignItems="center" justifyContent="space-between">
                <Box>
                  <Typography color="textSecondary" gutterBottom>
                    Total Devices
                  </Typography>
                  <Typography variant="h4">
                    {dashboardData?.totalDevices || 0}
                  </Typography>
                </Box>
                <CheckCircle color="primary" sx={{ fontSize: 40 }} />
              </Box>
            </CardContent>
          </Card>
        </Grid>

        <Grid item xs={12} sm={6} md={3}>
          <Card>
            <CardContent>
              <Box display="flex" alignItems="center" justifyContent="space-between">
                <Box>
                  <Typography color="textSecondary" gutterBottom>
                    Active Devices
                  </Typography>
                  <Typography variant="h4" color="success.main">
                    {dashboardData?.activeDevices || 0}
                  </Typography>
                </Box>
                <CheckCircle color="success" sx={{ fontSize: 40 }} />
              </Box>
            </CardContent>
          </Card>
        </Grid>

        <Grid item xs={12} sm={6} md={3}>
          <Card>
            <CardContent>
              <Box display="flex" alignItems="center" justifyContent="space-between">
                <Box>
                  <Typography color="textSecondary" gutterBottom>
                    Offline Devices
                  </Typography>
                  <Typography variant="h4" color="error.main">
                    {dashboardData?.offlineDevices || 0}
                  </Typography>
                </Box>
                <Error color="error" sx={{ fontSize: 40 }} />
              </Box>
            </CardContent>
          </Card>
        </Grid>

        <Grid item xs={12} sm={6} md={3}>
          <Card>
            <CardContent>
              <Box display="flex" alignItems="center" justifyContent="space-between">
                <Box>
                  <Typography color="textSecondary" gutterBottom>
                    Critical Alerts
                  </Typography>
                  <Typography variant="h4" color="warning.main">
                    {dashboardData?.criticalAlerts || 0}
                  </Typography>
                </Box>
                <Warning color="warning" sx={{ fontSize: 40 }} />
              </Box>
            </CardContent>
          </Card>
        </Grid>
      </Grid>

      <Grid container spacing={3}>
        {/* Map */}
        <Grid item xs={12} md={8}>
          <Card>
            <CardContent>
              <Typography variant="h6" gutterBottom>
                Device Locations
              </Typography>
              <Box height="400px">
                <MapContainer
                  center={[28.6139, 77.2090]} // Delhi coordinates
                  zoom={12}
                  style={{ height: '100%', width: '100%' }}
                >
                  <TileLayer
                    url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
                    attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
                  />
                  {dashboardData?.deviceLocations?.map((device) => (
                    <Marker
                      key={device.id}
                      position={[device.latitude, device.longitude]}
                    >
                      <Popup>
                        <Box>
                          <Typography variant="subtitle2">{device.name}</Typography>
                          <Typography variant="body2" color="textSecondary">
                            Type: {device.type}
                          </Typography>
                          <Typography variant="body2" color="textSecondary">
                            Status: {device.status}
                          </Typography>
                        </Box>
                      </Popup>
                    </Marker>
                  ))}
                </MapContainer>
              </Box>
            </CardContent>
          </Card>
        </Grid>

        {/* Recent Alerts */}
        <Grid item xs={12} md={4}>
          <Card>
            <CardContent>
              <Typography variant="h6" gutterBottom>
                Recent Alerts
              </Typography>
              <Box maxHeight="400px" overflow="auto">
                {dashboardData?.recentAlerts?.map((alert) => (
                  <Box key={alert.id} mb={2} p={2} border={1} borderColor="grey.300" borderRadius={1}>
                    <Box display="flex" justifyContent="space-between" alignItems="center" mb={1}>
                      <Chip
                        label={alert.severity}
                        color={getSeverityColor(alert.severity) as any}
                        size="small"
                      />
                      <Typography variant="caption" color="textSecondary">
                        {new Date(alert.timestamp).toLocaleTimeString()}
                      </Typography>
                    </Box>
                    <Typography variant="body2">
                      {alert.message}
                    </Typography>
                    {alert.deviceId && (
                      <Typography variant="caption" color="textSecondary">
                        Device: {alert.deviceId}
                      </Typography>
                    )}
                  </Box>
                ))}
              </Box>
            </CardContent>
          </Card>
        </Grid>

        {/* Consumption Trends */}
        <Grid item xs={12}>
          <Card>
            <CardContent>
              <Typography variant="h6" gutterBottom>
                Utility Consumption Trends
              </Typography>
              <Box height="300px">
                <ResponsiveContainer width="100%" height="100%">
                  <LineChart data={dashboardData?.consumptionTrends || []}>
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis dataKey="timestamp" />
                    <YAxis />
                    <Tooltip />
                    <Line type="monotone" dataKey="water" stroke="#2196f3" name="Water" />
                    <Line type="monotone" dataKey="electricity" stroke="#ff9800" name="Electricity" />
                    <Line type="monotone" dataKey="traffic" stroke="#4caf50" name="Traffic" />
                  </LineChart>
                </ResponsiveContainer>
              </Box>
            </CardContent>
          </Card>
        </Grid>

        {/* Utility Status Cards */}
        <Grid item xs={12} md={4}>
          <Card>
            <CardContent>
              <Box display="flex" alignItems="center" mb={2}>
                <WaterDrop color="primary" sx={{ mr: 1 }} />
                <Typography variant="h6">Water Quality</Typography>
              </Box>
              <Typography variant="h4" color="primary">
                {dashboardData?.waterQuality || 0}%
              </Typography>
              <Typography variant="body2" color="textSecondary">
                Overall water quality index
              </Typography>
            </CardContent>
          </Card>
        </Grid>

        <Grid item xs={12} md={4}>
          <Card>
            <CardContent>
              <Box display="flex" alignItems="center" mb={2}>
                <ElectricBolt color="warning" sx={{ mr: 1 }} />
                <Typography variant="h6">Power Grid</Typography>
              </Box>
              <Typography variant="h4" color="warning.main">
                {dashboardData?.powerGrid?.load || 0}%
              </Typography>
              <Typography variant="body2" color="textSecondary">
                Current grid load - {dashboardData?.powerGrid?.status || 'Unknown'}
              </Typography>
            </CardContent>
          </Card>
        </Grid>

        <Grid item xs={12} md={4}>
          <Card>
            <CardContent>
              <Box display="flex" alignItems="center" mb={2}>
                <DirectionsCar color="secondary" sx={{ mr: 1 }} />
                <Typography variant="h6">Traffic Incidents</Typography>
              </Box>
              <Typography variant="h4" color="secondary.main">
                {dashboardData?.trafficIncidents || 0}
              </Typography>
              <Typography variant="body2" color="textSecondary">
                Active traffic incidents
              </Typography>
            </CardContent>
          </Card>
        </Grid>
      </Grid>
    </Box>
  );
};