package services

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"novelclaw/internal/dtos"
	"novelclaw/internal/gen/sqlc"
	"novelclaw/pkg/graph"
	"novelclaw/pkg/llm"
	"novelclaw/pkg/storage"
)

// GraphService exposes progressive entity graph and temporal relations to Wails v3 frontend.
type GraphService struct {
	store  *storage.Storage
	pg     *graph.ProgressiveGraph
	client llm.LLMClient
}

// NewGraphService creates a new GraphService.
func NewGraphService(store *storage.Storage) *GraphService {
	return &GraphService{
		store: store,
		pg:    graph.NewProgressiveGraph(store),
	}
}

// SetGraphTestClient allows test suites to inject mock LLM clients without triggering Wails binding warnings.
func SetGraphTestClient(s *GraphService, client llm.LLMClient) {
	s.client = client
}

func (s *GraphService) resolveClient(ctx context.Context) (llm.LLMClient, string, error) {
	if s.client != nil {
		return s.client, "default-model", nil
	}
	cfg, err := s.store.GetDefaultLLMConfig(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("LLM API key is not configured. Please open Settings to set up your LLM model: %w", err)
	}
	timeout := time.Duration(cfg.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 120 * time.Second
	}
	var client llm.LLMClient
	if strings.Contains(cfg.ApiURL, "generativelanguage.googleapis.com") {
		client = llm.NewGeminiClient(cfg.Token, timeout)
	} else {
		client = llm.NewOpenAIClient(cfg.ApiURL, cfg.Token, timeout)
	}
	return client, cfg.ModelName, nil
}

// UpsertEntity inserts or updates an entity node in the graph using a DTO request.
func (s *GraphService) UpsertEntity(ctx context.Context, req dtos.UpsertEntityRequest) error {
	return s.store.UpsertEntity(ctx, req.ToDomainEntity())
}

// ListEntities lists all entities for a project as DTOs.
func (s *GraphService) ListEntities(ctx context.Context, projectID string) ([]dtos.EntityDTO, error) {
	entities, err := s.store.ListEntitiesByProject(ctx, projectID)
	if err != nil {
		return nil, err
	}
	dtosList := make([]dtos.EntityDTO, 0, len(entities))
	for _, e := range entities {
		dtosList = append(dtosList, dtos.ToEntityDTO(e))
	}
	return dtosList, nil
}

// ListActiveEntitiesAtChapter lists entities introduced at or before the given chapter as DTOs.
func (s *GraphService) ListActiveEntitiesAtChapter(ctx context.Context, projectID string, chapterIndex int64) ([]dtos.EntityDTO, error) {
	entities, err := s.store.ListActiveEntitiesAtChapter(ctx, projectID, chapterIndex)
	if err != nil {
		return nil, err
	}
	dtosList := make([]dtos.EntityDTO, 0, len(entities))
	for _, e := range entities {
		dtosList = append(dtosList, dtos.ToEntityDTO(e))
	}
	return dtosList, nil
}

// DeleteEntity deletes an entity from the graph.
func (s *GraphService) DeleteEntity(ctx context.Context, id string) error {
	return s.store.DeleteEntity(ctx, id)
}

// UpsertRelation inserts or updates an address term relationship between two characters from a DTO request.
func (s *GraphService) UpsertRelation(ctx context.Context, req dtos.UpsertRelationRequest) error {
	return s.store.UpsertRelation(ctx, req.ToUpsertParams())
}

// ListRelations lists all character relationships for a project as DTOs.
func (s *GraphService) ListRelations(ctx context.Context, projectID string) ([]dtos.RelationDTO, error) {
	relations, err := s.store.ListRelationsByProject(ctx, projectID)
	if err != nil {
		return nil, err
	}
	dtosList := make([]dtos.RelationDTO, 0, len(relations))
	for _, r := range relations {
		dtosList = append(dtosList, dtos.ToRelationDTO(r))
	}
	return dtosList, nil
}

// LockRelation locks a relationship to prevent automatic AI modification.
func (s *GraphService) LockRelation(ctx context.Context, id string) error {
	return s.store.LockRelation(ctx, id)
}

// DeleteRelation deletes a character relationship by id.
func (s *GraphService) DeleteRelation(ctx context.Context, id string) error {
	return s.store.DeleteRelation(ctx, id)
}

// ListEffectiveRelationsAtChapter returns character relationships effective at chapterIndex.
// If chapterIndex <= 0, returns all relations across the project.
func (s *GraphService) ListEffectiveRelationsAtChapter(ctx context.Context, projectID string, chapterIndex int64) ([]dtos.RelationDTO, error) {
	allRels, err := s.store.ListRelationsByProject(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if chapterIndex <= 0 {
		dtosList := make([]dtos.RelationDTO, 0, len(allRels))
		for _, r := range allRels {
			dtosList = append(dtosList, dtos.ToRelationDTO(r))
		}
		return dtosList, nil
	}

	effectiveRels := make(map[[2]string]sqlc.CharacterRelation)
	for _, r := range allRels {
		if r.SinceChapter <= chapterIndex {
			pair := [2]string{r.FromChar, r.ToChar}
			if prev, ok := effectiveRels[pair]; !ok || r.SinceChapter > prev.SinceChapter {
				effectiveRels[pair] = r
			}
		}
	}

	dtosList := make([]dtos.RelationDTO, 0, len(effectiveRels))
	for _, r := range effectiveRels {
		dtosList = append(dtosList, dtos.ToRelationDTO(r))
	}
	sort.Slice(dtosList, func(i, j int) bool {
		if dtosList[i].FromChar != dtosList[j].FromChar {
			return dtosList[i].FromChar < dtosList[j].FromChar
		}
		return dtosList[i].ToChar < dtosList[j].ToChar
	})
	return dtosList, nil
}

// CheckGraphReadiness checks whether the character relationship graph has been scanned and is up-to-date
// for the given chapter and its corresponding volume before translation.
func (s *GraphService) CheckGraphReadiness(ctx context.Context, projectID string, chapterIndex int64) (*dtos.GraphReadinessDTO, error) {
	chaps, err := s.store.ListChaptersByProject(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("list chapters: %w", err)
	}

	allRels, err := s.store.ListRelationsByProject(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("list relations: %w", err)
	}

	allEnts, err := s.store.ListEntitiesByProject(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("list entities: %w", err)
	}

	res := &dtos.GraphReadinessDTO{
		ProjectID:             projectID,
		ChapterIndex:          chapterIndex,
		TotalProjectRelations: len(allRels),
		TotalProjectEntities:  len(allEnts),
	}

	// Find the target chapter to detect its title and volume tag
	var targetChap *sqlc.Chapter
	for _, c := range chaps {
		if c.ChapterIndex == chapterIndex {
			targetChap = &c
			break
		}
	}
	if targetChap == nil && len(chaps) > 0 {
		targetChap = &chaps[0]
	}

	var volTag string
	if targetChap != nil {
		m := regexp.MustCompile(`^\[(.*?)\]`).FindStringSubmatch(targetChap.Title)
		if len(m) > 1 {
			volTag = m[1]
		}
	}
	res.VolumeTag = volTag

	if volTag != "" {
		prefix := fmt.Sprintf("[%s]", volTag)
		var volStart, volEnd int64 = 999999, 0
		for _, c := range chaps {
			if strings.HasPrefix(c.Title, prefix) {
				if c.ChapterIndex < volStart {
					volStart = c.ChapterIndex
				}
				if c.ChapterIndex > volEnd {
					volEnd = c.ChapterIndex
				}
			}
		}
		res.VolumeStartChapter = volStart
		res.VolumeEndChapter = volEnd

		// Count relations whose since_chapter falls within this volume's range
		matchingCount := 0
		for _, r := range allRels {
			if r.SinceChapter >= volStart && r.SinceChapter <= volEnd {
				matchingCount++
			}
		}
		res.VolumeRelationsCount = matchingCount

		if matchingCount > 0 {
			res.IsReady = true
			res.HasVolumeScan = true
			res.Message = fmt.Sprintf("Relationship graph for Volume [%s] is ready (%d relations).", volTag, matchingCount)
		} else if len(allRels) > 0 {
			res.IsReady = false
			res.HasVolumeScan = false
			res.Message = fmt.Sprintf("Volume [%s] has no dedicated relationship graph yet (only relations inherited from earlier volumes). Running an AI scan before translating is recommended.", volTag)
		} else {
			res.IsReady = false
			res.HasVolumeScan = false
			res.Message = fmt.Sprintf("This project has no character relationship graph yet. Please run an AI scan for Volume [%s] before translating.", volTag)
		}
		return res, nil
	}

	// Untagged chapters
	if len(allRels) > 0 {
		res.IsReady = true
		res.HasVolumeScan = true
		res.Message = fmt.Sprintf("Relationship graph is ready (%d relations).", len(allRels))
	} else {
		res.IsReady = false
		res.HasVolumeScan = false
		res.Message = "This project has no character relationship graph yet. Running an AI character scan before translating is recommended."
	}

	return res, nil
}

// BuildSelectiveContext scans text and returns only relevant characters, address terms and glossary as a DTO.
func (s *GraphService) BuildSelectiveContext(ctx context.Context, projectID string, chapterIndex int64, chunkText string) (*dtos.SelectiveContextDTO, error) {
	selCtx, err := s.pg.BuildSelectiveContext(ctx, projectID, chapterIndex, chunkText)
	if err != nil {
		return nil, err
	}
	return dtos.ToSelectiveContextDTO(selCtx), nil
}

// isStoryChapter determines whether a chapter contains actual story prose rather than cover/inserts/TOC.
func isStoryChapter(title, content string) bool {
	tagRegex := regexp.MustCompile(`<[^>]*>`)
	plainText := strings.TrimSpace(tagRegex.ReplaceAllString(content, " "))
	if len([]rune(plainText)) < 100 {
		return false
	}

	lowerTitle := strings.ToLower(title)
	// NOTE: the Vietnamese entries are intentional input-matching data used to detect
	// cover / table-of-contents pages in Vietnamese source books.
	skipPatterns := []string{
		"cover", "insert", "title page", "copyright",
		"table of contents", "contents", "toc", "colophon",
		"newsletter", "bìa", "mục lục", "bản quyền", "hậu ký", "lời bạt",
		"目次", "あとがき", "封面", "插图", "目录", "版权", "制作人员", "后记", "奥付",
	}
	for _, p := range skipPatterns {
		if strings.Contains(lowerTitle, p) {
			return false
		}
	}
	return true
}

// filterStoryChapters extracts up to maxCount chapters that contain actual narrative text.
func filterStoryChapters(chaps []sqlc.Chapter, maxCount int) []sqlc.Chapter {
	var storyChaps []sqlc.Chapter
	for _, c := range chaps {
		content := c.TranslatedContent
		if content == "" {
			content = c.RawContent
		}
		if isStoryChapter(c.Title, content) {
			storyChaps = append(storyChaps, c)
			if len(storyChaps) >= maxCount {
				break
			}
		}
	}

	// Fallback if no chapter met strict criteria: pick chapters with longest content
	if len(storyChaps) == 0 && len(chaps) > 0 {
		sorted := make([]sqlc.Chapter, len(chaps))
		copy(sorted, chaps)
		for i := 0; i < len(sorted)-1; i++ {
			for j := i + 1; j < len(sorted); j++ {
				lenI := len(sorted[i].TranslatedContent) + len(sorted[i].RawContent)
				lenJ := len(sorted[j].TranslatedContent) + len(sorted[j].RawContent)
				if lenJ > lenI {
					sorted[i], sorted[j] = sorted[j], sorted[i]
				}
			}
		}
		limit := maxCount
		if len(sorted) < limit {
			limit = len(sorted)
		}
		storyChaps = sorted[:limit]
	}

	return storyChaps
}

// extractSmartVolumeExcerpt extracts narrative dialogues and scene setups across all story chapters in a volume or project.
func extractSmartVolumeExcerpt(chaps []sqlc.Chapter, maxTotalRunes int) string {
	var builder strings.Builder
	htmlTagRegex := regexp.MustCompile(`<[^>]*>`)

	perChapLimit := 2500
	if len(chaps) > 0 {
		perChapLimit = maxTotalRunes / len(chaps)
		if perChapLimit < 1200 {
			perChapLimit = 1200
		}
	}

	for _, c := range chaps {
		content := c.TranslatedContent
		if content == "" {
			content = c.RawContent
		}
		clean := htmlTagRegex.ReplaceAllString(content, " ")
		clean = strings.TrimSpace(clean)
		if len([]rune(clean)) < 150 {
			continue
		}

		lines := strings.Split(clean, "\n")
		var chapBuf strings.Builder
		chapBuf.WriteString(fmt.Sprintf("\n--- Chapter %d: %s ---\n", c.ChapterIndex, c.Title))

		runeCount := 0
		for _, line := range lines {
			l := strings.TrimSpace(line)
			if l == "" {
				continue
			}
			hasQuote := strings.ContainsAny(l, "“\"「”」『』")
			if hasQuote || runeCount < 600 {
				chapBuf.WriteString(l)
				chapBuf.WriteString("\n")
				runeCount += len([]rune(l))
			}
			if runeCount >= perChapLimit {
				break
			}
		}
		builder.WriteString(chapBuf.String())
		if len([]rune(builder.String())) >= maxTotalRunes {
			break
		}
	}

	res := builder.String()
	runes := []rune(res)
	if len(runes) > maxTotalRunes {
		return string(runes[:maxTotalRunes])
	}
	return res
}

func sanitizeField(text string) string {
	return strings.TrimSpace(text)
}

func normalizeAddressTerm(term string) string {
	t := strings.TrimSpace(term)
	t = strings.Trim(t, "\"'`“”「」『』")
	return strings.TrimSpace(t)
}

func normalizeSelfAddressTerm(term string) string {
	t := strings.TrimSpace(term)
	t = strings.Trim(t, "\"'`“”「」『』")
	return strings.TrimSpace(t)
}

// AutoScanEntities uses LLM to discover characters, roles, facts, and address terms from volume or chapter content.
func (s *GraphService) AutoScanEntities(ctx context.Context, req dtos.AutoScanRequest) (*dtos.AutoScanResultDTO, error) {
	if req.ProjectID == "" {
		return &dtos.AutoScanResultDTO{Success: false, Message: "project_id is required"}, nil
	}

	startTime := time.Now()

	client, modelName, err := s.resolveClient(ctx)
	if err != nil {
		return &dtos.AutoScanResultDTO{Success: false, Message: err.Error()}, nil
	}

	var textToScan strings.Builder
	scanMode := req.ScanMode
	if scanMode == "" {
		scanMode = "volume" // Default mode: Full Volume Scan
	}

	htmlTagRegex := regexp.MustCompile(`<[^>]*>`)
	var targetVolTag string
	var volStartChapter int64 = 1

	switch scanMode {
	case "current_chapter", "chapter":
		chap, err := s.store.GetChapterByIndex(ctx, req.ProjectID, req.ChapterIndex)
		if err != nil {
			return &dtos.AutoScanResultDTO{Success: false, Message: fmt.Sprintf("chapter %d not found: %v", req.ChapterIndex, err)}, nil
		}
		volStartChapter = req.ChapterIndex
		content := chap.TranslatedContent
		if content == "" {
			content = chap.RawContent
		}
		cleaned := htmlTagRegex.ReplaceAllString(content, " ")
		if len([]rune(strings.TrimSpace(cleaned))) < 50 {
			return &dtos.AutoScanResultDTO{
				Success: false,
				Message: fmt.Sprintf("Chapter %d (%s) does not contain enough narrative text (images only or too short). Please select a main story chapter.", chap.ChapterIndex, chap.Title),
			}, nil
		}
		textToScan.WriteString(fmt.Sprintf("Chapter %d: %s\n%s\n", chap.ChapterIndex, chap.Title, cleaned))

	case "all":
		chaps, err := s.store.ListChaptersByProject(ctx, req.ProjectID)
		if err != nil || len(chaps) == 0 {
			return &dtos.AutoScanResultDTO{Success: false, Message: "no chapters found in this project to scan"}, nil
		}
		storyChaps := filterStoryChapters(chaps, len(chaps))
		if len(storyChaps) == 0 {
			return &dtos.AutoScanResultDTO{Success: false, Message: "no valid story chapters found to scan"}, nil
		}
		volStartChapter = 1
		textToScan.WriteString(extractSmartVolumeExcerpt(storyChaps, 35000))

	case "pre_scan":
		chaps, err := s.store.ListChaptersByProject(ctx, req.ProjectID)
		if err != nil || len(chaps) == 0 {
			return &dtos.AutoScanResultDTO{Success: false, Message: "no chapters found in this project to scout"}, nil
		}
		storyChaps := filterStoryChapters(chaps, 3)
		if len(storyChaps) == 0 {
			return &dtos.AutoScanResultDTO{Success: false, Message: "no story chapters with content found to scout"}, nil
		}
		volStartChapter = 1
		for _, c := range storyChaps {
			content := c.TranslatedContent
			if content == "" {
				content = c.RawContent
			}
			cleaned := htmlTagRegex.ReplaceAllString(content, " ")
			textToScan.WriteString(fmt.Sprintf("Chapter %d: %s\n%s\n\n", c.ChapterIndex, c.Title, cleaned))
		}

	case "volume":
		fallthrough
	default:
		chaps, err := s.store.ListChaptersByProject(ctx, req.ProjectID)
		if err != nil || len(chaps) == 0 {
			return &dtos.AutoScanResultDTO{Success: false, Message: "no chapters found in this project to scan"}, nil
		}

		volTag := strings.TrimSpace(req.Volume)
		if volTag == "" && req.ChapterIndex > 0 {
			for _, c := range chaps {
				if c.ChapterIndex == req.ChapterIndex {
					m := regexp.MustCompile(`^\[(.*?)\]`).FindStringSubmatch(c.Title)
					if len(m) > 1 {
						volTag = m[1]
					}
					break
				}
			}
		}

		var targetChaps []sqlc.Chapter
		if volTag != "" {
			targetVolTag = volTag
			prefix := fmt.Sprintf("[%s]", volTag)
			for _, c := range chaps {
				if strings.HasPrefix(c.Title, prefix) {
					targetChaps = append(targetChaps, c)
				}
			}
		}
		if len(targetChaps) == 0 {
			targetChaps = chaps
		}

		storyChaps := filterStoryChapters(targetChaps, len(targetChaps))
		if len(storyChaps) == 0 {
			return &dtos.AutoScanResultDTO{Success: false, Message: "no story chapters with content found to scan in this volume"}, nil
		}
		volStartChapter = storyChaps[0].ChapterIndex
		textToScan.WriteString(extractSmartVolumeExcerpt(storyChaps, 35000))
	}

	rawText := textToScan.String()
	if len([]rune(rawText)) > 35000 {
		runes := []rune(rawText)
		rawText = string(runes[:35000])
	}

	proj, err := s.store.GetProject(ctx, req.ProjectID)
	targetLang := "Vietnamese"
	if err == nil && proj.TargetLang != "" {
		targetLang = llm.NormalizeLanguage(proj.TargetLang)
	}

	existingEnts, _ := s.store.ListEntitiesByProject(ctx, req.ProjectID)
	canonicalMap := make(map[string]string)
	var existingRoster []string

	for _, e := range existingEnts {
		canonicalMap[e.Name] = e.Name
		canonicalMap[strings.ToLower(e.Name)] = e.Name
		for _, a := range e.Aliases {
			canonicalMap[a] = e.Name
			canonicalMap[strings.ToLower(a)] = e.Name
		}
		descStr := ""
		if d, ok := e.Metadata["description"].(string); ok && d != "" {
			descStr = fmt.Sprintf(" (%s)", d)
		}
		existingRoster = append(existingRoster, fmt.Sprintf("- %s [%s, %s]%s", e.Name, e.Role, e.Gender, descStr))
	}

	var rosterBlock strings.Builder
	if len(existingRoster) > 0 {
		rosterBlock.WriteString("\n[EXISTING CANONICAL CHARACTERS IN THIS SERIES]:\n")
		for _, r := range existingRoster {
			rosterBlock.WriteString(r + "\n")
		}
		rosterBlock.WriteString("CRITICAL: If the excerpt mentions any of these characters (including nicknames, honorifics, or foreign script variants like Simplified Chinese), YOU MUST use the established canonical name listed above so character identities stay unified across volumes!\n")
	}

	systemPrompt := fmt.Sprintf(`You are an expert literary continuity director and character relationship profiler for novel publishing, working on a project with target language '%s'.

CRITICAL LANGUAGE RULE:
Even if the novel excerpt is in Japanese or Chinese, EVERY text field (description, facts, relation, call_as, self_call_as, tone) MUST be written in natural %s!
Zero raw Japanese (e.g. 先輩, お兄ちゃん, 私, 俺) and zero raw Chinese (e.g. 学长, 哥哥, 我, 恋人, 挚友) are allowed in descriptions, facts, relation, call_as, self_call_as, or tone!

Analyze the provided novel chapter excerpt and extract:
1. All characters:
   - "name": Standardized canonical character name. Preserve established canonical names; do not create duplicate entries for nicknames or Simplified Chinese characters.
   - "category": "character"
   - "gender": "male", "female", or "other"
   - "role": "protagonist", "deuteragonist", "heroine", "supporting", or "antagonist"
   - "aliases": array of nicknames, pet names, alternative spellings, or foreign script variants
   - "description": 1-2 sentence overview of the character's identity, profession/status, and personality traits, written in %s.
   - "facts": array of 2-5 canonical facts or core traits in %s.
2. All interpersonal addressing terms & relationships between character pairs:
   - "from_char": Name of speaker character (must match canonical name above)
   - "to_char": Name of listener/addressed character (must match canonical name above)
   - "relation": Nature of relationship, in %s (e.g. "lovers / couple", "best friends", "siblings", "senior - junior", "classmates").
   - "call_as": How from_char addresses to_char, STRICTLY in natural %s address style!
     * ABSOLUTELY NEVER output raw Japanese words (such as '先輩', '私', '俺', 'お兄ちゃん') or Chinese words (such as '学长', '我', '哥哥')!
     * NEVER invent names from other novels! Extract strictly from the provided text!
   - "self_call_as": How from_char refers to themselves, STRICTLY in %s!
     * NEVER output raw Japanese ('私', '俺', 'アタシ') or Chinese ('我')!
   - "tone": Specific interpersonal nuance and emotion in %s (e.g. "intimate, affectionate", "playful, teasing", "respectful, deferential", "casual, direct").
%s
Respond strictly with a JSON object in this format:
{
  "characters": [
    {
      "name": "...",
      "gender": "male",
      "role": "protagonist",
      "aliases": ["..."],
      "description": "...",
      "facts": ["...", "..."]
    }
  ],
  "relationships": [
    {
      "from_char": "...",
      "to_char": "...",
      "relation": "...",
      "call_as": "...",
      "self_call_as": "...",
      "tone": "..."
    }
  ]
}`, targetLang, targetLang, targetLang, targetLang, targetLang, targetLang, targetLang, targetLang, rosterBlock.String())

	userPrompt := fmt.Sprintf("Extract all characters, rich profile facts, and address term relationships from the following text:\n\n%s", rawText)

	resp, err := client.Generate(ctx, llm.CompletionRequest{
		Model: modelName,
		Messages: []llm.Message{
			{Role: llm.RoleSystem, Content: systemPrompt},
			{Role: llm.RoleUser, Content: userPrompt},
		},
		Temperature: 0.1,
	})
	if err != nil {
		return &dtos.AutoScanResultDTO{Success: false, Message: fmt.Sprintf("LLM scan failed: %v", err)}, nil
	}

	respContent := strings.TrimSpace(resp.Content)
	firstBrace := strings.Index(respContent, "{")
	lastBrace := strings.LastIndex(respContent, "}")
	if firstBrace == -1 || lastBrace == -1 || lastBrace <= firstBrace {
		return &dtos.AutoScanResultDTO{Success: false, Message: "LLM did not return a valid JSON object"}, nil
	}

	var parsed struct {
		Characters []struct {
			Name        string   `json:"name"`
			Category    string   `json:"category"`
			Gender      string   `json:"gender"`
			Role        string   `json:"role"`
			Aliases     []string `json:"aliases"`
			Description string   `json:"description"`
			Facts       []string `json:"facts"`
		} `json:"characters"`
		Relationships []struct {
			FromChar   string `json:"from_char"`
			ToChar     string `json:"to_char"`
			Relation   string `json:"relation"`
			CallAs     string `json:"call_as"`
			SelfCallAs string `json:"self_call_as"`
			Tone       string `json:"tone"`
		} `json:"relationships"`
	}

	if err := json.Unmarshal([]byte(respContent[firstBrace:lastBrace+1]), &parsed); err != nil {
		return &dtos.AutoScanResultDTO{Success: false, Message: fmt.Sprintf("failed to parse LLM JSON: %v", err)}, nil
	}

	firstSeen := volStartChapter
	if firstSeen <= 0 {
		firstSeen = 1
	}

	var discoveredEntities []dtos.EntityDTO
	for _, c := range parsed.Characters {
		name := strings.TrimSpace(c.Name)
		if name == "" {
			continue
		}
		rawInputName := name
		if canon, ok := canonicalMap[name]; ok {
			name = canon
		} else if canon, ok := canonicalMap[strings.ToLower(name)]; ok {
			name = canon
		}

		role := strings.TrimSpace(c.Role)
		if role == "" {
			role = "supporting"
		}
		gender := strings.TrimSpace(c.Gender)
		if gender == "" {
			gender = "other"
		}
		category := strings.TrimSpace(c.Category)
		if category == "" {
			category = "character"
		}

		metadata := map[string]any{}
		if desc := strings.TrimSpace(c.Description); desc != "" {
			metadata["description"] = sanitizeField(desc)
		}

		var finalFacts []string
		for _, f := range c.Facts {
			if tf := strings.TrimSpace(f); tf != "" {
				finalFacts = append(finalFacts, sanitizeField(tf))
			}
		}

		var finalAliases []string
		for _, a := range c.Aliases {
			if ta := strings.TrimSpace(a); ta != "" {
				finalAliases = append(finalAliases, ta)
			}
		}
		if rawInputName != name {
			finalAliases = append(finalAliases, rawInputName)
		}

		entFirstSeen := firstSeen
		existingEnt, err := s.store.GetEntityByName(ctx, req.ProjectID, name)
		if err == nil && existingEnt.ID != "" {
			if existingEnt.FirstSeenChapter < entFirstSeen {
				entFirstSeen = existingEnt.FirstSeenChapter
			}
			if existingFactsRaw, ok := existingEnt.Metadata["facts"].([]any); ok {
				seenFactMap := make(map[string]bool)
				for _, f := range existingFactsRaw {
					if str, ok := f.(string); ok && str != "" {
						seenFactMap[str] = true
						finalFacts = append(finalFacts, str)
					}
				}
				uniqueFacts := make([]string, 0, len(finalFacts))
				factDedup := make(map[string]bool)
				for _, f := range finalFacts {
					if !factDedup[f] {
						factDedup[f] = true
						uniqueFacts = append(uniqueFacts, f)
					}
				}
				finalFacts = uniqueFacts
			}
			for _, a := range existingEnt.Aliases {
				finalAliases = append(finalAliases, a)
			}
			if metadata["description"] == nil && existingEnt.Metadata["description"] != nil {
				metadata["description"] = existingEnt.Metadata["description"]
			}
		}

		// Deduplicate aliases
		aliasDedup := make(map[string]bool)
		uniqueAliases := make([]string, 0, len(finalAliases))
		for _, a := range finalAliases {
			if !aliasDedup[a] && a != name {
				aliasDedup[a] = true
				uniqueAliases = append(uniqueAliases, a)
			}
		}

		if len(finalFacts) > 0 {
			metadata["facts"] = finalFacts
		}

		ent := storage.Entity{
			ID:               fmt.Sprintf("ent_%s_%s", req.ProjectID, name),
			ProjectID:        req.ProjectID,
			Name:             name,
			Category:         category,
			Gender:           gender,
			Role:             role,
			Aliases:          uniqueAliases,
			FirstSeenChapter: entFirstSeen,
			Metadata:         metadata,
		}
		if err := s.store.UpsertEntity(ctx, ent); err == nil {
			discoveredEntities = append(discoveredEntities, dtos.ToEntityDTO(ent))
			canonicalMap[name] = name
			for _, a := range uniqueAliases {
				canonicalMap[a] = name
			}
		}
	}

	var discoveredRelations []dtos.RelationDTO
	for _, r := range parsed.Relationships {
		from := strings.TrimSpace(r.FromChar)
		to := strings.TrimSpace(r.ToChar)
		if canon, ok := canonicalMap[from]; ok {
			from = canon
		}
		if canon, ok := canonicalMap[to]; ok {
			to = canon
		}
		if from == "" || to == "" || from == to {
			continue
		}

		callAs := normalizeAddressTerm(r.CallAs)
		if callAs == "" {
			continue
		}
		selfCallAs := normalizeSelfAddressTerm(r.SelfCallAs)
		tone := sanitizeField(r.Tone)
		relNature := sanitizeField(r.Relation)
		if relNature != "" && !strings.Contains(tone, fmt.Sprintf("[%s]", relNature)) {
			if tone != "" {
				tone = fmt.Sprintf("[%s] %s", relNature, tone)
			} else {
				tone = fmt.Sprintf("[%s]", relNature)
			}
		}

		existingRel, err := s.store.GetRelationBetween(ctx, req.ProjectID, from, to, firstSeen)
		if err == nil && existingRel.IsLocked == 1 {
			continue
		}

		relID := fmt.Sprintf("rel_%s_%s_%s_ch%d", req.ProjectID, from, to, firstSeen)
		relParams := sqlc.UpsertRelationParams{
			ID:           relID,
			ProjectID:    req.ProjectID,
			FromChar:     from,
			ToChar:       to,
			CallAs:       callAs,
			SelfCallAs:   selfCallAs,
			SinceChapter: firstSeen,
			Tone:         tone,
			IsLocked:     0,
		}
		if err := s.store.UpsertRelation(ctx, relParams); err == nil {
			discoveredRelations = append(discoveredRelations, dtos.RelationDTO{
				ID:           relID,
				ProjectID:    req.ProjectID,
				FromChar:     from,
				ToChar:       to,
				CallAs:       callAs,
				SelfCallAs:   selfCallAs,
				SinceChapter: firstSeen,
				Tone:         tone,
				IsLocked:     false,
			})
		}
	}

	durationMs := time.Since(startTime).Milliseconds()

	if len(discoveredEntities) == 0 {
		return &dtos.AutoScanResultDTO{
			Success:          false,
			Message:          "The AI scan completed but no characters could be extracted from the selected chapters. Please choose chapters that contain the main story events.",
			EntitiesFound:    0,
			RelationsFound:   0,
			DurationMs:       durationMs,
			VolumeScanned:    targetVolTag,
			EffectiveChapter: firstSeen,
		}, nil
	}

	sec := float64(durationMs) / 1000.0
	var msg string
	if targetVolTag != "" {
		msg = fmt.Sprintf("AI scan succeeded for Volume [%s] (effective from chapter %d) in %.1fs! Extracted %d characters and %d relations into the L2 Graph.", targetVolTag, firstSeen, sec, len(discoveredEntities), len(discoveredRelations))
	} else {
		msg = fmt.Sprintf("AI scan succeeded (effective from chapter %d) in %.1fs! Extracted %d characters and %d relations into the L2 Graph.", firstSeen, sec, len(discoveredEntities), len(discoveredRelations))
	}

	return &dtos.AutoScanResultDTO{
		Success:             true,
		Message:             msg,
		EntitiesFound:       len(discoveredEntities),
		RelationsFound:      len(discoveredRelations),
		DurationMs:          durationMs,
		VolumeScanned:       targetVolTag,
		EffectiveChapter:    firstSeen,
		DiscoveredEntities:  discoveredEntities,
		DiscoveredRelations: discoveredRelations,
	}, nil
}
