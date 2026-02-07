from pydantic import BaseModel
from typing import List, Optional

class StockData(BaseModel):
    symbol: str
    price: float
    change_percent: float
    volume: int
    last_updated: str

class TechnicalIndicators(BaseModel):
    rsi: float
    macd: float
    macd_signal: float
    signal: str  # BUY, SELL, NEUTRAL

class AnalysisResult(BaseModel):
    symbol: str
    current_price: float
    technical: TechnicalIndicators
    summary: str  # Narasi singkat hasil analisa gabungan
