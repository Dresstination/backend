package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/generative-ai-go/genai"

	// "cloud.google.com/go/storage"
	"github.com/kshitij-404/dresstination-backend/models"
	"github.com/kshitij-404/dresstination-backend/modules"
	"google.golang.org/api/option"
)

type OutfitRequest struct {
	Requirements string `json:"requirements" binding:"required"`
}

func GenerateOutfitsObject(requirements string) (*models.Outfit, error) {

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

	prompt := "You will be given a requirements. You are supposed to generate a title for the occassion/need and then provide an array of strictly 5 different outfits. Each element in the array will have a title, a description, a detailed image prompt that can be used to feed to an AI image generation engine to generate the image of the outift, a search query that can be fed into a shopping website like Amazon.\n\nRequirements: " + requirements

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, err
	}

	// fmt.Println("RESP", resp.Candidates[0].Content.Parts[0])

	marshalResponse, _ := json.Marshal(resp.Candidates[0].Content.Parts[0])

	// Debugging: Print the marshalled response
	// fmt.Println("MRS", string(marshalResponse))

	stringMarshalResponse := string(marshalResponse)

	stringMarshalResponse = strings.ReplaceAll(stringMarshalResponse, "\\\"", "\"")

	// fmt.Println("alpha", stringMarshalResponse)

	stringMarshalResponse = strings.Trim(stringMarshalResponse, "\"")
	// fmt.Println("correctedString", stringMarshalResponse)

	var outfitResponse models.Outfit

	err = json.Unmarshal([]byte(stringMarshalResponse), &outfitResponse)

	if err != nil {
		return nil, fmt.Errorf("error unmarshalling response: %v", err)
	}

	outfitResponse.Timestamp = time.Now().Unix()

	return &outfitResponse, nil
}

func GenerateImageLinks(output *models.Outfit, fs *modules.FS) error {
	apiURL := "https://api.openai.com/v1/images/generations"
	apiKey := os.Getenv("OPENAI_API")
	var wg sync.WaitGroup
	errChan := make(chan error, len(output.OutfitElements))

	bucketName := "dresstination-a2b2f"

	for i, element := range output.OutfitElements {
		wg.Add(1)
		go func(i int, element models.OutfitElement) {
			defer wg.Done()

			requestBody, err := json.Marshal(map[string]interface{}{
				"model":  "dall-e-3",
				"prompt": element.ImagePrompt,
				"n":      1,
				"size":   "1024x1024",
			})
			if err != nil {
				errChan <- fmt.Errorf("error marshalling request body: %v", err)
				return
			}

			req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(requestBody))
			if err != nil {
				errChan <- fmt.Errorf("error creating request: %v", err)
				return
			}

			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+apiKey)

			// Log the request details
			log.Printf("Request URL: %s", apiURL)
			log.Printf("Request Headers: %v", req.Header)
			log.Printf("Request Body: %s", requestBody)

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				errChan <- fmt.Errorf("error making API request: %v", err)
				return
			}
			defer resp.Body.Close()

			body, _ := io.ReadAll(resp.Body)
			log.Printf("Response Status: %s", resp.Status)
			log.Printf("Response Body: %s", string(body))

			if resp.StatusCode != http.StatusOK {
				errChan <- fmt.Errorf("API request failed with status: %v", resp.Status)
				return
			}

			var response map[string]interface{}
			log.Println("Trying to unmarshal")
			if err := json.Unmarshal(body, &response); err != nil {
				errChan <- fmt.Errorf("error decoding response: %v", err)
				return
			}

			log.Println("Unmarshalled Response", response)
			data, ok := response["data"].([]interface{})
			if !ok || len(data) == 0 {
				errChan <- fmt.Errorf("invalid response format")
				return
			}

			imageData, ok := data[0].(map[string]interface{})
			if !ok {
				errChan <- fmt.Errorf("invalid image data format")
				return
			}

			imageURL, ok := imageData["url"].(string)
			if !ok {
				errChan <- fmt.Errorf("image URL not found in response")
				return
			}
			log.Println("Unmarshalled Image URL", imageURL)

			// Download the image
			resp, err = http.Get(imageURL)
			if err != nil {
				errChan <- fmt.Errorf("error downloading image: %v", err)
				return
			}
			defer resp.Body.Close()

			imageBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				errChan <- fmt.Errorf("error reading image data: %v", err)
				return
			}

			// Upload the image to Firebase Storage
			firebaseFilePath := fmt.Sprintf("outfits/image_%d.png", i)
			if err := fs.Upload(imageBytes, bucketName, firebaseFilePath); err != nil {
				errChan <- fmt.Errorf("error uploading image to Firebase: %v", err)
				return
			}

			// Get the public URL of the uploaded image
			imageLink := fmt.Sprintf("https://storage.googleapis.com/%s/%s", bucketName, firebaseFilePath)
			output.OutfitElements[i].ImageLink = imageLink
		}(i, element)
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil
}
func GenerateProducts(outfitItem *models.Outfit) ([]*models.Product, error) {
	var wg sync.WaitGroup
	productChan := make(chan *models.Product, len(outfitItem.OutfitElements))
	errChan := make(chan error, len(outfitItem.OutfitElements))

	for _, element := range outfitItem.OutfitElements {
		wg.Add(1)
		go func(element models.OutfitElement) {
			defer wg.Done()

			searchQuery := element.SearchQuery
			results, err := SearchProductsOnInternet(searchQuery)
			if err != nil {
				errChan <- fmt.Errorf("error searching products: %v", err)
				return
			}

			productChan <- results
		}(element)
	}

	wg.Wait()
	close(productChan)
	close(errChan)

	var products []*models.Product
	for product := range productChan {
		products = append(products, product)
	}

	for err := range errChan {
		if err != nil {
			return nil, err
		}
	}

	return products, nil
}

func SearchProductsOnInternet(searchQuery string) (*models.Product, error) {
	searchQuery = searchQuery + " Amazon, Flipkart, Myntra"
	resultsV1, err := FindProducts(searchQuery)
	if err != nil {
		return nil, err
	}

	resultsV2, err := FindProductsV2(searchQuery)
	if err != nil {
		return nil, err
	}

	combinedResults := append(resultsV1, resultsV2...)
	sort.Slice(combinedResults, func(i, j int) bool {
		return (combinedResults[j].Platform != "") && !(combinedResults[i].Platform != "")
	})

	return combinedResults, nil
}

func FindProducts(topic string) ([]*models.Product, error) {
	var results []*models.Product

	params := url.Values{}
	params.Set("q", topic)
	params.Set("key", os.Getenv("GOOGLE_API_KEY"))
	params.Set("cx", os.Getenv("GOOGLE_ACCOUNT_CX"))

	resp, err := http.Get("https://www.googleapis.com/customsearch/v1?" + params.Encode())
	if err != nil {
		return nil, fmt.Errorf("error making Google Custom Search API request: %v", err)
	}
	defer resp.Body.Close()

	var response struct {
		Items []struct {
			Kind    string `json:"kind"`
			Pagemap struct {
				CseImage []struct {
					Src string `json:"src"`
				} `json:"cse_image"`
			} `json:"pagemap"`
			Title string `json:"title"`
			Link  string `json:"link"`
		} `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("error decoding Google Custom Search API response: %v", err)
	}

	for _, item := range response.Items {
		imageURL := item.Pagemap.CseImage[0].Src
		title := item.Title
		link := item.Link

		if imageURL == "" || title == "" || link == "" {
			continue
		}

		price, currency := ExtractPriceAndCurrency(item)

		var platform string
		for key, value := range predefinedPlatforms {
			for _, v := range value {
				if strings.Contains(link, v) {
					platform = key
					break
				}
			}
		}

		product := &models.Product{
			ImageURL: imageURL,
			Title:    title,
			Link:     link,
			Price:    price,
			Platform: platform,
		}

		results = append(results, product)
	}

	return results, nil
}

func FindProductsV2(topic string) ([]*models.Product, error) {
	params := map[string]string{
		"engine":  "google_shopping",
		"q":       topic,
		"api_key": os.Getenv("SERAPAPI_KEY"),
		"num":     "10",
	}

	res, err := getJson(params)
	if err != nil {
		return nil, fmt.Errorf("error making SERP API request: %v", err)
	}

	var results []*models.Product
	for _, item := range res.ShoppingResults {
		var platform string
		for key, value := range predefinedPlatforms {
			for _, v := range value {
				if strings.Contains(item.Source, v) {
					platform = key
					break
				}
			}
		}

		product := &models.Product{
			ImageURL: item.Thumbnail,
			Title:    item.Title,
			Link:     item.Link,
			Price:    item.Price,
			Platform: platform,
		}

		results = append(results, product)
	}

	return results, nil
}

func ExtractPriceAndCurrency(item struct{}) (string, string) {
	if item.Pagemap.Metatags[0]["product:price:amount"] != nil {
		return item.Pagemap.Metatags[0]["product:price:amount"].(string), item.Pagemap.Metatags[0]["product:price:currency"].(string)
	}

	if item.Pagemap.Offer[0].PriceCurrency != nil {
		return item.Pagemap.Offer[0].Price.(string), item.Pagemap.Offer[0].PriceCurrency.(string)
	}

	if item.Pagemap.Product[0].Price != nil {
		return item.Pagemap.Product[0].Price.(string), item.Pagemap.Product[0].PriceCurrency.(string)
	}

	return "", ""
}

func CreateOutfit(c *gin.Context, firebaseClient *modules.FirebaseClient, fs *modules.FS) {
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

	if err := GenerateImageLinks(output, fs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Println("Output", output)

	// Convert the output to a map[string]interface{}
	outputMap := make(map[string]interface{})
	outputBytes, err := json.Marshal(output)
	log.Println("Marshalled Output", string(outputBytes))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Marshalling error for firestore": err.Error()})
		return
	}
	if err := json.Unmarshal(outputBytes, &outputMap); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Unmarshalling error for firestore": err.Error})
		return
	}

	log.Println("Map Output", outputMap)

	// Generate and populate Prodcuts

	// Upsert the document in Firestore
	collection := "outfits" // Replace with your Firestore collection name
	if err := firebaseClient.InsertDocument(collection, outputMap); err != nil {
		log.Println("Insertion error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"Insertion error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"output": output})
}
