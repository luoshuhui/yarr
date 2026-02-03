# AI Summary Feature Design

## Overview

Add AI-powered article summarization to yarr RSS reader with support for both Gemini API and local Ollama deployment.

## User Requirements

1. Support Gemini API key configuration
2. Support Ollama local deployment
3. Add "Summarize" button in the article reading view (top-right toolbar)
4. Display AI-generated summary in a box inserted before the article content
5. Cache summaries to avoid redundant API calls

## Architecture Overview

### Data Layer

**Settings Table Extension:**
```sql
ALTER TABLE settings ADD COLUMN ai_provider TEXT DEFAULT 'disabled';
ALTER TABLE settings ADD COLUMN gemini_api_key TEXT DEFAULT '';
ALTER TABLE settings ADD COLUMN ollama_url TEXT DEFAULT 'http://localhost:11434';
ALTER TABLE settings ADD COLUMN ollama_model TEXT DEFAULT 'qwen2:4b';
```

**Items Table Extension:**
```sql
ALTER TABLE items ADD COLUMN ai_summary TEXT;
ALTER TABLE items ADD COLUMN ai_summary_at INTEGER;
```

- `ai_summary`: Stores the generated summary text
- `ai_summary_at`: Unix timestamp for displaying "Generated X minutes ago"

### Backend Layer

**New Package: `src/summarizer/`**

Core interface design:
```go
// summarizer.go - Core interface
type Summarizer interface {
    Summarize(ctx context.Context, title, content string) (string, error)
}

// Factory method to create summarizer based on provider
func New(provider string, config map[string]string) (Summarizer, error)
```

**Gemini Implementation** (`gemini.go`):
- API endpoint: `https://generativelanguage.googleapis.com/v1beta/models/gemini-pro:generateContent`
- Constructs prompt with title and content
- Handles API responses and errors

**Ollama Implementation** (`ollama.go`):
- API endpoint: `POST {url}/api/generate`
- Uses configured model name
- Handles local service connection failures

**New API Endpoint:**
```
POST /api/items/:id/summarize?regenerate=false
```

Logic flow:
1. Check if AI is enabled (provider != 'disabled')
2. Fetch article content from database
3. If cached summary exists and `regenerate=false`, return cache
4. Otherwise, call summarizer to generate summary
5. Save summary to database (`ai_summary` and `ai_summary_at`)
6. Return JSON: `{summary: "...", generatedAt: timestamp}`

### Frontend Layer

**Settings Page Extension** (in `src/assets/index.html`):

```html
<div class="settings-section">
  <h3>AI Summary Settings</h3>

  <div class="form-group">
    <label>AI Provider</label>
    <select v-model="settings.ai_provider">
      <option value="disabled">Disabled</option>
      <option value="gemini">Gemini API</option>
      <option value="ollama">Ollama (Local)</option>
    </select>
  </div>

  <!-- Gemini config - shown only when gemini selected -->
  <div v-if="settings.ai_provider === 'gemini'" class="form-group">
    <label>Gemini API Key</label>
    <input type="password" v-model="settings.gemini_api_key" />
  </div>

  <!-- Ollama config - shown only when ollama selected -->
  <div v-if="settings.ai_provider === 'ollama'">
    <div class="form-group">
      <label>Ollama URL</label>
      <input type="text" v-model="settings.ollama_url" placeholder="http://localhost:11434" />
    </div>
    <div class="form-group">
      <label>Model Name</label>
      <input type="text" v-model="settings.ollama_model" placeholder="qwen2:4b" />
    </div>
  </div>
</div>
```

**Article Reading View Modifications:**

1. Add Summarize button (top-right toolbar):
```html
<div class="article-toolbar">
  <!-- Existing buttons: mark read, favorite, etc -->
  <button v-if="settings.ai_provider !== 'disabled'"
          @click="summarizeArticle"
          :disabled="summarizing">
    {{ selectedItem.ai_summary ? 'Regenerate Summary' : 'Summarize' }}
  </button>
</div>
```

2. AI Summary display area (inserted before article content):
```html
<div v-if="selectedItem.ai_summary" class="ai-summary-box">
  <div class="ai-summary-header">
    <strong>🤖 AI Summary</strong>
    <span class="summary-time">Generated {{ formatTime(selectedItem.ai_summary_at) }}</span>
  </div>
  <div class="ai-summary-content">{{ selectedItem.ai_summary }}</div>
</div>

<!-- Original article content -->
<div class="article-content" v-html="selectedItem.content"></div>
```

**JavaScript Logic** (in `app.js`):

```javascript
summarizeArticle() {
  this.summarizing = true;
  const regenerate = !!this.selectedItem.ai_summary;

  api.summarizeItem(this.selectedItem.id, regenerate)
    .then(result => {
      this.selectedItem.ai_summary = result.summary;
      this.selectedItem.ai_summary_at = result.generatedAt;
    })
    .catch(err => {
      alert('Failed to generate summary: ' + err.message);
    })
    .finally(() => {
      this.summarizing = false;
    });
}
```

In `api.js`:
```javascript
summarizeItem(id, regenerate) {
  return fetch(`/api/items/${id}/summarize?regenerate=${regenerate}`, {
    method: 'POST'
  }).then(r => r.json());
}
```

## Error Handling

### Backend Error Handling

1. **Configuration Validation**:
   - Empty Gemini API Key → 400 error: "Gemini API key not configured"
   - Ollama URL unreachable → 503 error: "Ollama service unavailable"

2. **API Call Failures**:
   - Gemini API timeout/rate limit → Log and return specific error
   - Ollama model not found → 404 error: "Model not found"

3. **Content Length Handling**:
   - Article exceeds API limit → Auto-truncate to first 10,000 characters
   - Or return error: "Article too long to summarize"

### Frontend User Experience

1. **Loading State**:
   - Button shows "Summarizing..." with loading animation
   - Disable button to prevent duplicate clicks

2. **Error Messages**:
   - Display friendly error messages (not raw API errors)
   - Configuration errors guide user to settings page

## Prompt Design

To ensure high-quality summaries, use structured prompts:

```
Please provide a concise summary of the following article in 3-5 bullet points. Focus on the main ideas and key takeaways.

Title: {title}

Content: {content}

Requirements:
- Use the same language as the article
- Keep each point under 50 words
- Focus on facts, not opinions
```

## Styling

CSS for AI Summary box (add to `app.css`):

```css
.ai-summary-box {
  background: #f8f9fa;
  border-left: 4px solid #007bff;
  padding: 15px;
  margin-bottom: 20px;
  border-radius: 4px;
}

.ai-summary-header {
  display: flex;
  justify-content: space-between;
  margin-bottom: 10px;
  font-size: 14px;
  color: #6c757d;
}

.ai-summary-content {
  line-height: 1.6;
  color: #212529;
}
```

## Implementation Phases

### Phase 1: Database Foundation
1. Add migration in `src/storage/migrations.go`
2. Add AI config field read/write methods in `src/storage/settings.go`
3. Add `AISummary` and `AISummaryAt` fields to Item struct in `src/storage/storage.go`

### Phase 2: Summarizer Implementation
4. Create `src/summarizer/summarizer.go` - interface definition
5. Implement `src/summarizer/gemini.go` - Gemini API client
6. Implement `src/summarizer/ollama.go` - Ollama API client
7. Add unit tests (mock API responses)

### Phase 3: Backend API
8. Add `/api/items/:id/summarize` endpoint in `src/server/routes.go`
9. Implement business logic: check cache, call summarizer, save results
10. Modify `/api/items` endpoint to include `ai_summary` and `ai_summary_at` fields

### Phase 4: Frontend UI
11. Modify settings page HTML, add AI configuration area
12. Handle settings save/load in `app.js`
13. Modify article reading view, add Summarize button and Summary display box
14. Add `summarizeItem` method in `api.js`
15. Add CSS styles

## Testing Plan

### Unit Tests
- Gemini API client (using mock server)
- Ollama API client
- Prompt construction logic

### Integration Tests
- Full flow: configure → click button → generate summary → verify cache
- Error scenarios: invalid API key, Ollama service unavailable
- Regenerate summary functionality

### Manual Testing Checklist
- [ ] Settings page correctly shows/hides configuration fields
- [ ] Gemini API successfully generates summaries
- [ ] Ollama successfully generates summaries
- [ ] Summaries correctly cached to database
- [ ] Button text correctly toggles (Summarize ↔ Regenerate)
- [ ] Error messages are friendly and accurate
- [ ] Styling displays correctly on different screen sizes

## Summary

This design provides:
- ✅ Flexible AI provider selection (Gemini / Ollama)
- ✅ Smart caching to reduce API call costs
- ✅ Clean user interface (settings page + reading view)
- ✅ Robust error handling
- ✅ Modular code structure, easy to extend

## Future Enhancements (Out of Scope)

- Support for additional AI providers (OpenAI, Claude, etc.)
- Customizable prompt templates
- Summary language selection
- Batch summarization for multiple articles
- Summary quality feedback mechanism
