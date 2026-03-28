package places

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type PlaceDetailsResponse struct {
	Result PlaceResult `json:"result"`
	Status string      `json:"status"`
}

type PlaceResult struct {
	Name                 string               `json:"name"`
	FormattedPhoneNumber string               `json:"formatted_phone_number"`
	CurrentOpeningHours  *CurrentOpeningHours `json:"current_opening_hours"`
}

type CurrentOpeningHours struct {
	OpenNow bool `json:"open_now"`
}

func FetchStoreDetails(placeID string, apiKey string) (*PlaceResult, error) {
	// 指定只需要回傳 name, formatted_phone_number, current_opening_hours 欄位以節省網路頻寬與 API 費用
	url := fmt.Sprintf("https://maps.googleapis.com/maps/api/place/details/json?place_id=%s&fields=name,formatted_phone_number,current_opening_hours&key=%s&language=zh-TW", placeID, apiKey)

	// 設定 10 秒 Timeout，避免第三方 API 沒回應導致我們的 Goroutine 卡死
	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("HTTP 請求失敗: %w", err)
	}
	defer resp.Body.Close()

	var apiResp PlaceDetailsResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("JSON 解析失敗: %w", err)
	}

	if apiResp.Status != "OK" {
		return nil, fmt.Errorf("GCP API 回傳錯誤狀態: %s", apiResp.Status)
	}

	return &apiResp.Result, nil
}
