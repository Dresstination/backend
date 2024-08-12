package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/google/generative-ai-go/genai"
	"github.com/joho/godotenv"
	"github.com/kshitij-404/dresstination-backend/models"
	"google.golang.org/api/option"
)

type OutfitRequest struct {
	Requirements string `json:"requirements" binding:"required"`
}

type Content struct {
	Parts []string `json:"Parts"`
	Role  string   `json:"Role"`
}

type Candidates struct {
	Content *Content `json:"Content"`
}

type ContentResponse struct {
	Candidates []Candidates `json:"Candidates"`
}

func GenerateOutfitsObject(requirements string) (*models.Outfit, error) {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env key")
	}

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Fatal("Environment variable GEMINI_API_KEY not set")
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Fatalf("Error creating client: %v\n", err)
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-1.5-pro")
	model.SetTemperature(1)
	model.SetTopK(64)
	model.SetTopP(0.95)
	model.SetMaxOutputTokens(8192)
	model.ResponseMIMEType = "application/json"
	model.ResponseSchema = &genai.Schema{
		Type:     genai.TypeObject,
		Enum:     []string{},
		Required: []string{"title", "outfit_elements"},
		Properties: map[string]*genai.Schema{
			"title": &genai.Schema{
				Type: genai.TypeString,
			},
			"outfit_elements": &genai.Schema{
				Type: genai.TypeArray,
				Items: &genai.Schema{
					Type:     genai.TypeObject,
					Enum:     []string{},
					Required: []string{"title", "description", "search_query", "image_prompt"},
					Properties: map[string]*genai.Schema{
						"title": &genai.Schema{
							Type: genai.TypeString,
						},
						"description": &genai.Schema{
							Type: genai.TypeString,
						},
						"search_query": &genai.Schema{
							Type: genai.TypeString,
						},
						"image_prompt": &genai.Schema{
							Type: genai.TypeString,
						},
					},
				},
			},
		},
	}

	model.SafetySettings = []*genai.SafetySetting{
		{
			Category:  genai.HarmCategoryHarassment,
			Threshold: genai.HarmBlockNone,
		},
		{
			Category:  genai.HarmCategoryHateSpeech,
			Threshold: genai.HarmBlockNone,
		},
		{
			Category:  genai.HarmCategorySexuallyExplicit,
			Threshold: genai.HarmBlockNone,
		},
		{
			Category:  genai.HarmCategoryDangerousContent,
			Threshold: genai.HarmBlockNone,
		},
	}

	prompt := "You will be given a requirements. You are supposed to generate a title for the occassion/need and then provide an array of strictly 4 different outfits. Each element in the array will have a title, a description, a detailed image prompt that can be used to feed to an AI image generation engine to generate the image of the outift, a search query that can be fed into a shopping website like Amazon.\n\nRequirements: " + requirements

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, err
	}

	fmt.Println("RESP", resp.Candidates[0].Content.Parts[0])
	type OutfitElement struct {
		Description string `json:"description"`
		ImagePrompt string `json:"image_prompt"`
		SearchQuery string `json:"search_query"`
		Title       string `json:"title"`
	}

	type Outfit struct {
		OutfitElements []OutfitElement `json:"outfit_elements"`
		Title          string          `json:"title"`
	}

	var formattedData Outfit
	marshalResponse, _ := json.Marshal(resp.Candidates[0].Content.Parts[0])
	err = json.Unmarshal(marshalResponse, &formattedData)
	fmt.Println("MYDATA", formattedData, err)

	// Debugging: Print the marshalled response
	fmt.Println("MRS", marshalResponse)

	// Debugging: Print the raw response
	// fmt.Printf("Raw response: %+v\n", resp)

	var generateResponse ContentResponse
	// marshalResponse, _ := json.Marshal(resp)
	// if err := json.Unmarshal(marshalResponse, &generateResponse); err != nil {
	//     return nil, fmt.Errorf("error unmarshalling response: %v", err)
	// }

	var outfitResponse models.Outfit

	for _, cad := range generateResponse.Candidates {
		if cad.Content != nil {
			for _, part := range cad.Content.Parts {
				fmt.Printf("%T\n", part)
			}
		}
	}

	//         contentBytes, err := json.Marshal(cad.Content)
	//         if err != nil {
	//             return nil, fmt.Errorf("error marshalling content: %v", err)
	//         }
	//         // Debugging: Print the marshalled content
	//         fmt.Printf("Marshalled content: %s\n", string(contentBytes))

	//         // Unmarshal the content into the outfitResponse
	//         err = json.Unmarshal(contentBytes, &outfitResponse)
	//         if err != nil {
	//             return nil, fmt.Errorf("error unmarshalling response: %v", err)
	//         }
	//         break
	//     }
	// }

	// // Debugging: Print the final outfit response
	// fmt.Printf("Final outfit response: %+v\n", outfitResponse)

	return &outfitResponse, nil
}

func CreateOutfits(c *gin.Context) {
	var req OutfitRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	output, err := GenerateOutfitsObject(req.Requirements)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"output": output})
}
