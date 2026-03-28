package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/redjackfred/voicely/backend/internal/database"
	"github.com/redjackfred/voicely/backend/internal/models"
	"github.com/redjackfred/voicely/backend/internal/places"
	"github.com/redjackfred/voicely/backend/internal/voiceai"
)

type CallRequest struct {
	PlaceID string `json:"place_id" binding:"required"` // Google Place ID
	Item    string `json:"item" binding:"required"`     // Expected order item, e.g., "fried chicken"
	Action  string `json:"action" binding:"required"`   // Action type, e.g., "order", "inquiry", "complaint"
	UserID  string `json:"user_id" binding:"required"`  // Who is making the call, used to fetch user preferences and history
}

func main() {
	dsn := "host=localhost user=postgres password=jackfred dbname=voicely port=5432 sslmode=disable TimeZone=Asia/Taipei"
	database.ConnectDB(dsn)

	router := gin.Default()

	v1 := router.Group("/api/v1")
	{
		v1.POST("/call", handleCallRequest)
	}

	log.Println("🚀 Voicely API server is running at http://localhost:8080")
	router.Run(":8080")
}

func handleCallRequest(c *gin.Context) {
	var req CallRequest

	// Binding the incoming JSON to the CallRequest struct, if the format is invalid, return 400 automatically (external knowledge)
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format: " + err.Error()})
		return
	}

	// ==========================================
	// Pahse 1: Fallback Mechanism - Smart Phone Fetching (Location API)
	// ==========================================
	// Project Specification: Call GCP Place Details API to get current_opening_hours, if not open, block the request to save API costs.
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on environment variables for configuration")
	}
	apiKey := os.Getenv("GCP_PLACES_API_KEY")
	if apiKey == "" {
		log.Println("GCP_PLACES_API_KEY environment variable is not set")
	}
	storeDetails, err := places.FetchStoreDetails(req.PlaceID, apiKey)
	if err != nil {
		log.Printf("Failed to fetch store details for Place ID %s: %v", req.PlaceID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot validate store information. Please make sure PlaceID is correct."})
		return
	}
	if storeDetails.CurrentOpeningHours == nil || !storeDetails.CurrentOpeningHours.OpenNow {
		c.JSON(http.StatusConflict, gin.H{
			"status":  "failed",
			"message": "Store ( " + storeDetails.Name + " ) is not opened now, so the order is cancelled",
		})
		return
	}

	// ==========================================
	// Phase 2: Context Gathering
	// ==========================================
	// Project Specification: Extract "User Preferences" and "Store Historical Intelligence" from the database.
	var user models.User
	var store models.Store

	// Use GORM's First method to perform a basic read operation based on ID
	if err := database.DB.First(&user, "id = ?", req.UserID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Could not find user with the provided ID: " + err.Error()})
		return
	}

	// Attempt to find the store, if not found, you can decide whether to create a new store record on the fly based on your needs
	if err := database.DB.First(&store, "id = ?", req.PlaceID).Error; err != nil {
		store = models.Store{
			ID:          req.PlaceID,
			Name:        storeDetails.Name,
			PhoneNumber: storeDetails.FormattedPhoneNumber,
		}
		database.DB.Create(&store)
		log.Printf("Store with Place ID %s not found in database, created new record with name: %s", req.PlaceID, storeDetails.Name)
	}

	// ==========================================
	// Phase 3: Call Engine Activation (Sending to Vapi.ai / Bland AI)
	// ==========================================
	// Here, you would typically assemble the prompt for the LLM by combining the user preferences and store knowledge base,
	// and then send it to your chosen AI agent (e.g., Vapi.ai or Bland AI) to generate the phone call script and execute the call.
	vapiAPIKey := os.Getenv("VAPI_API_KEY")
	if vapiAPIKey == "" {
		log.Println("VAPI_API_KEY environment variable is not set")
	}
	vapiPhoneID := os.Getenv("VAPI_PHONE_NUMBER_ID")
	if vapiPhoneID == "" {
		log.Println("VAPI_PHONE_NUMBER_ID environment variable is not set")
	}

	myTestPhoneNumber := os.Getenv("MY_TEST_PHONE_NUMBER")
	log.Printf("Testing call initiation to %s for store: %s at phone number: %s", req.Item, store.Name, myTestPhoneNumber)

	// 呼叫我們剛剛建立的 voiceai package 發起通話
	// 將 JSONB 型態的 Preferences 與 KnowledgeBase 轉為字串注入 Prompt
	callID, err := voiceai.InitiateCall(
		myTestPhoneNumber,
		store.Name,
		req.Item,
		string(user.Preferences),
		string(store.KnowledgeBase),
		vapiAPIKey,
		vapiPhoneID,
	)
	if err != nil {
		log.Printf("Failed to initiate the call: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot initiate the call, please try again later"})
		return
	}

	// ==========================================
	// 階段四：建立通話紀錄 (Database Logging)
	// ==========================================
	// 根據專案規格，將本次任務寫入 Calls Table [3]
	newCall := models.Call{
		UserID:  user.ID,
		StoreID: store.ID,
		Status:  "calling", // 稍後 Webhook 會更新為 success/failed/retrying
	}
	database.DB.Create(&newCall)

	// 回傳最終成功訊息給前端
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "AI agent has received the request, and it's calling: " + store.Name + " at " + store.PhoneNumber,
		"data": gin.H{
			"place_id":    req.PlaceID,
			"store_phone": store.PhoneNumber,
			"item":        req.Item,
			"user_name":   user.Name,
			"preferences": user.Preferences,
			"call_id":     newCall.ID,
			"vapi_id":     callID,
		},
	})
}

// Simulate GCP Places API check for store opening hours.
func checkStoreOpeningHours(placeID string) bool {
	// Temporarily assume all stores are open.
	return true
}
