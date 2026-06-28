import requests
import base64
import numpy as np
import matplotlib.pyplot as plt
from scipy import signal

API_URL = "http://localhost:8080/frames"

def fetch_frames(from_id, to_id):
    params = {"from": from_id, "to": to_id}
    resp = requests.get(API_URL, params=params)
    resp.raise_for_status()
    frames = resp.json()
    return frames

def process_frame(frame):
    # Декодируем base64 в байты
    raw_bytes = base64.b64decode(frame["data"])
    # Предположим, что данные – массив float32
    data = np.frombuffer(raw_bytes, dtype=np.float32)
    return data

def main():
    # Запросим последние 10 кадров (например, с 1 по 10)
    frames = fetch_frames(1, 10)
    all_data = []
    timestamps = []
    for f in frames:
        data = process_frame(f)
        all_data.append(data)
        timestamps.append(f["timestamp"])
    
    # Объединим все кадры в один массив
    combined = np.concatenate(all_data)
    
    # Фильтрация шума (например, медианный фильтр)
    filtered = signal.medfilt(combined, kernel_size=5)
    
    # Поиск аномалий (например, значения > 3 sigma)
    mean = np.mean(filtered)
    std = np.std(filtered)
    anomalies = np.where(np.abs(filtered - mean) > 3 * std)[0]
    
    # Построение графика
    plt.figure(figsize=(12, 6))
    plt.plot(combined, label="Raw data", alpha=0.5)
    plt.plot(filtered, label="Filtered", linewidth=2)
    plt.scatter(anomalies, filtered[anomalies], color='red', label="Anomalies")
    plt.xlabel("Sample index")
    plt.ylabel("Amplitude")
    plt.title("Telescope Data Analysis")
    plt.legend()
    plt.grid(True)
    plt.show()
    
    # Вывод статистики
    print(f"Total samples: {len(combined)}")
    print(f"Anomalies found: {len(anomalies)}")

if __name__ == "__main__":
    main()