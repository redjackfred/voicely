package voiceai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// ==========================================
// Vapi 官方 API Payload 結構定義
// ==========================================

// VapiCustomer 定義接收來電的對象 (店家)
type VapiCustomer struct {
	Number string `json:"number"`
}

// VapiMessage 定義給 LLM 的訊息格式
type VapiMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// VapiModel 定義 Assistant 內部使用的大型語言模型 (LLM) 設定
type VapiModel struct {
	Provider string        `json:"provider"` // 例如: "openai" 或 "anthropic"
	Model    string        `json:"model"`    // 例如: "gpt-4-turbo"
	Messages []VapiMessage `json:"messages"` // 這裡放置我們的 System Prompt
}

// VapiVoice 定義 Assistant 使用的文字轉語音 (TTS) 聲音模型
type VapiVoice struct {
	Provider string `json:"provider"` // 例如: "11labs", "openai"
	VoiceID  string `json:"voiceId"`  // 該 Provider 平台上的特定 Voice ID
}

// VapiAssistant 結合了模型、聲音與對話設定
type VapiAssistant struct {
	FirstMessage string    `json:"firstMessage"` // 通話接通後 AI 講的第一句話
	Model        VapiModel `json:"model"`
	Voice        VapiVoice `json:"voice"`
}

// VapiCallRequest 完整的發起通話請求 Payload
type VapiCallRequest struct {
	PhoneNumberID string        `json:"phoneNumberId"` // 你在 Vapi 平台上購買或綁定的發話號碼 ID
	Customer      VapiCustomer  `json:"customer"`
	Assistant     VapiAssistant `json:"assistant"`
}

// ==========================================
// 啟動通話邏輯
// ==========================================

// InitiateCall 動態組裝 Assistant 屬性並呼叫 Vapi 發起通話
func InitiateCall(storePhone, storeName, item, userPrefs, storeKnowledge, apiKey, vapiPhoneNumberID string) (string, error) {
	// 1. 動態組裝 System Prompt (不再需要把 FirstMessage 寫進 prompt 裡，交給系統處理)
	systemPrompt := fmt.Sprintf(`
你現在是一個名為「替聲」的 AI 語音助理，你的任務是向「%s」預約餐點。

【任務目標】
請幫我訂購：%s。

【使用者授權與偏好 (User Preferences)】
以下是使用者的飲食禁忌與替代方案授權，若店家表示原餐點沒有，請直接依據以下授權範圍自主決策，不要中斷通話回問我：
%s

【店家歷史情報 (Store Knowledge)】
以下是過去通話累積的店家情報，請作為參考：
%s

請用自然、有禮貌的台灣口音中文與店家溝通，完成訂餐後禮貌道別並掛斷電話。
`, storeName, item, userPrefs, storeKnowledge)

	// 2. 依照 Vapi 官方結構組裝 Payload
	reqBody := VapiCallRequest{
		PhoneNumberID: vapiPhoneNumberID, // 必須提供外撥的虛擬號碼 ID
		Customer: VapiCustomer{
			Number: storePhone, // 目標店家的電話 (測試時請填自己號碼)
		},
		Assistant: VapiAssistant{
			// 利用 Vapi 原生的 FirstMessage 完美實作我們 MVP 的信任宣告防呆機制
			FirstMessage: "您好，這是由 AI 助理代為撥打的電話。",
			Model: VapiModel{
				Provider: "openai",
				Model:    "gpt-4-turbo", // 推薦使用較快的模型以降低延遲
				Messages: []VapiMessage{
					{
						Role:    "system",
						Content: systemPrompt,
					},
				},
			},
			Voice: VapiVoice{
				Provider: "11labs",               // 假設我們使用 ElevenLabs
				VoiceID:  "4aW8bNY2tSD8eaHmuXZ0", // 請替換為適合台灣口音的 Voice ID
			},
		},
	}
	jsonData, _ := json.Marshal(reqBody)

	// 3. 發送 HTTP POST 請求至 Vapi 官方發起通話的端點
	url := "https://api.vapi.ai/call/phone"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("建立請求失敗: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	// 設定 Timeout 避免阻斷 Goroutine
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("呼叫 Vapi API 失敗: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		// 建議在開發期印出錯誤 body 幫助除錯
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		return "", fmt.Errorf("Vapi 回傳錯誤 (HTTP %d): %s", resp.StatusCode, buf.String())
	}

	// 這裡實務上會解析 resp.Body 取得 Vapi 的 Call ID
	return "vapi_call_id_parsed_from_response", nil
}
