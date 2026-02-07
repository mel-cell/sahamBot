import yfinance as yf
import pandas as pd
from typing import Optional

def get_stock_history(symbol: str, period: str = "3mo") -> Optional[pd.DataFrame]:
    """
    Mengambil data history saham dari Yahoo Finance.
    
    Args:
        symbol (str): Kode saham (e.g., 'BBRI.JK', 'NVDA')
        period (str): Durasi data (e.g., '1mo', '3mo', '1y')
        
    Returns:
        pd.DataFrame: DataFrame berisi OHLCV
    """
    try:
        # Menghapus suffix .JK jika user lupa, tapi yfinance butuh .JK untuk saham Indo
        if not symbol.endswith(".JK") and len(symbol) == 4 and symbol.isalpha(): 
            # Asumsi sederhana jika 4 huruf kemungkinan saham Indo (perlu validasi lebih baik nanti)
            # Tapi untuk sekarang kita biarkan user input raw symbol dulu agar fleksibel (US Stocks)
            pass

        ticker = yf.Ticker(symbol)
        df = ticker.history(period=period)
        
        if df.empty:
            print(f"Warning: No data found for {symbol}")
            return None
            
        return df
    except Exception as e:
        print(f"Error fetching data for {symbol}: {e}")
        return None

def get_current_price(symbol: str) -> Optional[dict]:
    """
    Mengambil harga terakhir dan info dasar.
    """
    try:
        ticker = yf.Ticker(symbol)
        # fast_info lebih cepat dari .info
        price = ticker.fast_info['last_price']
        prev_close = ticker.fast_info['previous_close']
        change_pct = ((price - prev_close) / prev_close) * 100
        
        return {
            "symbol": symbol,
            "price": price,
            "change_pct": change_pct,
            "volume": 0 # to do: ambil volume real-time jika perlu
        }
    except Exception as e:
        print(f"Error fetching price for {symbol}: {e}")
        return None
