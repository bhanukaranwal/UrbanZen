#!/usr/bin/env python3
"""
Water Meter IoT Device Simulator
Simulates water consumption and quality data for UrbanZen platform
"""

import json
import time
import random
import logging
from datetime import datetime, timedelta
from typing import Dict, Any

import paho.mqtt.client as mqtt
import numpy as np
from dataclasses import dataclass


@dataclass
class WaterMeterConfig:
    device_id: str
    location: Dict[str, float]
    mqtt_broker: str
    mqtt_port: int
    base_flow_rate: float = 15.0  # L/min
    max_flow_rate: float = 50.0   # L/min
    pressure_range: tuple = (1.5, 3.0)  # bar
    ph_range: tuple = (6.5, 8.5)
    temperature_range: tuple = (15.0, 25.0)  # Celsius
    publish_interval: int = 60  # seconds


class WaterMeterSimulator:
    def __init__(self, config: WaterMeterConfig):
        self.config = config
        self.mqtt_client = None
        self.is_running = False
        self.cumulative_volume = 0.0
        self.last_reading_time = datetime.now()
        
        # Setup logging
        logging.basicConfig(level=logging.INFO)
        self.logger = logging.getLogger(f"WaterMeter-{config.device_id}")
        
        # Initialize MQTT client
        self._setup_mqtt()
    
    def _setup_mqtt(self):
        """Setup MQTT client"""
        self.mqtt_client = mqtt.Client(client_id=f"water_meter_{self.config.device_id}")
        self.mqtt_client.on_connect = self._on_mqtt_connect
        self.mqtt_client.on_disconnect = self._on_mqtt_disconnect
        self.mqtt_client.on_publish = self._on_mqtt_publish
    
    def _on_mqtt_connect(self, client, userdata, flags, rc):
        """MQTT connection callback"""
        if rc == 0:
            self.logger.info(f"Connected to MQTT broker at {self.config.mqtt_broker}")
            # Subscribe to command topic
            command_topic = f"urbanzen/devices/{self.config.device_id}/commands"
            client.subscribe(command_topic)
            self.logger.info(f"Subscribed to {command_topic}")
        else:
            self.logger.error(f"Failed to connect to MQTT broker, return code {rc}")
    
    def _on_mqtt_disconnect(self, client, userdata, rc):
        """MQTT disconnection callback"""
        self.logger.warning("Disconnected from MQTT broker")
    
    def _on_mqtt_publish(self, client, userdata, mid):
        """MQTT publish callback"""
        self.logger.debug(f"Message published with mid: {mid}")
    
    def _generate_flow_rate(self) -> float:
        """Generate realistic flow rate based on time of day"""
        current_hour = datetime.now().hour
        
        # Peak hours: 6-9 AM and 6-9 PM
        if 6 <= current_hour <= 9 or 18 <= current_hour <= 21:
            base_multiplier = 1.5
        # Moderate hours: 10 AM - 5 PM
        elif 10 <= current_hour <= 17:
            base_multiplier = 1.0
        # Low hours: 10 PM - 5 AM
        else:
            base_multiplier = 0.3
        
        # Add random variation
        variation = random.uniform(0.8, 1.2)
        flow_rate = self.config.base_flow_rate * base_multiplier * variation
        
        # Occasionally simulate high usage
        if random.random() < 0.05:  # 5% chance
            flow_rate = random.uniform(self.config.max_flow_rate * 0.7, self.config.max_flow_rate)
        
        return min(flow_rate, self.config.max_flow_rate)
    
    def _generate_sensor_data(self) -> Dict[str, Any]:
        """Generate sensor data"""
        current_time = datetime.now()
        time_diff = (current_time - self.last_reading_time).total_seconds() / 60  # minutes
        
        # Flow rate
        flow_rate = self._generate_flow_rate()
        
        # Calculate volume consumed
        volume_consumed = flow_rate * time_diff
        self.cumulative_volume += volume_consumed
        
        # Pressure (varies with flow rate)
        pressure = random.uniform(*self.config.pressure_range)
        if flow_rate > self.config.base_flow_rate * 1.5:
            pressure *= 0.9  # Pressure drops with high flow
        
        # Water quality parameters
        ph_level = random.uniform(*self.config.ph_range)
        temperature = random.uniform(*self.config.temperature_range)
        turbidity = random.uniform(0.1, 2.0)  # NTU
        chlorine_level = random.uniform(0.2, 1.5)  # mg/L
        tds = random.uniform(150, 300)  # ppm
        
        # Leak detection (random event)
        leak_detected = random.random() < 0.001  # 0.1% chance
        
        # Valve position (normally 100% open)
        valve_position = 100.0
        if leak_detected:
            valve_position = random.uniform(20, 80)  # Partially close valve
        
        self.last_reading_time = current_time
        
        return {
            "timestamp": current_time.isoformat(),
            "device_id": self.config.device_id,
            "device_type": "water_meter",
            "location": self.config.location,
            "measurements": {
                "flow_rate": round(flow_rate, 2),
                "pressure": round(pressure, 2),
                "temperature": round(temperature, 1),
                "ph_level": round(ph_level, 2),
                "turbidity": round(turbidity, 2),
                "chlorine_level": round(chlorine_level, 2),
                "total_dissolved_solids": round(tds, 1),
                "cumulative_volume": round(self.cumulative_volume, 2),
                "leak_detected": leak_detected,
                "valve_position": round(valve_position, 1)
            },
            "quality_score": random.uniform(0.95, 1.0),
            "battery_level": random.uniform(85, 100),
            "signal_strength": random.randint(-80, -45)
        }
    
    def _publish_data(self, data: Dict[str, Any]):
        """Publish data to MQTT"""
        topic = f"urbanzen/devices/{self.config.device_id}/telemetry"
        payload = json.dumps(data)
        
        result = self.mqtt_client.publish(topic, payload, qos=1)
        if result.rc == mqtt.MQTT_ERR_SUCCESS:
            self.logger.info(f"Published telemetry data: flow_rate={data['measurements']['flow_rate']}L/min")
        else:
            self.logger.error(f"Failed to publish data, error code: {result.rc}")
    
    def start(self):
        """Start the simulator"""
        try:
            self.mqtt_client.connect(self.config.mqtt_broker, self.config.mqtt_port, 60)
            self.mqtt_client.loop_start()
            
            self.is_running = True
            self.logger.info(f"Water meter simulator started for device {self.config.device_id}")
            
            while self.is_running:
                try:
                    # Generate and publish sensor data
                    sensor_data = self._generate_sensor_data()
                    self._publish_data(sensor_data)
                    
                    # Wait for next measurement
                    time.sleep(self.config.publish_interval)
                    
                except KeyboardInterrupt:
                    self.logger.info("Received keyboard interrupt, stopping simulator")
                    break
                except Exception as e:
                    self.logger.error(f"Error in simulation loop: {e}")
                    time.sleep(5)  # Wait before retrying
            
        except Exception as e:
            self.logger.error(f"Failed to start simulator: {e}")
        finally:
            self.stop()
    
    def stop(self):
        """Stop the simulator"""
        self.is_running = False
        if self.mqtt_client:
            self.mqtt_client.loop_stop()
            self.mqtt_client.disconnect()
        self.logger.info("Water meter simulator stopped")


def main():
    """Main function to run the simulator"""
    config = WaterMeterConfig(
        device_id="WM001",
        location={"lat": 28.4595, "lng": 77.0266},
        mqtt_broker="localhost",
        mqtt_port=1883,
        publish_interval=60
    )
    
    simulator = WaterMeterSimulator(config)
    
    try:
        simulator.start()
    except KeyboardInterrupt:
        print("\nShutting down simulator...")
    except Exception as e:
        print(f"Simulator error: {e}")
    finally:
        simulator.stop()


if __name__ == "__main__":
    main()