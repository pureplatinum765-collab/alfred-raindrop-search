/*
	Functions for AI-powered search and bookmark enhancement using Perplexity AI

	By Andreas Westerlind, 2021-2025
*/

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	aw "github.com/deanishe/awgo"
)

// Perplexity API structures
type PerplexityRequest struct {
	Model       string                   `json:"model"`
	Messages    []PerplexityMessage      `json:"messages"`
	MaxTokens   int                      `json:"max_tokens,omitempty"`
	Temperature float64                  `json:"temperature,omitempty"`
	Stream      bool                     `json:"stream"`
}

type PerplexityMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type PerplexityResponse struct {
	ID      string                 `json:"id"`
	Object  string                 `json:"object"`
	Created int64                  `json:"created"`
	Model   string                 `json:"model"`
	Choices []PerplexityChoice     `json:"choices"`
	Usage   PerplexityUsage        `json:"usage"`
	Error   *PerplexityError       `json:"error,omitempty"`
}

type PerplexityChoice struct {
	Index        int               `json:"index"`
	FinishReason string            `json:"finish_reason"`
	Message      PerplexityMessage `json:"message"`
	Delta        PerplexityMessage `json:"delta,omitempty"`
}

type PerplexityUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type PerplexityError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code,omitempty"`
}

// AI search function
func ai_search(query string, token RaindropToken, descr_in_list bool) {
	// Check if Perplexity API key is configured
	perplexity_api_key := wf.Config.Get("perplexity_api_key", "")
	if perplexity_api_key == "" {
		wf.NewItem("Perplexity AI Search - API Key Required").
			Subtitle("Configure your Perplexity API key in workflow settings").
			Valid(false).
			Icon(&aw.Icon{Value: "icon.png", Type: ""})
		return
	}

	if query == "" {
		// Show help when no query is provided
		wf.NewItem("AI-Powered Bookmark Search").
			Subtitle("Type your question or search query to find relevant bookmarks using AI").
			Valid(false).
			Icon(&aw.Icon{Value: "icon.png", Type: ""})
		
		wf.NewItem("Example: \"Find articles about machine learning\"").
			Subtitle("AI will search your bookmarks and provide intelligent results").
			Valid(false).
			Icon(&aw.Icon{Value: "icon.png", Type: ""})
		return
	}

	// Get all bookmarks to provide context to AI
	bookmarks := get_all_bookmarks(token, "check")
	if len(bookmarks) == 0 {
		wf.NewItem("No bookmarks found").
			Subtitle("Your Raindrop.io account appears to be empty").
			Valid(false).
			Icon(&aw.Icon{Value: "icon.png", Type: ""})
		return
	}

	// Prepare context for AI from bookmarks
	bookmark_context := prepare_bookmark_context(bookmarks, 50) // Limit to 50 bookmarks for context

	// Query Perplexity AI
	ai_results, err := query_perplexity_ai(query, bookmark_context, perplexity_api_key)
	if err != nil {
		wf.NewItem("AI Search Failed").
			Subtitle(fmt.Sprintf("Error: %s", err.Error())).
			Valid(false).
			Icon(&aw.Icon{Value: "icon.png", Type: ""})
		return
	}

	// Parse AI response and match with actual bookmarks
	matched_bookmarks := match_ai_response_with_bookmarks(ai_results, bookmarks, query)
	
	if len(matched_bookmarks) == 0 {
		wf.NewItem("No AI matches found").
			Subtitle("Try rephrasing your query or use regular search").
			Valid(false).
			Icon(&aw.Icon{Value: "icon.png", Type: ""})
		return
	}

	// Get collection names for display
	raindrop_collections := get_collections(token, false, "check")
	raindrop_collections_sublevel := get_collections(token, true, "check")
	var current_object []string
	collection_names := collection_paths(raindrop_collections, raindrop_collections_sublevel, make(map[int]string), 0, current_object, -1)

	// Render the AI-enhanced results
	render_results(matched_bookmarks, "only", collection_names, descr_in_list)

	// Add AI explanation at the end
	if ai_results != "" {
		explanation := extract_ai_explanation(ai_results)
		if explanation != "" {
			wf.NewItem("🤖 AI Insight").
				Subtitle(explanation).
				Valid(false).
				Icon(&aw.Icon{Value: "icon.png", Type: ""})
		}
	}
}

// Prepare bookmark context for AI query
func prepare_bookmark_context(bookmarks []interface{}, limit int) string {
	var context_items []string
	count := 0
	
	for _, bookmark_interface := range bookmarks {
		if count >= limit {
			break
		}
		
		bookmark := bookmark_interface.(map[string]interface{})
		title := ""
		excerpt := ""
		link := ""
		tags := ""
		
		if bookmark["title"] != nil {
			title = bookmark["title"].(string)
		}
		if bookmark["excerpt"] != nil && bookmark["excerpt"].(string) != "" {
			excerpt = bookmark["excerpt"].(string)
		}
		if bookmark["link"] != nil {
			link = bookmark["link"].(string)
		}
		if bookmark["tags"] != nil {
			tag_array := bookmark["tags"].([]interface{})
			tag_strings := make([]string, len(tag_array))
			for i, tag := range tag_array {
				tag_strings[i] = tag.(string)
			}
			tags = strings.Join(tag_strings, ", ")
		}
		
		// Create a compact context entry
		context_entry := fmt.Sprintf("Title: %s", title)
		if excerpt != "" {
			context_entry += fmt.Sprintf(" | Description: %s", excerpt[:min(150, len(excerpt))])
		}
		if tags != "" {
			context_entry += fmt.Sprintf(" | Tags: %s", tags)
		}
		context_entry += fmt.Sprintf(" | URL: %s", link)
		
		context_items = append(context_items, context_entry)
		count++
	}
	
	return strings.Join(context_items, "\n")
}

// Query Perplexity AI
func query_perplexity_ai(query, context, api_key string) (string, error) {
	system_prompt := `You are an AI assistant helping to search through bookmarks. Given a user's query and their bookmark collection, identify the most relevant bookmarks and explain why they match. Respond with:

1. A list of the most relevant bookmark titles (exactly as they appear in the context)
2. A brief explanation of why these bookmarks are relevant

Keep your response concise and focus on the most relevant matches.`

	user_prompt := fmt.Sprintf("Query: %s\n\nBookmark Collection:\n%s", query, context)

	request := PerplexityRequest{
		Model: "llama-3.1-sonar-small-128k-online",
		Messages: []PerplexityMessage{
			{Role: "system", Content: system_prompt},
			{Role: "user", Content: user_prompt},
		},
		MaxTokens:   500,
		Temperature: 0.2,
		Stream:      false,
	}

	json_data, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %v", err)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("POST", "https://api.perplexity.ai/chat/completions", bytes.NewBuffer(json_data))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+api_key)

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}

	var perplexity_response PerplexityResponse
	if err := json.Unmarshal(body, &perplexity_response); err != nil {
		return "", fmt.Errorf("failed to parse response: %v", err)
	}

	if perplexity_response.Error != nil {
		return "", fmt.Errorf("Perplexity API error: %s", perplexity_response.Error.Message)
	}

	if len(perplexity_response.Choices) == 0 {
		return "", fmt.Errorf("no response from Perplexity AI")
	}

	return perplexity_response.Choices[0].Message.Content, nil
}

// Match AI response with actual bookmarks
func match_ai_response_with_bookmarks(ai_response string, bookmarks []interface{}, original_query string) []interface{} {
	var matched_bookmarks []interface{}
	ai_response_lower := strings.ToLower(ai_response)
	original_query_lower := strings.ToLower(original_query)
	
	// Score each bookmark based on AI response and original query
	type BookmarkScore struct {
		Bookmark interface{}
		Score    int
	}
	
	var scored_bookmarks []BookmarkScore
	
	for _, bookmark_interface := range bookmarks {
		bookmark := bookmark_interface.(map[string]interface{})
		score := 0
		
		title := ""
		excerpt := ""
		tags := ""
		
		if bookmark["title"] != nil {
			title = strings.ToLower(bookmark["title"].(string))
		}
		if bookmark["excerpt"] != nil && bookmark["excerpt"].(string) != "" {
			excerpt = strings.ToLower(bookmark["excerpt"].(string))
		}
		if bookmark["tags"] != nil {
			tag_array := bookmark["tags"].([]interface{})
			tag_strings := make([]string, len(tag_array))
			for i, tag := range tag_array {
				tag_strings[i] = strings.ToLower(tag.(string))
			}
			tags = strings.Join(tag_strings, " ")
		}
		
		// Check if bookmark title is mentioned in AI response
		if title != "" && strings.Contains(ai_response_lower, title) {
			score += 10
		}
		
		// Check for query terms in title, excerpt, and tags
		query_terms := strings.Fields(original_query_lower)
		for _, term := range query_terms {
			if len(term) > 2 { // Ignore very short terms
				if strings.Contains(title, term) {
					score += 3
				}
				if strings.Contains(excerpt, term) {
					score += 2
				}
				if strings.Contains(tags, term) {
					score += 1
				}
			}
		}
		
		// Add to scored list if has any relevance
		if score > 0 {
			scored_bookmarks = append(scored_bookmarks, BookmarkScore{
				Bookmark: bookmark_interface,
				Score:    score,
			})
		}
	}
	
	// Sort by score (highest first)
	for i := 0; i < len(scored_bookmarks); i++ {
		for j := i + 1; j < len(scored_bookmarks); j++ {
			if scored_bookmarks[j].Score > scored_bookmarks[i].Score {
				scored_bookmarks[i], scored_bookmarks[j] = scored_bookmarks[j], scored_bookmarks[i]
			}
		}
	}
	
	// Return top matches
	limit := min(10, len(scored_bookmarks))
	for i := 0; i < limit; i++ {
		matched_bookmarks = append(matched_bookmarks, scored_bookmarks[i].Bookmark)
	}
	
	return matched_bookmarks
}

// Extract AI explanation from response
func extract_ai_explanation(ai_response string) string {
	lines := strings.Split(ai_response, "\n")
	var explanation_lines []string
	in_explanation := false
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		
		// Look for explanation indicators
		if strings.Contains(strings.ToLower(line), "relevant") || 
		   strings.Contains(strings.ToLower(line), "because") ||
		   strings.Contains(strings.ToLower(line), "why") ||
		   strings.Contains(strings.ToLower(line), "these bookmarks") {
			in_explanation = true
		}
		
		if in_explanation && len(line) > 20 && !strings.HasPrefix(line, "Title:") {
			explanation_lines = append(explanation_lines, line)
		}
	}
	
	if len(explanation_lines) > 0 {
		explanation := strings.Join(explanation_lines, " ")
		// Limit explanation length for Alfred display
		if len(explanation) > 200 {
			explanation = explanation[:200] + "..."
		}
		return explanation
	}
	
	return ""
}

// AI-powered bookmark summarization
func ai_summarize_bookmark(bookmark_url, perplexity_api_key string) (string, error) {
	system_prompt := "You are a helpful assistant that creates concise summaries of web pages. Provide a brief 2-3 sentence summary of the main content and key points."
	
	user_prompt := fmt.Sprintf("Please provide a concise summary of this webpage: %s", bookmark_url)
	
	request := PerplexityRequest{
		Model: "llama-3.1-sonar-small-128k-online",
		Messages: []PerplexityMessage{
			{Role: "system", Content: system_prompt},
			{Role: "user", Content: user_prompt},
		},
		MaxTokens:   150,
		Temperature: 0.3,
		Stream:      false,
	}
	
	json_data, err := json.Marshal(request)
	if err != nil {
		return "", err
	}
	
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("POST", "https://api.perplexity.ai/chat/completions", bytes.NewBuffer(json_data))
	if err != nil {
		return "", err
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+perplexity_api_key)
	
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	
	var perplexity_response PerplexityResponse
	if err := json.Unmarshal(body, &perplexity_response); err != nil {
		return "", err
	}
	
	if perplexity_response.Error != nil {
		return "", fmt.Errorf("Perplexity API error: %s", perplexity_response.Error.Message)
	}
	
	if len(perplexity_response.Choices) == 0 {
		return "", fmt.Errorf("no response from Perplexity AI")
	}
	
	return perplexity_response.Choices[0].Message.Content, nil
}

// AI-powered tag suggestions
func ai_suggest_tags(title, excerpt, url, perplexity_api_key string) ([]string, error) {
	system_prompt := "You are a helpful assistant that suggests relevant tags for bookmarks. Based on the title, description, and URL provided, suggest 3-5 concise, relevant tags that would help categorize this bookmark. Return only the tags, separated by commas, without any additional text."
	
	user_prompt := fmt.Sprintf("Title: %s\nDescription: %s\nURL: %s\n\nSuggest relevant tags:", title, excerpt, url)
	
	request := PerplexityRequest{
		Model: "llama-3.1-sonar-small-128k-online",
		Messages: []PerplexityMessage{
			{Role: "system", Content: system_prompt},
			{Role: "user", Content: user_prompt},
		},
		MaxTokens:   100,
		Temperature: 0.3,
		Stream:      false,
	}
	
	json_data, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("POST", "https://api.perplexity.ai/chat/completions", bytes.NewBuffer(json_data))
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+perplexity_api_key)
	
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	var perplexity_response PerplexityResponse
	if err := json.Unmarshal(body, &perplexity_response); err != nil {
		return nil, err
	}
	
	if perplexity_response.Error != nil {
		return nil, fmt.Errorf("Perplexity API error: %s", perplexity_response.Error.Message)
	}
	
	if len(perplexity_response.Choices) == 0 {
		return nil, fmt.Errorf("no response from Perplexity AI")
	}
	
	// Parse the response to extract tags
	response := strings.TrimSpace(perplexity_response.Choices[0].Message.Content)
	tags := strings.Split(response, ",")
	
	// Clean up tags
	var clean_tags []string
	for _, tag := range tags {
		clean_tag := strings.TrimSpace(tag)
		if clean_tag != "" && len(clean_tag) > 1 {
			clean_tags = append(clean_tags, clean_tag)
		}
	}
	
	return clean_tags, nil
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}