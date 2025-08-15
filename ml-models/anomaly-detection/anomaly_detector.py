#!/usr/bin/env python3
"""
Anomaly Detection Model for UrbanZen IoT Platform
Detects anomalies in water consumption, electricity usage, and environmental data
"""

import numpy as np
import pandas as pd
from sklearn.ensemble import IsolationForest
from sklearn.preprocessing import StandardScaler
from sklearn.model_selection import train_test_split
from sklearn.metrics import classification_report, confusion_matrix
import joblib
import logging
from datetime import datetime, timedelta
from typing import Dict, List, Tuple, Any
import json


class AnomalyDetector:
    """
    Anomaly detection using Isolation Forest algorithm
    Optimized for IoT sensor data with temporal patterns
    """
    
    def __init__(self, contamination=0.1, random_state=42):
        """
        Initialize the anomaly detector
        
        Args:
            contamination: Expected proportion of anomalies in the data
            random_state: Random seed for reproducibility
        """
        self.contamination = contamination
        self.random_state = random_state
        self.model = IsolationForest(
            contamination=contamination,
            random_state=random_state,
            n_estimators=200,
            max_samples='auto',
            behaviour='new'
        )
        self.scaler = StandardScaler()
        self.feature_columns = []
        self.is_trained = False
        
        # Setup logging
        logging.basicConfig(level=logging.INFO)
        self.logger = logging.getLogger(__name__)
    
    def prepare_features(self, data: pd.DataFrame) -> pd.DataFrame:
        """
        Prepare features for anomaly detection
        
        Args:
            data: Raw sensor data DataFrame
            
        Returns:
            DataFrame with engineered features
        """
        features = data.copy()
        
        # Time-based features
        if 'timestamp' in features.columns:
            features['timestamp'] = pd.to_datetime(features['timestamp'])
            features['hour'] = features['timestamp'].dt.hour
            features['day_of_week'] = features['timestamp'].dt.dayofweek
            features['month'] = features['timestamp'].dt.month
            features['is_weekend'] = features['day_of_week'].isin([5, 6]).astype(int)
            
            # Cyclical encoding for time features
            features['hour_sin'] = np.sin(2 * np.pi * features['hour'] / 24)
            features['hour_cos'] = np.cos(2 * np.pi * features['hour'] / 24)
            features['dow_sin'] = np.sin(2 * np.pi * features['day_of_week'] / 7)
            features['dow_cos'] = np.cos(2 * np.pi * features['day_of_week'] / 7)
        
        # Rolling statistics for sensor values
        sensor_columns = ['flow_rate', 'pressure', 'temperature', 'ph_level', 
                         'turbidity', 'chlorine_level', 'power_consumption',
                         'voltage', 'current', 'pm25', 'pm10', 'co2']
        
        for col in sensor_columns:
            if col in features.columns:
                # Rolling mean and std
                features[f'{col}_rolling_mean_6h'] = features[col].rolling(
                    window=6, min_periods=1
                ).mean()
                features[f'{col}_rolling_std_6h'] = features[col].rolling(
                    window=6, min_periods=1
                ).std()
                
                # Difference from rolling mean
                features[f'{col}_diff_from_mean'] = (
                    features[col] - features[f'{col}_rolling_mean_6h']
                )
                
                # Rate of change
                features[f'{col}_rate_of_change'] = features[col].pct_change()
        
        # Device-specific features
        if 'device_type' in features.columns:
            features = pd.get_dummies(features, columns=['device_type'], prefix='device')
        
        # Quality and connectivity features
        quality_features = ['quality_score', 'battery_level', 'signal_strength']
        for col in quality_features:
            if col in features.columns:
                # Normalize quality features
                features[f'{col}_normalized'] = (
                    features[col] - features[col].min()
                ) / (features[col].max() - features[col].min() + 1e-8)
        
        # Remove non-numeric columns
        numeric_columns = features.select_dtypes(include=[np.number]).columns
        features = features[numeric_columns]
        
        # Handle missing values
        features = features.fillna(features.median())
        
        return features
    
    def train(self, data: pd.DataFrame, labels: np.ndarray = None) -> Dict[str, Any]:
        """
        Train the anomaly detection model
        
        Args:
            data: Training data DataFrame
            labels: Optional labels for evaluation (1 for normal, -1 for anomaly)
            
        Returns:
            Training metrics dictionary
        """
        self.logger.info("Preparing features for training...")
        features = self.prepare_features(data)
        
        # Store feature columns for later use
        self.feature_columns = features.columns.tolist()
        
        # Scale features
        self.logger.info("Scaling features...")
        scaled_features = self.scaler.fit_transform(features)
        
        # Train model
        self.logger.info("Training Isolation Forest model...")
        self.model.fit(scaled_features)
        self.is_trained = True
        
        # Generate predictions for evaluation
        predictions = self.model.predict(scaled_features)
        anomaly_scores = self.model.decision_function(scaled_features)
        
        # Calculate metrics if labels are provided
        metrics = {}
        if labels is not None:
            metrics = {
                'classification_report': classification_report(labels, predictions),
                'confusion_matrix': confusion_matrix(labels, predictions).tolist(),
                'anomaly_ratio': np.sum(predictions == -1) / len(predictions),
                'mean_anomaly_score': np.mean(anomaly_scores),
                'std_anomaly_score': np.std(anomaly_scores)
            }
        else:
            metrics = {
                'anomaly_ratio': np.sum(predictions == -1) / len(predictions),
                'mean_anomaly_score': np.mean(anomaly_scores),
                'std_anomaly_score': np.std(anomaly_scores)
            }
        
        self.logger.info(f"Training completed. Anomaly ratio: {metrics['anomaly_ratio']:.4f}")
        return metrics
    
    def predict(self, data: pd.DataFrame) -> Dict[str, Any]:
        """
        Predict anomalies in new data
        
        Args:
            data: New data DataFrame
            
        Returns:
            Dictionary containing predictions and scores
        """
        if not self.is_trained:
            raise ValueError("Model must be trained before making predictions")
        
        # Prepare features
        features = self.prepare_features(data)
        
        # Ensure same features as training
        missing_features = set(self.feature_columns) - set(features.columns)
        if missing_features:
            self.logger.warning(f"Missing features: {missing_features}")
            for feature in missing_features:
                features[feature] = 0
        
        # Reorder columns to match training
        features = features[self.feature_columns]
        
        # Scale features
        scaled_features = self.scaler.transform(features)
        
        # Make predictions
        predictions = self.model.predict(scaled_features)
        anomaly_scores = self.model.decision_function(scaled_features)
        
        # Convert to probabilities (higher means more anomalous)
        anomaly_probabilities = 1 / (1 + np.exp(anomaly_scores))
        
        return {
            'predictions': predictions.tolist(),
            'anomaly_scores': anomaly_scores.tolist(),
            'anomaly_probabilities': anomaly_probabilities.tolist(),
            'anomaly_count': np.sum(predictions == -1),
            'total_count': len(predictions)
        }
    
    def save_model(self, filepath: str):
        """Save the trained model to disk"""
        if not self.is_trained:
            raise ValueError("Model must be trained before saving")
        
        model_data = {
            'model': self.model,
            'scaler': self.scaler,
            'feature_columns': self.feature_columns,
            'contamination': self.contamination,
            'random_state': self.random_state,
            'training_date': datetime.now().isoformat()
        }
        
        joblib.dump(model_data, filepath)
        self.logger.info(f"Model saved to {filepath}")
    
    def load_model(self, filepath: str):
        """Load a trained model from disk"""
        model_data = joblib.load(filepath)
        
        self.model = model_data['model']
        self.scaler = model_data['scaler']
        self.feature_columns = model_data['feature_columns']
        self.contamination = model_data['contamination']
        self.random_state = model_data['random_state']
        self.is_trained = True
        
        self.logger.info(f"Model loaded from {filepath}")
        self.logger.info(f"Training date: {model_data.get('training_date', 'Unknown')}")


def generate_synthetic_data(n_samples=10000) -> Tuple[pd.DataFrame, np.ndarray]:
    """
    Generate synthetic IoT sensor data for testing
    
    Args:
        n_samples: Number of samples to generate
        
    Returns:
        Tuple of (features DataFrame, labels array)
    """
    np.random.seed(42)
    
    # Generate timestamps
    start_date = datetime.now() - timedelta(days=30)
    timestamps = [start_date + timedelta(minutes=i*5) for i in range(n_samples)]
    
    data = []
    labels = []
    
    for i, timestamp in enumerate(timestamps):
        # Normal patterns based on time of day
        hour = timestamp.hour
        if 6 <= hour <= 9 or 18 <= hour <= 21:  # Peak hours
            base_flow = 25.0
        elif 10 <= hour <= 17:  # Moderate hours
            base_flow = 15.0
        else:  # Low hours
            base_flow = 5.0
        
        # Normal data (90%)
        if np.random.random() < 0.9:
            flow_rate = base_flow + np.random.normal(0, 3)
            pressure = 2.5 + np.random.normal(0, 0.2)
            temperature = 20 + np.random.normal(0, 2)
            ph_level = 7.2 + np.random.normal(0, 0.3)
            turbidity = 0.5 + np.random.uniform(-0.2, 0.2)
            chlorine_level = 1.0 + np.random.normal(0, 0.1)
            quality_score = 0.98 + np.random.uniform(-0.03, 0.02)
            battery_level = 90 + np.random.uniform(-5, 5)
            signal_strength = -60 + np.random.uniform(-10, 10)
            label = 1  # Normal
        else:
            # Anomalous data (10%)
            if np.random.random() < 0.5:
                # High consumption anomaly
                flow_rate = base_flow * (3 + np.random.uniform(0, 2))
                pressure = 2.5 * 0.7  # Pressure drop
            else:
                # Low consumption anomaly (leak)
                flow_rate = base_flow * 0.1
                pressure = 2.5 * 1.3  # Pressure spike
            
            temperature = 20 + np.random.normal(0, 5)  # More variation
            ph_level = 7.2 + np.random.normal(0, 1)  # More variation
            turbidity = 0.5 + np.random.uniform(0, 2)  # Higher turbidity
            chlorine_level = 1.0 + np.random.normal(0, 0.5)  # More variation
            quality_score = 0.85 + np.random.uniform(-0.1, 0.1)  # Lower quality
            battery_level = 50 + np.random.uniform(-20, 40)  # More variation
            signal_strength = -60 + np.random.uniform(-20, 20)  # More variation
            label = -1  # Anomaly
        
        data.append({
            'timestamp': timestamp,
            'device_id': f'WM{i%10:03d}',
            'device_type': 'water_meter',
            'flow_rate': max(0, flow_rate),
            'pressure': max(0, pressure),
            'temperature': temperature,
            'ph_level': max(0, min(14, ph_level)),
            'turbidity': max(0, turbidity),
            'chlorine_level': max(0, chlorine_level),
            'quality_score': max(0, min(1, quality_score)),
            'battery_level': max(0, min(100, battery_level)),
            'signal_strength': signal_strength
        })
        labels.append(label)
    
    return pd.DataFrame(data), np.array(labels)


def main():
    """Main function for testing the anomaly detector"""
    # Generate synthetic data
    print("Generating synthetic data...")
    data, labels = generate_synthetic_data(5000)
    
    # Split data
    train_data, test_data, train_labels, test_labels = train_test_split(
        data, labels, test_size=0.3, random_state=42, stratify=labels
    )
    
    # Initialize and train detector
    detector = AnomalyDetector(contamination=0.1)
    
    print("Training anomaly detector...")
    metrics = detector.train(train_data, train_labels)
    print(f"Training metrics: {json.dumps(metrics, indent=2)}")
    
    # Test detector
    print("Testing on new data...")
    predictions = detector.predict(test_data)
    print(f"Test predictions: {json.dumps(predictions, indent=2)}")
    
    # Save model
    model_path = "anomaly_detector_model.joblib"
    detector.save_model(model_path)
    print(f"Model saved to {model_path}")
    
    # Load and test saved model
    new_detector = AnomalyDetector()
    new_detector.load_model(model_path)
    test_predictions = new_detector.predict(test_data.head(10))
    print(f"Loaded model predictions: {json.dumps(test_predictions, indent=2)}")


if __name__ == "__main__":
    main()