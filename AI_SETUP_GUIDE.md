# AI Features Setup Guide

This workflow now includes powerful AI capabilities powered by Perplexity AI. Follow this guide to enable and configure these features.

## Prerequisites

1. **Perplexity AI Account**: Sign up at [Perplexity AI](https://www.perplexity.ai/)
2. **API Access**: Get API access from [Perplexity AI Developers](https://docs.perplexity.ai/docs/getting-started)
3. **API Key**: Obtain your API key from the Perplexity developer console

## Configuration Steps

### 1. Configure API Key

1. Open Alfred Preferences
2. Navigate to Workflows → Search Raindrop.io
3. Click the "Configure Workflow" button (gear icon in top-right)
4. Add your Perplexity API key to the `perplexity_api_key` variable

### 2. Enable AI Features (Optional)

Add these workflow variables to enable specific AI enhancements:

- `ai_tag_suggestions` = `true` - Enable AI-powered tag suggestions when adding bookmarks
- `ai_summaries` = `true` - Enable AI-generated webpage summaries when adding bookmarks

## Features Overview

### 🔍 AI-Powered Search

**Command**: Type `rai` + space + your query

**What it does**:
- Understands natural language queries
- Analyzes your entire bookmark collection for context
- Returns intelligently ranked results with AI explanations

**Examples**:
- "Find articles about machine learning"
- "Show me productivity tools for developers" 
- "What do I have saved about React development?"

### 🏷️ Smart Tag Suggestions

**When**: Automatically appears during bookmark adding process (step 3)

**What it does**:
- Analyzes webpage title, content, and URL
- Suggests relevant tags based on content understanding
- Appears as "🤖 AI Suggestions" in the tag selection interface

### 📄 Content Summaries

**When**: Appears during bookmark title setting (step 2)

**What it does**:
- Generates concise webpage summaries
- Helps you understand content before saving
- Appears as "🤖 AI Summary" below title options

## Usage Tips

1. **API Rate Limits**: Be mindful of Perplexity API usage limits
2. **Internet Required**: AI features require internet connectivity
3. **Fallback**: Regular search still works if AI features fail
4. **Performance**: AI search may be slower than regular search due to analysis

## Troubleshooting

### "Perplexity AI Search - API Key Required"
- Verify your API key is correctly set in workflow configuration
- Check that the key has the correct permissions

### "AI Search Failed"
- Check internet connectivity
- Verify API key is valid and not expired
- Check Perplexity API status

### No AI Suggestions Appearing
- Ensure `ai_tag_suggestions` or `ai_summaries` are set to `true`
- Verify API key is configured
- Check that you're in the correct step of the bookmark adding process

## Privacy & Data

- Your bookmarks are sent to Perplexity AI for analysis during AI search
- Webpage URLs are sent for summarization and tag suggestion
- No data is permanently stored by Perplexity beyond the request
- Consider your privacy preferences when enabling these features

## Cost Considerations

Perplexity AI has usage-based pricing. Each AI search, tag suggestion, or summary request counts toward your API usage. Monitor your usage through the Perplexity developer console.

## Need Help?

If you encounter issues with the AI features:

1. Check this guide for troubleshooting steps
2. Verify your Perplexity AI account and API key status
3. Test with a simple query first
4. Disable AI features if needed by removing the workflow variables

The regular (non-AI) search and bookmark functionality will continue to work regardless of AI configuration.