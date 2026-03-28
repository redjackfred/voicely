package voiceai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// CallRequest payload 發送給 Vapi.ai 的結構
type VapiCallRequest struct {
	PhoneNumber string `json:"phoneNumber"`
	Prompt      string `json:"prompt"`
	// 可根據 Vapi 官方文件加入其他設定，例如使用的 Voice ID
}

// InitiateCall 組合 Prompt 並呼叫第三方 Voice AI API
func InitiateCall(storePhone, storeName, item, userPrefs, storeKnowledge, apiKey string) (string, error) {
	// 1. 動態組裝 System Prompt
	// 根據 MVP 規格，在此注入使用者的動態替代方案與店家情報，並加入來電信任宣告
	systemPrompt := fmt.Sprintf(`
你現在是一個名為「替聲」的 AI 語音助理，你的任務是打電話向「%s」預約餐點。

【重要通話規則】
通話接通後的前 3 秒，你必須固定宣告：「您好，這是由 AI 助理代為撥打的預約電話...」。

【任務目標】
請幫我訂購：%s。

【使用者授權與偏好 (User Preferences)】
以下是使用者的飲食禁忌與替代方案授權，若店家表示原餐點沒有，請直接依據以下授權範圍自主決策，不要中斷通話回問我：
%s

【店家歷史情報 (Store Knowledge)】
以下是過去通話累積的店家情報，請作為參考（例如是否已知某些品項常售完）：
%s

請用自然、有禮貌的台灣口音中文與店家溝通，完成訂餐後禮貌道別並掛斷電話。
`, storeName, item, userPrefs, storeKnowledge)

	// 2. 準備發送給 Vapi 的 Payload
	reqBody := VapiCallRequest{
		PhoneNumber: storePhone,
		Prompt:      systemPrompt,
	}
	jsonData, _ := json.Marshal(reqBody)

	// 3. 發送 HTTP POST 請求至 Vapi /call 端點
	// 注意：此為示意 URL，實際請依據 Vapi.ai 或 Bland AI 的 API 文件填寫端點
	url := "https://api.vapi.ai/call"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("建立請求失敗: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("呼叫 Voice AI API 失敗: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("Voice AI API 回傳錯誤狀態碼: %d", resp.StatusCode)
	}

	// 這裡可以解析 Vapi 回傳的 Call ID 以供後續追蹤
	return "vapi_call_id_12345", nil
}
