// Continuing from previous file...

                  child: UtilityCard(
                    title: 'Electricity',
                    value: '${dataProvider.electricityConsumption?.toStringAsFixed(1) ?? "0"} kWh',
                    change: '-2%',
                    isPositive: true,
                    icon: Icons.electric_bolt,
                    color: Colors.amber,
                    onTap: () => Navigator.pushNamed(context, Routes.consumption),
                  ),
                ),
              ],
            ),
            SizedBox(height: 12),
            UtilityCard(
              title: 'Air Quality Index',
              value: '${dataProvider.airQualityIndex ?? "Good"}',
              change: 'Moderate',
              isPositive: dataProvider.airQualityIndex != null && dataProvider.airQualityIndex! < 100,
              icon: Icons.air,
              color: Colors.green,
              onTap: () {},
            ),
          ],
        );
      },
    );
  }

  Widget _buildRecentAlerts() {
    return Consumer<DataProvider>(
      builder: (context, dataProvider, child) {
        return Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text(
                  'Recent Alerts',
                  style: Theme.of(context).textTheme.titleLarge,
                ),
                TextButton(
                  onPressed: () => Navigator.pushNamed(context, Routes.alerts),
                  child: Text('View All'),
                ),
              ],
            ),
            SizedBox(height: 16),
            if (dataProvider.recentAlerts.isEmpty)
              Card(
                child: Padding(
                  padding: EdgeInsets.all(16),
                  child: Center(
                    child: Column(
                      children: [
                        Icon(Icons.check_circle, size: 48, color: Colors.green),
                        SizedBox(height: 8),
                        Text('All systems operational'),
                      ],
                    ),
                  ),
                ),
              )
            else
              ...dataProvider.recentAlerts.take(3).map((alert) => 
                AlertCard(alert: alert)
              ).toList(),
          ],
        );
      },
    );
  }

  Widget _buildMapTab() {
    return Consumer<DataProvider>(
      builder: (context, dataProvider, child) {
        return GoogleMap(
          initialCameraPosition: CameraPosition(
            target: LatLng(28.6139, 77.2090), // Delhi coordinates
            zoom: 12,
          ),
          markers: dataProvider.deviceMarkers,
          onMapCreated: (GoogleMapController controller) {
            dataProvider.setMapController(controller);
          },
        );
      },
    );
  }

  Widget _buildServicesTab() {
    return SingleChildScrollView(
      padding: EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            'Citizen Services',
            style: Theme.of(context).textTheme.titleLarge,
          ),
          SizedBox(height: 16),
          _buildServiceGrid(),
        ],
      ),
    );
  }

  Widget _buildServiceGrid() {
    final services = [
      {
        'title': 'File Complaint',
        'icon': Icons.report_problem,
        'color': Colors.red,
        'route': Routes.complaints,
      },
      {
        'title': 'View Bills',
        'icon': Icons.receipt,
        'color': Colors.blue,
        'route': Routes.consumption,
      },
      {
        'title': 'Emergency Services',
        'icon': Icons.emergency,
        'color': Colors.red[700],
        'route': '/emergency',
      },
      {
        'title': 'Public Transport',
        'icon': Icons.directions_bus,
        'color': Colors.green,
        'route': '/transport',
      },
    ];

    return GridView.builder(
      shrinkWrap: true,
      physics: NeverScrollableScrollPhysics(),
      gridDelegate: SliverGridDelegateWithFixedCrossAxisCount(
        crossAxisCount: 2,
        crossAxisSpacing: 16,
        mainAxisSpacing: 16,
        childAspectRatio: 1.2,
      ),
      itemCount: services.length,
      itemBuilder: (context, index) {
        final service = services[index];
        return Card(
          child: InkWell(
            onTap: () {
              if (service['route'] != null) {
                Navigator.pushNamed(context, service['route'] as String);
              }
            },
            child: Padding(
              padding: EdgeInsets.all(16),
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Icon(
                    service['icon'] as IconData,
                    size: 48,
                    color: service['color'] as Color,
                  ),
                  SizedBox(height: 12),
                  Text(
                    service['title'] as String,
                    textAlign: TextAlign.center,
                    style: Theme.of(context).textTheme.titleMedium,
                  ),
                ],
              ),
            ),
          ),
        );
      },
    );
  }
}