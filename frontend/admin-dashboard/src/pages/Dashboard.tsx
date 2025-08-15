import React from 'react';
import {
  Grid,
  Paper,
  Typography,
  Box,
  Card,
  CardContent,
  LinearProgress,
  Chip,
  Avatar,
  List,
  ListItem,
  ListItemAvatar,
  ListItemText,
  IconButton,
} from '@mui/material';
import {
  TrendingUp,
  TrendingDown,
  DeviceHub,
  WaterDrop,
  ElectricBolt,
  Traffic,
  Air,
  Warning,
  CheckCircle,
  Error,
  Refresh,
} from '@mui/icons-material';
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  BarChart,
  Bar,
  PieChart,
  Pie,
  Cell,
} from 'recharts';

// Mock data for demonstration
const systemStats = [
  {
    title: 'Total Devices',
    value: '12,547',
    change: '+5.2%',
    trend: 'up',
    icon: <DeviceHub />,
    color: '#1976d2',
  },
  {
    title: 'Water Consumption',
    value: '2.4M L',
    change: '-2.1%',
    trend: 'down',
    icon: <WaterDrop />,
    color: '#2196f3',
  },
  {
    title: 'Energy Usage',
    value: '1,847 kWh',
    change: '+1.8%',
    trend: 'up',
    icon: <ElectricBolt />,
    color: '#ff9800',
  },
  {
    title: 'Traffic Flow',
    value: '89.2%',
    change: '+3.4%',
    trend: 'up',
    icon: <Traffic />,
    color: '#4caf50',
  },
];

const consumptionData = [
  { time: '00:00', water: 850, electricity: 1200 },
  { time: '04:00', water: 720, electricity: 980 },
  { time: '08:00', water: 1100, electricity: 1800 },
  { time: '12:00', water: 1350, electricity: 2200 },
  { time: '16:00', water: 1200, electricity: 2100 },
  { time: '20:00', water: 1000, electricity: 1900 },
];

const alertData = [
  { name: 'Water Leaks', value: 3, color: '#f44336' },
  { name: 'Power Outages', value: 1, color: '#ff9800' },
  { name: 'Traffic Issues', value: 5, color: '#2196f3' },
  { name: 'Air Quality', value: 2, color: '#9c27b0' },
];

const recentAlerts = [
  {
    id: 1,
    type: 'Water Leak',
    location: 'Sector 15, Block A',
    severity: 'high',
    time: '2 mins ago',
    status: 'active',
  },
  {
    id: 2,
    type: 'Power Outage',
    location: 'Industrial Area',
    severity: 'critical',
    time: '5 mins ago',
    status: 'resolving',
  },
  {
    id: 3,
    type: 'Traffic Congestion',
    location: 'Main Street Junction',
    severity: 'medium',
    time: '8 mins ago',
    status: 'monitoring',
  },
  {
    id: 4,
    type: 'Air Quality Alert',
    location: 'City Center',
    severity: 'low',
    time: '12 mins ago',
    status: 'resolved',
  },
];

const getSeverityColor = (severity: string) => {
  switch (severity) {
    case 'critical': return '#f44336';
    case 'high': return '#ff5722';
    case 'medium': return '#ff9800';
    case 'low': return '#4caf50';
    default: return '#9e9e9e';
  }
};

const getStatusIcon = (status: string) => {
  switch (status) {
    case 'active': return <Error color="error" />;
    case 'resolving': return <Warning color="warning" />;
    case 'monitoring': return <Refresh color="info" />;
    case 'resolved': return <CheckCircle color="success" />;
    default: return <Warning />;
  }
};

const Dashboard: React.FC = () => {
  return (
    <Box>
      <Typography variant="h4" gutterBottom>
        Smart City Dashboard
      </Typography>
      <Typography variant="subtitle1" color="text.secondary" gutterBottom>
        Real-time monitoring and analytics for city infrastructure
      </Typography>

      {/* Key Metrics */}
      <Grid container spacing={3} sx={{ mb: 3 }}>
        {systemStats.map((stat, index) => (
          <Grid item xs={12} sm={6} md={3} key={index}>
            <Card>
              <CardContent>
                <Box sx={{ display: 'flex', alignItems: 'center', mb: 2 }}>
                  <Avatar sx={{ bgcolor: stat.color, mr: 2 }}>
                    {stat.icon}
                  </Avatar>
                  <Box>
                    <Typography variant="h5" fontWeight="bold">
                      {stat.value}
                    </Typography>
                    <Typography variant="body2" color="text.secondary">
                      {stat.title}
                    </Typography>
                  </Box>
                </Box>
                <Box sx={{ display: 'flex', alignItems: 'center' }}>
                  {stat.trend === 'up' ? (
                    <TrendingUp color="success" fontSize="small" />
                  ) : (
                    <TrendingDown color="error" fontSize="small" />
                  )}
                  <Typography
                    variant="body2"
                    color={stat.trend === 'up' ? 'success.main' : 'error.main'}
                    sx={{ ml: 0.5 }}
                  >
                    {stat.change}
                  </Typography>
                  <Typography variant="body2" color="text.secondary" sx={{ ml: 1 }}>
                    vs last month
                  </Typography>
                </Box>
              </CardContent>
            </Card>
          </Grid>
        ))}
      </Grid>

      <Grid container spacing={3}>
        {/* Consumption Trends */}
        <Grid item xs={12} lg={8}>
          <Paper sx={{ p: 3 }}>
            <Typography variant="h6" gutterBottom>
              Resource Consumption Trends
            </Typography>
            <Box sx={{ height: 300 }}>
              <ResponsiveContainer width="100%" height="100%">
                <LineChart data={consumptionData}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="time" />
                  <YAxis />
                  <Tooltip />
                  <Line
                    type="monotone"
                    dataKey="water"
                    stroke="#2196f3"
                    strokeWidth={2}
                    name="Water (L)"
                  />
                  <Line
                    type="monotone"
                    dataKey="electricity"
                    stroke="#ff9800"
                    strokeWidth={2}
                    name="Electricity (kWh)"
                  />
                </LineChart>
              </ResponsiveContainer>
            </Box>
          </Paper>
        </Grid>

        {/* Alert Distribution */}
        <Grid item xs={12} lg={4}>
          <Paper sx={{ p: 3 }}>
            <Typography variant="h6" gutterBottom>
              Active Alerts
            </Typography>
            <Box sx={{ height: 300 }}>
              <ResponsiveContainer width="100%" height="100%">
                <PieChart>
                  <Pie
                    data={alertData}
                    cx="50%"
                    cy="50%"
                    outerRadius={80}
                    dataKey="value"
                    label
                  >
                    {alertData.map((entry, index) => (
                      <Cell key={`cell-${index}`} fill={entry.color} />
                    ))}
                  </Pie>
                  <Tooltip />
                </PieChart>
              </ResponsiveContainer>
            </Box>
          </Paper>
        </Grid>

        {/* Recent Alerts */}
        <Grid item xs={12} lg={6}>
          <Paper sx={{ p: 3 }}>
            <Typography variant="h6" gutterBottom>
              Recent Alerts
            </Typography>
            <List>
              {recentAlerts.map((alert) => (
                <ListItem key={alert.id} sx={{ px: 0 }}>
                  <ListItemAvatar>
                    {getStatusIcon(alert.status)}
                  </ListItemAvatar>
                  <ListItemText
                    primary={
                      <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                        <Typography variant="subtitle2">{alert.type}</Typography>
                        <Chip
                          size="small"
                          label={alert.severity}
                          sx={{
                            backgroundColor: getSeverityColor(alert.severity),
                            color: 'white',
                            fontSize: '0.7rem',
                          }}
                        />
                      </Box>
                    }
                    secondary={
                      <Box>
                        <Typography variant="body2" color="text.secondary">
                          {alert.location}
                        </Typography>
                        <Typography variant="caption" color="text.secondary">
                          {alert.time}
                        </Typography>
                      </Box>
                    }
                  />
                </ListItem>
              ))}
            </List>
          </Paper>
        </Grid>

        {/* System Health */}
        <Grid item xs={12} lg={6}>
          <Paper sx={{ p: 3 }}>
            <Typography variant="h6" gutterBottom>
              System Health
            </Typography>
            <Box sx={{ mb: 2 }}>
              <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 1 }}>
                <Typography variant="body2">Database Performance</Typography>
                <Typography variant="body2">95%</Typography>
              </Box>
              <LinearProgress variant="determinate" value={95} color="success" />
            </Box>
            <Box sx={{ mb: 2 }}>
              <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 1 }}>
                <Typography variant="body2">API Response Time</Typography>
                <Typography variant="body2">87%</Typography>
              </Box>
              <LinearProgress variant="determinate" value={87} color="success" />
            </Box>
            <Box sx={{ mb: 2 }}>
              <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 1 }}>
                <Typography variant="body2">Device Connectivity</Typography>
                <Typography variant="body2">92%</Typography>
              </Box>
              <LinearProgress variant="determinate" value={92} color="success" />
            </Box>
            <Box>
              <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 1 }}>
                <Typography variant="body2">Storage Usage</Typography>
                <Typography variant="body2">68%</Typography>
              </Box>
              <LinearProgress variant="determinate" value={68} color="warning" />
            </Box>
          </Paper>
        </Grid>
      </Grid>
    </Box>
  );
};

export default Dashboard;