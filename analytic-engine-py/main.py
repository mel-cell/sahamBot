from fastapi import FastAPI, HTTPException
from typing import Optional
import uvicorn
from app.services.market_data import get_stock_history, get_current_price
from app.services.indicators import calculate_rsi, calculate_macd, analyze_trend
from app.models.market import AnalysisResult, TechnicalIndicators

app = FastAPI(title="Saham Analytic Engine (Python)", version="1.0")

@app.get("/")
def read_root():
    return {"status": "ok", "service": "Analytic Engine v1"}

@app.get("/analyze/{symbol}", response_model=AnalysisResult)
def analyze_stock(symbol: str):
    """
    Menganalisa saham berdasarkan simbol.
    Contoh: /analyze/BBRI.JK
    """
    try:
        # 1. Ambil Data Harga Terkini
        stock_data = get_stock_history(symbol)
        
        if stock_data is None or stock_data.empty:
            raise HTTPException(status_code=404, detail="Stock not found or data unavailable")
            
        current_data = get_current_price(symbol)
        if current_data is None:
             raise HTTPException(status_code=500, detail="Failed to fetch current price")

        # 2. Hitung Indikator (Math)
        rsi_val = calculate_rsi(stock_data)
        macd_val = calculate_macd(stock_data)
        
        # 3. Tentukan Sinyal (Logic/AI Placeholder)
        # Pass individual float values to analyze_trend
        trend = analyze_trend(rsi_val, macd_val['macd'], macd_val['signal'])
        
        # 4. Generate AI Summary (Real AI now!)
        ai_summary = "AI model loading..."
        try:
            from app.services.llm import ai_engine
            ai_summary = ai_engine.analyze(symbol, current_data['price'], rsi_val, trend)
        except Exception as ai_err:
            print(f"AI Error: {ai_err}")
            ai_summary = f"AI Failed: {ai_err}"

        # 5. Return Format JSON
        return AnalysisResult(
            symbol=symbol,
            current_price=float(current_data['price']),
            technical=TechnicalIndicators(
                rsi=float(rsi_val),
                macd=float(macd_val['macd']),
                macd_signal=float(macd_val['signal']),
                signal=trend
            ),
            summary=ai_summary
        )

    except Exception as e:
        # Print error to console for debugging
        print(f"Error analyzing {symbol}: {e}")
        raise HTTPException(status_code=500, detail=str(e))

if __name__ == "__main__":
    uvicorn.run("main:app", host="0.0.0.0", port=8000, reload=True)
