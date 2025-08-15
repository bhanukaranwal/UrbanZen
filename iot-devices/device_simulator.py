import asyncio
import json
import random
import time
import uuid
from datetime import datetime, timedelta
from typing import Dict, List
import aiohttp
import asyncio_mqtt as aiomqtt
from dataclasses import dataclass
import logging

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

@dataclass
class DeviceConfig:
    device_id: str
    device_type: str
    location: Dict[str, float]
    mqtt_topic: str
    interval: int  # seconds
    
class BaseDeviceSimulator:
    def __init__(self, config: DeviceConfig, mqtt_broker: str):
        self.config = config
        self.mqtt_broker = mqtt_broker
        self.running = False
        
    async def start(self):
        """Start the device simulation"""
        self.running = True
        
        async with aiomqtt.Client(self.mqtt_broker) as client:
            logger.info(f"Started {self.config.device_type} simulator: {self.config.device_id}")
            
            while self.running:
                try:
                    # Generate sensor data
                    data = self.generate_data()
                    
                    # Add common fields
                    message = {
                        'device_id': self.config.device_id,
                        'device_type': self.config.device_type,
                        'timestamp': datetime.utcnow().isoformat(),
                        'location': self.config.location,
                        'metrics': data,
                        'metadata': {
                            'firmware_version': '1.2.3',
                            'battery_level': random.uniform(20, 100)
                        }
                    }
                    
                    # Publish to MQTT
                    await client.publish(
                        self.config.mqtt_topic,
                        json.dumps(message)
                    )
                    
                    logger.debug(f"Published data for {self.config.device_id}")
                    
                    # Wait for next interval
                    await asyncio.sleep(self.config.interval)
                    
                except Exception as e:
                    logger.error(f"Error in {self.config.device_id}: {e}")
                    await asyncio.sleep(5)
    
    def stop(self):
        """Stop the device simulation"""
        self.running = False
    
    def generate_data(self) -> Dict:
        """Generate sensor data - to be implemented by subclasses"""
        raise NotImplementedError

class WaterSensorSimulator(BaseDeviceSimulator):
    def __init__(self, config: DeviceConfig, mqtt_broker: str):
        super().__init__(config, mqtt_broker)
        self.base_flow_rate = random.uniform(10, 50)
        self.base_pressure = random.uniform(2, 5)
        self.base_ph = random.uniform(6.5, 8.5)
        
    def generate_data(self) -> Dict:
        # Simulate daily patterns
        hour = datetime.now().hour
        peak_multiplier = 1.5 if 6 <= hour <= 9 or 18 <= hour <= 21 else 1.0
        
        # Add some randomness and occasional anomalies
        anomaly_chance = 0.02  # 2% chance of anomaly
        
        if random.random() < anomaly_chance:
            # Generate anomalous data
            flow_rate = self.base_flow_rate * random.uniform(3, 10)  # Sudden spike
            pressure = self.base_pressure * random.uniform(0.1, 0.3)  # Pressure drop
            ph_level = random.uniform(4, 11)  # Extreme pH
        else:
            # Normal data with daily patterns
            flow_rate = self.base_flow_rate * peak_multiplier * random.uniform(0.8, 1.2)
            pressure = self.base_pressure * random.uniform(0.9, 1.1)
            ph_level = self.base_ph + random.uniform(-0.5, 0.5)
        
        return {
            'flow_rate': round(flow_rate, 2),
            'pressure': round(pressure, 2),
            'ph_level': round(ph_level, 2),
            'turbidity': round(random.uniform(0, 5), 2),
            'temperature': round(random.uniform(15, 35), 1),
            'conductivity': round(random.uniform(200, 800), 1)
        }

class ElectricityMeterSimulator(BaseDeviceSimulator):
    def __init__(self, config: DeviceConfig, mqtt_broker: str):
        super().__init__(config, mqtt_broker)
        self.base_voltage = 230  # Standard voltage
        self.base_current = random.uniform(5, 25)
        
    def generate_data(self) -> Dict:
        # Simulate daily consumption patterns
        hour = datetime.now().hour
        if 6 <= hour <= 9 or 18 <= hour <= 23:
            load_factor = random.uniform(0.7, 1.0)  # Peak hours
        elif 10 <= hour <= 17:
            load_factor = random.uniform(0.3, 0.6)  # Day time
        else:
            load_factor = random.uniform(0.1, 0.3)  # Night time
        
        # Occasional power quality issues
        anomaly_chance = 0.03
        
        if random.random() < anomaly_chance:
            voltage = self.base_voltage * random.uniform(0.8, 1.2)  # Voltage fluctuation
            current = self.base_current * random.uniform(2, 5)  # Current spike
            frequency = 50 + random.uniform(-2, 2)  # Frequency deviation
        else:
            voltage = self.base_voltage * random.uniform(0.95, 1.05)
            current = self.base_current * load_factor * random.uniform(0.9, 1.1)
            frequency = 50 + random.uniform(-0.1, 0.1)
        
        power = voltage * current * random.uniform(0.8, 1.0)  # Power factor
        energy = power * (self.config.interval / 3600)  # kWh for this interval
        
        return {
            'voltage': round(voltage, 1),
            'current': round(current, 2),
            'power': round(power, 2),
            'energy': round(energy, 4),
            'frequency': round(frequency, 2),
            'power_factor': round(random.uniform(0.8, 1.0), 3),
            'total_energy': round(random.uniform(1000, 5000), 2)  # Cumulative reading
        }

class TrafficCameraSimulator(BaseDeviceSimulator):
    def __init__(self, config: DeviceConfig, mqtt_broker: str):
        super().__init__(config, mqtt_broker)
        self.road_capacity = random.randint(100, 500)
        
    def generate_data(self) -> Dict:
        # Simulate traffic patterns
        hour = datetime.now().hour
        day_of_week = datetime.now().weekday()
        
        # Weekend vs weekday patterns
        if day_of_week >= 5:  # Weekend
            if 10 <= hour <= 14 or 19 <= hour <= 22:
                traffic_factor = random.uniform(0.4, 0.7)
            else:
                traffic_factor = random.uniform(0.1, 0.3)
        else:  # Weekday
            if 7 <= hour <= 9 or 17 <= hour <= 19:
                traffic_factor = random.uniform(0.7, 1.0)  # Rush hours
            elif 10 <= hour <= 16:
                traffic_factor = random.uniform(0.3, 0.6)
            else:
                traffic_factor = random.uniform(0.1, 0.3)
        
        vehicle_count = int(self.road_capacity * traffic_factor)
        
        # Calculate congestion and average speed
        congestion_level = min(vehicle_count / self.road_capacity, 1.0)
        base_speed = 60  # km/h
        avg_speed = base_speed * (1 - congestion_level * 0.8)
        
        # Simulate incidents occasionally
        incident_chance = 0.01
        if random.random() < incident_chance:
            vehicle_count = int(vehicle_count * 1.5)  # Traffic backup
            avg_speed = avg_speed * 0.3  # Slow down
            incident_detected = True
        else:
            incident_detected = False
        
        return {
            'vehicle_count': vehicle_count,
            'avg_speed': round(avg_speed, 1),
            'congestion_level': round(congestion_level, 3),
            'incident_detected': incident_detected,
            'air_quality_index': random.randint(50, 200),
            'noise_level': round(random.uniform(45, 85), 1),
            'visibility': round(random.uniform(5, 15), 1)  # km
        }

class DeviceFleetSimulator:
    def __init__(self, mqtt_broker: str = "localhost:1883"):
        self.mqtt_broker = mqtt_broker
        self.simulators: List[BaseDeviceSimulator] = []
        
    def add_device(self, device_type: str, location: Dict[str, float], 
                   interval: int = 30) -> str:
        """Add a new device to the simulation fleet"""
        device_id = f"{device_type}_{uuid.uuid4().hex[:8]}"
        
        config = DeviceConfig(
            device_id=device_id,
            device_type=device_type,
            location=location,
            mqtt_topic=f"devices/{device_type}/{device_id}",
            interval=interval
        )
        
        if device_type == "water_sensor":
            simulator = WaterSensorSimulator(config, self.mqtt_broker)
        elif device_type == "electricity_meter":
            simulator = ElectricityMeterSimulator(config, self.mqtt_broker)
        elif device_type == "traffic_camera":
            simulator = TrafficCameraSimulator(config, self.mqtt_broker)
        else:
            raise ValueError(f"Unknown device type: {device_type}")
        
        self.simulators.append(simulator)
        logger.info(f"Added {device_type} simulator: {device_id}")
        return device_id
    
    async def start_all(self):
        """Start all device simulators"""
        logger.info(f"Starting {len(self.simulators)} device simulators")
        
        tasks = []
        for simulator in self.simulators:
            task = asyncio.create_task(simulator.start())
            tasks.append(task)
        
        try:
            await asyncio.gather(*tasks)
        except KeyboardInterrupt:
            logger.info("Stopping all simulators...")
            for simulator in self.simulators:
                simulator.stop()
    
    def create_delhi_fleet(self):
        """Create a realistic fleet of devices for Delhi"""
        # Delhi bounding box coordinates
        delhi_bounds = {
            'north': 28.8836,
            'south': 28.4024,
            'east': 77.3460,
            'west': 76.8389
        }
        
        # Generate random locations within Delhi
        def random_delhi_location():
            return {
                'latitude': random.uniform(delhi_bounds['south'], delhi_bounds['north']),
                'longitude': random.uniform(delhi_bounds['west'], delhi_bounds['east'])
            }
        
        # Add water sensors (50 devices)
        for _ in range(50):
            self.add_device("water_sensor", random_delhi_location(), interval=60)
        
        # Add electricity meters (100 devices)
        for _ in range(100):
            self.add_device("electricity_meter", random_delhi_location(), interval=300)
        
        # Add traffic cameras (30 devices)
        for _ in range(30):
            self.add_device("traffic_camera", random_delhi_location(), interval=30)
        
        logger.info("Created Delhi device fleet: 50 water sensors, 100 electricity meters, 30 traffic cameras")

async def main():
    """Main function to run the device fleet simulation"""
    fleet = DeviceFleetSimulator()
    
    # Create Delhi fleet
    fleet.create_delhi_fleet()
    
    # Start all simulators
    try:
        await fleet.start_all()
    except KeyboardInterrupt:
        logger.info("Simulation stopped by user")

if __name__ == "__main__":
    asyncio.run(main())