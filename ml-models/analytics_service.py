import os
import logging
import asyncio
import json
import numpy as np
import pandas as pd
from datetime import datetime, timedelta
from typing import Dict, List, Optional, Tuple
import joblib
from sklearn.ensemble import IsolationForest, RandomForestRegressor
from sklearn.preprocessing import StandardScaler
from sklearn.model_selection import train_test_split
import tensorflow as tf
from tensorflow import keras
import asyncpg
import aioredis
from confluent_kafka import Consumer, Producer
from prometheus_client import Counter, Histogram, Gauge, start_http_server

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# Prometheus metrics
anomaly_counter = Counter('anomalies_detected_total', 'Total anomalies detected', ['device_type', 'anomaly_type'])
prediction_histogram = Histogram('prediction_duration_seconds', 'Time spent on predictions')
model_accuracy_gauge = Gauge('model_accuracy', 'Current model accuracy', ['model_type'])

class AnalyticsService:
    def __init__(self, config: Dict):
        self.config = config
        self.db_pool = None
        self.redis = None
        self.kafka_consumer = None
        self.kafka_producer = None
        
        # ML Models
        self.anomaly_detectors = {}
        self.consumption_predictors = {}
        self.traffic_predictor = None
        self.scalers = {}
        
        # Model paths
        self.model_dir = config.get('model_dir', './models')
        os.makedirs(self.model_dir, exist_ok=True)
        
    async def initialize(self):
        """Initialize all connections and load models"""
        try:
            # Database connection
            self.db_pool = await asyncpg.create_pool(
                host=self.config['database']['host'],
                port=self.config['database']['port'],
                user=self.config['database']['user'],
                password=self.config['database']['password'],
                database=self.config['database']['name'],
                min_size=5,
                max_size=20
            )
            
            # Redis connection
            self.redis = await aioredis.create_redis_pool(
                f"redis://{self.config['redis']['host']}:{self.config['redis']['port']}"
            )
            
            # Kafka setup
            self._setup_kafka()
            
            # Load or train models
            await self._load_models()
            
            logger.info("Analytics service initialized successfully")
            
        except Exception as e:
            logger.error(f"Failed to initialize analytics service: {e}")
            raise
    
    def _setup_kafka(self):
        """Setup Kafka consumer and producer"""
        consumer_config = {
            'bootstrap.servers': ','.join(self.config['kafka']['brokers']),
            'group.id': 'analytics-service',
            'auto.offset.reset': 'latest'
        }
        
        producer_config = {
            'bootstrap.servers': ','.join(self.config['kafka']['brokers']),
        }
        
        self.kafka_consumer = Consumer(consumer_config)
        self.kafka_producer = Producer(producer_config)
        
        # Subscribe to topics
        topics = ['device-telemetry', 'analytics-requests']
        self.kafka_consumer.subscribe(topics)
    
    async def _load_models(self):
        """Load pre-trained models or train new ones"""
        try:
            # Load anomaly detection models
            for device_type in ['water_sensor', 'electricity_meter', 'traffic_camera']:
                model_path = os.path.join(self.model_dir, f'{device_type}_anomaly_detector.joblib')
                scaler_path = os.path.join(self.model_dir, f'{device_type}_scaler.joblib')
                
                if os.path.exists(model_path) and os.path.exists(scaler_path):
                    self.anomaly_detectors[device_type] = joblib.load(model_path)
                    self.scalers[device_type] = joblib.load(scaler_path)
                    logger.info(f"Loaded anomaly detector for {device_type}")
                else:
                    await self._train_anomaly_detector(device_type)
            
            # Load consumption prediction models
            for utility_type in ['water', 'electricity']:
                model_path = os.path.join(self.model_dir, f'{utility_type}_consumption_predictor.joblib')
                if os.path.exists(model_path):
                    self.consumption_predictors[utility_type] = joblib.load(model_path)
                    logger.info(f"Loaded consumption predictor for {utility_type}")
                else:
                    await self._train_consumption_predictor(utility_type)
            
            # Load traffic prediction model
            traffic_model_path = os.path.join(self.model_dir, 'traffic_predictor.h5')
            if os.path.exists(traffic_model_path):
                self.traffic_predictor = keras.models.load_model(traffic_model_path)
                logger.info("Loaded traffic prediction model")
            else:
                await self._train_traffic_predictor()
                
        except Exception as e:
            logger.error(f"Failed to load models: {e}")
            raise
    
    async def _train_anomaly_detector(self, device_type: str):
        """Train anomaly detection model for specific device type"""
        logger.info(f"Training anomaly detector for {device_type}")
        
        # Fetch training data
        data = await self._fetch_training_data(device_type, days=30)
        
        if len(data) < 100:
            logger.warning(f"Insufficient data for training {device_type} anomaly detector")
            return
        
        # Prepare features
        features = self._extract_features(data, device_type)
        
        # Scale features
        scaler = StandardScaler()
        scaled_features = scaler.fit_transform(features)
        
        # Train Isolation Forest
        model = IsolationForest(contamination=0.1, random_state=42)
        model.fit(scaled_features)
        
        # Save model and scaler
        model_path = os.path.join(self.model_dir, f'{device_type}_anomaly_detector.joblib')
        scaler_path = os.path.join(self.model_dir, f'{device_type}_scaler.joblib')
        
        joblib.dump(model, model_path)
        joblib.dump(scaler, scaler_path)
        
        self.anomaly_detectors[device_type] = model
        self.scalers[device_type] = scaler
        
        logger.info(f"Trained and saved anomaly detector for {device_type}")
    
    async def _train_consumption_predictor(self, utility_type: str):
        """Train consumption prediction model"""
        logger.info(f"Training consumption predictor for {utility_type}")
        
        # Fetch historical consumption data
        data = await self._fetch_consumption_data(utility_type, days=90)
        
        if len(data) < 500:
            logger.warning(f"Insufficient data for training {utility_type} consumption predictor")
            return
        
        # Prepare features and target
        features, target = self._prepare_consumption_features(data)
        
        # Split data
        X_train, X_test, y_train, y_test = train_test_split(
            features, target, test_size=0.2, random_state=42
        )
        
        # Train Random Forest model
        model = RandomForestRegressor(n_estimators=100, random_state=42)
        model.fit(X_train, y_train)
        
        # Evaluate model
        accuracy = model.score(X_test, y_test)
        model_accuracy_gauge.labels(model_type=f'{utility_type}_consumption').set(accuracy)
        
        # Save model
        model_path = os.path.join(self.model_dir, f'{utility_type}_consumption_predictor.joblib')
        joblib.dump(model, model_path)
        
        self.consumption_predictors[utility_type] = model
        
        logger.info(f"Trained consumption predictor for {utility_type} with accuracy: {accuracy:.3f}")
    
    async def _train_traffic_predictor(self):
        """Train traffic prediction neural network"""
        logger.info("Training traffic prediction model")
        
        # Fetch traffic data
        data = await self._fetch_traffic_data(days=60)
        
        if len(data) < 1000:
            logger.warning("Insufficient data for training traffic predictor")
            return
        
        # Prepare sequences for LSTM
        X, y = self._prepare_traffic_sequences(data)
        
        # Split data
        split_idx = int(0.8 * len(X))
        X_train, X_test = X[:split_idx], X[split_idx:]
        y_train, y_test = y[:split_idx], y[split_idx:]
        
        # Build LSTM model
        model = keras.Sequential([
            keras.layers.LSTM(50, return_sequences=True, input_shape=(X.shape[1], X.shape[2])),
            keras.layers.Dropout(0.2),
            keras.layers.LSTM(50, return_sequences=False),
            keras.layers.Dropout(0.2),
            keras.layers.Dense(25),
            keras.layers.Dense(1)
        ])
        
        model.compile(optimizer='adam', loss='mse', metrics=['mae'])
        
        # Train model
        history = model.fit(
            X_train, y_train,
            batch_size=32,
            epochs=50,
            validation_data=(X_test, y_test),
            verbose=0
        )
        
        # Save model
        model_path = os.path.join(self.model_dir, 'traffic_predictor.h5')
        model.save(model_path)
        
        self.traffic_predictor = model
        
        # Update accuracy metric
        test_loss = model.evaluate(X_test, y_test, verbose=0)[0]
        model_accuracy_gauge.labels(model_type='traffic_prediction').set(1 / (1 + test_loss))
        
        logger.info("Trained traffic prediction model")
    
    async def _fetch_training_data(self, device_type: str, days: int) -> pd.DataFrame:
        """Fetch training data for anomaly detection"""
        query = """
            SELECT device_id, timestamp, metrics, metadata
            FROM device_telemetry
            WHERE device_type = $1
            AND timestamp >= NOW() - INTERVAL '%d days'
            ORDER BY timestamp
        """ % days
        
        async with self.db_pool.acquire() as conn:
            rows = await conn.fetch(query, device_type)
        
        return pd.DataFrame([
            {
                'device_id': row['device_id'],
                'timestamp': row['timestamp'],
                **json.loads(row['metrics'])
            }
            for row in rows
        ])
    
    def _extract_features(self, data: pd.DataFrame, device_type: str) -> np.ndarray:
        """Extract features for anomaly detection"""
        if device_type == 'water_sensor':
            feature_cols = ['flow_rate', 'pressure', 'ph_level', 'turbidity']
        elif device_type == 'electricity_meter':
            feature_cols = ['voltage', 'current', 'power', 'frequency']
        elif device_type == 'traffic_camera':
            feature_cols = ['vehicle_count', 'avg_speed', 'congestion_level']
        else:
            # Use all numeric columns
            feature_cols = data.select_dtypes(include=[np.number]).columns.tolist()
        
        # Fill missing values and return features
        features = data[feature_cols].fillna(0)
        return features.values
    
    async def detect_anomalies(self, device_data: Dict) -> Optional[Dict]:
        """Detect anomalies in real-time device data"""
        device_type = device_data.get('device_type')
        
        if device_type not in self.anomaly_detectors:
            return None
        
        try:
            # Extract features
            features = self._extract_single_features(device_data, device_type)
            
            # Scale features
            scaler = self.scalers[device_type]
            scaled_features = scaler.transform([features])
            
            # Predict anomaly
            model = self.anomaly_detectors[device_type]
            anomaly_score = model.decision_function(scaled_features)[0]
            is_anomaly = model.predict(scaled_features)[0] == -1
            
            if is_anomaly:
                anomaly_counter.labels(device_type=device_type, anomaly_type='statistical').inc()
                
                return {
                    'device_id': device_data['device_id'],
                    'device_type': device_type,
                    'anomaly_score': float(anomaly_score),
                    'timestamp': datetime.utcnow().isoformat(),
                    'severity': self._calculate_severity(anomaly_score),
                    'description': f'Statistical anomaly detected in {device_type}',
                    'metrics': device_data.get('metrics', {})
                }
            
            return None
            
        except Exception as e:
            logger.error(f"Error detecting anomalies: {e}")
            return None
    
    def _extract_single_features(self, device_data: Dict, device_type: str) -> List[float]:
        """Extract features from a single device data point"""
        metrics = device_data.get('metrics', {})
        
        if device_type == 'water_sensor':
            return [
                metrics.get('flow_rate', 0),
                metrics.get('pressure', 0),
                metrics.get('ph_level', 7),
                metrics.get('turbidity', 0)
            ]
        elif device_type == 'electricity_meter':
            return [
                metrics.get('voltage', 0),
                metrics.get('current', 0),
                metrics.get('power', 0),
                metrics.get('frequency', 50)
            ]
        elif device_type == 'traffic_camera':
            return [
                metrics.get('vehicle_count', 0),
                metrics.get('avg_speed', 0),
                metrics.get('congestion_level', 0)
            ]
        else:
            return list(metrics.values())[:10]  # Limit to first 10 metrics
    
    def _calculate_severity(self, anomaly_score: float) -> str:
        """Calculate anomaly severity based on score"""
        if anomaly_score < -0.5:
            return 'critical'
        elif anomaly_score < -0.3:
            return 'high'
        elif anomaly_score < -0.1:
            return 'medium'
        else:
            return 'low'
    
    async def predict_consumption(self, utility_type: str, user_id: str, days_ahead: int = 7) -> Dict:
        """Predict future consumption for a user"""
        if utility_type not in self.consumption_predictors:
            return {'error': f'No predictor available for {utility_type}'}
        
        try:
            # Fetch user's historical data
            historical_data = await self._fetch_user_consumption(user_id, utility_type, days=30)
            
            if len(historical_data) < 7:
                return {'error': 'Insufficient historical data'}
            
            # Prepare features
            features = self._prepare_prediction_features(historical_data, days_ahead)
            
            # Make prediction
            model = self.consumption_predictors[utility_type]
            predictions = model.predict(features)
            
            return {
                'utility_type': utility_type,
                'user_id': user_id,
                'predictions': [
                    {
                        'date': (datetime.now() + timedelta(days=i)).isoformat(),
                        'predicted_consumption': float(pred),
                        'confidence': 0.85  # Placeholder confidence
                    }
                    for i, pred in enumerate(predictions[:days_ahead], 1)
                ]
            }
            
        except Exception as e:
            logger.error(f"Error predicting consumption: {e}")
            return {'error': str(e)}
    
    async def process_kafka_messages(self):
        """Process incoming Kafka messages"""
        while True:
            try:
                msg = self.kafka_consumer.poll(timeout=1.0)
                
                if msg is None:
                    continue
                
                if msg.error():
                    logger.error(f"Kafka error: {msg.error()}")
                    continue
                
                # Process message based on topic
                topic = msg.topic()
                data = json.loads(msg.value().decode('utf-8'))
                
                if topic == 'device-telemetry':
                    await self._process_telemetry_message(data)
                elif topic == 'analytics-requests':
                    await self._process_analytics_request(data)
                
            except Exception as e:
                logger.error(f"Error processing Kafka message: {e}")
            
            await asyncio.sleep(0.1)
    
    async def _process_telemetry_message(self, data: Dict):
        """Process device telemetry data"""
        # Check for anomalies
        anomaly = await self.detect_anomalies(data)
        
        if anomaly:
            # Send anomaly alert
            await self._send_anomaly_alert(anomaly)
        
        # Store processed data in cache for quick access
        cache_key = f"device:{data['device_id']}:latest"
        await self.redis.setex(cache_key, 3600, json.dumps(data))
    
    async def _send_anomaly_alert(self, anomaly: Dict):
        """Send anomaly alert to appropriate channels"""
        alert_message = {
            'type': 'anomaly_detected',
            'severity': anomaly['severity'],
            'device_id': anomaly['device_id'],
            'device_type': anomaly['device_type'],
            'description': anomaly['description'],
            'timestamp': anomaly['timestamp'],
            'anomaly_score': anomaly['anomaly_score']
        }
        
        # Send to alerts topic
        self.kafka_producer.produce(
            'alerts',
            key=anomaly['device_id'],
            value=json.dumps(alert_message)
        )
        
        self.kafka_producer.flush()
        
        logger.info(f"Sent anomaly alert for device {anomaly['device_id']}")
    
    async def run(self):
        """Main service loop"""
        try:
            await self.initialize()
            
            # Start Prometheus metrics server
            start_http_server(8000)
            
            # Start Kafka message processing
            await self.process_kafka_messages()
            
        except Exception as e:
            logger.error(f"Service error: {e}")
            raise
        finally:
            await self.cleanup()
    
    async def cleanup(self):
        """Cleanup resources"""
        if self.db_pool:
            await self.db_pool.close()
        
        if self.redis:
            self.redis.close()
        
        if self.kafka_consumer:
            self.kafka_consumer.close()
        
        if self.kafka_producer:
            self.kafka_producer.flush()

# Configuration
config = {
    'database': {
        'host': os.getenv('TIMESCALEDB_HOST', 'localhost'),
        'port': int(os.getenv('TIMESCALEDB_PORT', 5432)),
        'user': os.getenv('TIMESCALEDB_USER', 'postgres'),
        'password': os.getenv('TIMESCALEDB_PASSWORD', 'password'),
        'name': os.getenv('TIMESCALEDB_NAME', 'urbanzen')
    },
    'redis': {
        'host': os.getenv('REDIS_HOST', 'localhost'),
        'port': int(os.getenv('REDIS_PORT', 6379))
    },
    'kafka': {
        'brokers': os.getenv('KAFKA_BROKERS', 'localhost:9092').split(',')
    },
    'model_dir': os.getenv('MODEL_DIR', './models')
}

if __name__ == '__main__':
    service = AnalyticsService(config)
    asyncio.run(service.run())