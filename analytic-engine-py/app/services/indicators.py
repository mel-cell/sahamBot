import pandas as pd
import numpy as np

def calculate_rsi(df: pd.DataFrame, length: int = 14) -> float:
    """
    Menghitung Relative Strength Index (RSI) PURE PANDAS.
    Gak butuh library aneh-aneh.
    """
    if df.empty or 'Close' not in df.columns:
        return 0.0
    
    # Hitung selisih harga
    delta = df['Close'].diff()
    
    # Pisahkan untung (gain) dan rugi (loss)
    gain = (delta.where(delta > 0, 0)).rolling(window=length).mean()
    loss = (-delta.where(delta < 0, 0)).rolling(window=length).mean()

    # Hindari pembagian dengan nol
    rs = gain / loss.replace(0, np.nan)
    rsi = 100 - (100 / (1 + rs))
    
    # Isi NaN dengan 50 (Nutral) kalau data kurang
    rsi = rsi.fillna(50)
    
    return rsi.iloc[-1]

def calculate_macd(df: pd.DataFrame, fast: int = 12, slow: int = 26, signal: int = 9) -> dict:
    """
    Menghitung MACD Manual pakai Exponential Moving Average (EMA).
    """
    if df.empty or 'Close' not in df.columns:
        return {'macd': 0.0, 'signal': 0.0}
    
    # Hitung EMA Fast & Slow
    ema_fast = df['Close'].ewm(span=fast, adjust=False).mean()
    ema_slow = df['Close'].ewm(span=slow, adjust=False).mean()
    
    # MACD Line = Fast - Slow
    macd_line = ema_fast - ema_slow
    
    # Signal Line = EMA dari MACD Line
    signal_line = macd_line.ewm(span=signal, adjust=False).mean()
    
    return {
        'macd': macd_line.iloc[-1],
        'signal': signal_line.iloc[-1]
    }

def analyze_trend(rsi: float, macd: float, signal: float) -> str:
    """
    Logic analisa sederhana tapi masuk akal.
    """
    score = 0
    
    # Analisa RSI
    if rsi < 30: score += 1      # Murah
    elif rsi > 70: score -= 1    # Mahal
    
    # Analisa MACD (Momentum)
    if macd > signal: score += 1  # Golden Cross (Naik)
    elif macd < signal: score -= 1 # Dead Cross (Turun)
    
    # Kesimpulan
    if score >= 2: return "STRONG_BUY 🚀"
    if score == 1: return "BUY_POTENTIAL 📈"
    if score == 0: return "NEUTRAL 😐"
    if score == -1: return "SELL_POTENTIAL 📉"
    return "STRONG_SELL 🩸"
