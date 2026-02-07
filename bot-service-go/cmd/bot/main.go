package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	tele "gopkg.in/telebot.v3"
)

// --- Structs for Analyic Engine Response ---
type AnalysisResult struct {
	Symbol       string              `json:"symbol"`
	CurrentPrice float64             `json:"current_price"`
	Technical    TechnicalIndicators `json:"technical"`
	Summary      string              `json:"summary"`
}

type TechnicalIndicators struct {
	RSI        float64 `json:"rsi"`
	MACD       float64 `json:"macd"`
	MACDSignal float64 `json:"macd_signal"`
	Signal     string  `json:"signal"`
}

func main() {
	// 1. Load Environment Variables from root .env or bot-service-go/.env
	// Try loading from relative path first
	if err := godotenv.Load(".env"); err != nil {
		// If fails, try one level up (if running from root)
		if err := godotenv.Load("../.env"); err != nil {
			log.Println("Warning: Error loading .env file, checking system env")
		}
	}

	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	analyticURL := os.Getenv("ANALYTIC_ENGINE_URL")

	if botToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN is not set in .env")
	}
	if analyticURL == "" {
		analyticURL = "http://localhost:8000" // Default fallback
	}

	// 2. Setup Bot Settings
	pref := tele.Settings{
		Token:  botToken,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		log.Fatal(err)
		return
	}

	log.Println("🤖 Bot Started! Listening for messages...")

	// 3. Define Handlers

	// /start
	b.Handle("/start", func(c tele.Context) error {
		return c.Send("Halo! Saya Bot Analisa Saham 📈.\n\nKetik `/analisa [KODE]` untuk melihat potensi saham.\nContoh: `/analisa BBRI.JK`")
	})

	// /analisa [SYMBOL]
	b.Handle("/analisa", func(c tele.Context) error {
		var symbol string

		// Handle command argument: /analisa BBRI
		args := c.Args()
		if len(args) > 0 {
			symbol = args[0]
		} else {
			// Handle payload (e.g. from button callback)
			symbol = c.Message().Payload
		}

		if symbol == "" {
			return c.Send("Mohon masukkan kode saham.\nContoh: `/analisa BBRI.JK`")
		}

		symbol = strings.ToUpper(symbol)

		// Notify user processing...
		msg, err := b.Send(c.Chat(), fmt.Sprintf("⏳ Sedang menganalisa %s...", symbol))
		// Ignore error if send fails (rare)

		// Call Python API
		result, err := fetchAnalysis(analyticURL, symbol)
		if err != nil {
			if msg != nil {
				b.Edit(msg, fmt.Sprintf("❌ Gagal menganalisa %s.\nError: %v", symbol, err))
			}
			return err
		}

		// Format Response
		signalIcon := "⚪"
		if strings.Contains(result.Technical.Signal, "BUY") {
			signalIcon = "🟢"
		} else if strings.Contains(result.Technical.Signal, "SELL") {
			signalIcon = "🔴"
		}

		reply := fmt.Sprintf(
			"📊 **Analisa Saham: %s**\n\n"+
				"💰 **Harga:** Rp %.0f\n"+
				"📈 **RSI:** %.2f\n"+
				"🌊 **MACD:** %.2f (Signal: %.2f)\n\n"+
				"%s **Sinyal AI:** %s\n"+
				"📝 **Ringkasan:**\n%s",
			result.Symbol,
			result.CurrentPrice,
			result.Technical.RSI,
			result.Technical.MACD,
			result.Technical.MACDSignal,
			signalIcon,
			result.Technical.Signal,
			result.Summary,
		)

		// Update loading message with result
		if msg != nil {
			_, err := b.Edit(msg, reply, tele.ModeMarkdown)
			return err
		}
		return c.Send(reply, tele.ModeMarkdown)
	})

	// 4. Start Bot
	b.Start()
}

// Function to call Python Analytic Engine
func fetchAnalysis(baseURL, symbol string) (*AnalysisResult, error) {
	url := fmt.Sprintf("%s/analyze/%s", baseURL, symbol)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("Gagal menghubungi server Python: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API Error: %s", resp.Status)
	}

	var result AnalysisResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("Gagal baca JSON: %v", err)
	}

	return &result, nil
}
