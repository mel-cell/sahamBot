from transformers import AutoModelForCausalLM, AutoTokenizer
import torch

# Model Name: Qwen2.5-0.5B-Instruct (Super tiny but smart)
MODEL_NAME = "Qwen/Qwen2.5-0.5B-Instruct"

class AIAnalyst:
    def __init__(self):
        print("🤖 Loading AI Model (this might take a while first time)...")
        try:
            self.tokenizer = AutoTokenizer.from_pretrained(MODEL_NAME, trust_remote_code=True)
            self.model = AutoModelForCausalLM.from_pretrained(
                MODEL_NAME, 
                torch_dtype=torch.float32, # Pakai float32 biar aman di CPU biasa
                device_map="cpu", 
                trust_remote_code=True
            )
            print("✅ AI Model Loaded Successfully!")
        except Exception as e:
            print(f"❌ Failed to load AI: {e}")
            self.model = None

    def analyze(self, symbol: str, price: float, rsi: float, macd_signal: str) -> str:
        if not self.model:
            return "AI Brain not loaded."

        # Prompt Engineering Simple
        prompt = f"""You are a friendly Stock Market Analyst.
        Analyze stock {symbol}.
        Data:
        - Price: {price}
        - RSI: {rsi:.2f} (Overbought > 70, Oversold < 30)
        - Trend: {macd_signal}
        
        Is it good to buy based on this? Answer in 2 short sentences.
        Use simple language."""

        inputs = self.tokenizer(prompt, return_tensors="pt")
        
        # Generate Response
        with torch.no_grad():
            outputs = self.model.generate(
                **inputs, 
                max_new_tokens=100,
                do_sample=True,
                temperature=0.7
            )
        
        response = self.tokenizer.decode(outputs[0], skip_special_tokens=True)
        
        # Bersihkan prompt dari jawaban (karena model CausalLM mengulang input)
        clean_response = response.replace(prompt, "").strip()
        return clean_response

# Singleton Instance (Biar model cuma di-load sekali pas server nyala)
ai_engine = AIAnalyst()
