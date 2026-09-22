package skills

// FewShotExample illustrates an input-output translation or decision pattern for a skill.
type FewShotExample struct {
	Input  string `json:"input" yaml:"input"`
	Output string `json:"output" yaml:"output"`
	Note   string `json:"note,omitempty" yaml:"note,omitempty"`
}

// SkillDefinition defines a modular capability unit of the autonomous translation agent.
type SkillDefinition struct {
	ID                   string           `json:"id" yaml:"id"`
	Name                 string           `json:"name" yaml:"name"`
	Description          string           `json:"description" yaml:"description"`
	Category             string           `json:"category" yaml:"category"`
	EstimatedTokenWeight int              `json:"estimated_token_weight" yaml:"estimated_token_weight"`
	WeightBadge          string           `json:"weight_badge" yaml:"weight_badge"` // "light" | "low", "medium", "heavy" | "high"
	IsEnabled            bool             `json:"is_enabled" yaml:"is_enabled"`
	SystemPromptTemplate string           `json:"system_prompt_template" yaml:"system_prompt_template"`
	ProhibitedRules      []string         `json:"prohibited_rules" yaml:"prohibited_rules"`
	CustomRules          string           `json:"custom_rules" yaml:"custom_rules"`
	FewShots             []FewShotExample `json:"few_shots" yaml:"few_shots"`
}

// DefaultBuiltinSkills returns the 6 canonical skills defined in Stage 9.
func DefaultBuiltinSkills() []SkillDefinition {
	return []SkillDefinition{
		{
			ID:                   "skill_style_scout",
			Name:                 "Style Scout",
			Description:          "Analyzes author narrative rhythm, lexical cadence, and stylistic density to maintain voice continuity.",
			Category:             "analysis",
			EstimatedTokenWeight: 80,
			WeightBadge:          "light",
			IsEnabled:            true,
			SystemPromptTemplate: "Analyze author narrative rhythm, lexical cadence, and stylistic density to maintain voice continuity across chapters. Calibrate target diction: for classical or period fiction, maintain authentic literary dignity (e.g. Sino-Vietnamese density in VI, elevated syntax in EN, classical idioms/四字熟語 in CJK); for modern light novels, preserve conversational velocity, irony, and colloquial rhythm.",
			ProhibitedRules: []string{
				"Do not impose an arbitrary narrative style that conflicts with the extracted Author Fingerprint.",
				"Do not alter the original pacing, sentence cadence, or emotional intensity of the scene.",
				"Do not force modern slang into classical period prose or archaic stiff grammar into contemporary dialogue.",
			},
			FewShots: []FewShotExample{
				{
					Input:  "Archaic classical wuxia/historical narrative segment with formal elevated diction",
					Output: "Calibrate elevated literary register: in Vietnamese, preserve 25-30% Sino-Vietnamese (Hán-Việt) vocabulary density and strict hierarchical honorifics; in English, employ formal dignified syntax and poetic cadence; in East Asian targets (ZH/JA/KO), maintain authentic period idioms (四字熟語/成语), classical verb endings, and courtly hierarchy.",
					Note:   "Preserves dignified classical literary register across all target languages",
				},
				{
					Input:  "Fast-paced first-person modern monologue with colloquial banter, irony, and internal thoughts",
					Output: "Capture conversational velocity: use natural contemporary phrasing and lively modal particles (e.g. 'à, nhỉ, đấy' in VI; 'you know, right' in EN; natural sentence-ending particles '〜じゃん, 〜よ' in JA; '거든, 잖아' in KO), avoiding textbook woodenness.",
					Note:   "Matches modern urban and light novel voice rhythm",
				},
			},
		},
		{
			ID:                   "skill_entity_extraction",
			Name:                 "Entity & Relationship Extractor",
			Description:          "Detects characters, honorifics, faction hierarchies, and dynamic relationships from chapter context.",
			Category:             "entity",
			EstimatedTokenWeight: 150,
			WeightBadge:          "medium",
			IsEnabled:            true,
			SystemPromptTemplate: "Detect characters, honorifics, faction hierarchies, and dynamic relationships without hallucination. Map kinship, seniority, and formal addresses accurately into target linguistic norms.",
			ProhibitedRules: []string{
				"Do not invent phantom entities that do not appear in the source text.",
				"Do not extrapolate relational dynamics beyond the explicit scope of the current chapter.",
				"Do not flatten or invert hierarchical address terms between seniors/juniors or masters/disciples.",
			},
			FewShots: []FewShotExample{
				{
					Input:  "Detect junior disciple addressing senior mentor",
					Output: "Entities: Disciple [Protagonist, Male], Mentor [Mentor, Male]. Relation: Disciple addresses Mentor with deep reverence (e.g. 'Master' / 'Teacher').",
					Note:   "Accurate hierarchical extraction with target-appropriate honorific mapping",
				},
				{
					Input:  "Junior colleague speaking playfully to senior colleague",
					Output: "Entities: Junior Colleague [Female supporting], Senior Colleague [Male supporting]. Dynamic: Affectionate junior deference ('Senior' / 'Big Brother').",
					Note:   "Dynamic interpersonal register",
				},
			},
		},
		{
			ID:                   "skill_literary_translator",
			Name:                 "Literary Translator",
			Description:          "Produces faithful, highly expressive literary translations with natural cadence, anti-translationese de-calquing, idiom awareness, strict capitalization orthography, and clean typography.",
			Category:             "translation",
			EstimatedTokenWeight: 350,
			WeightBadge:          "medium",
			IsEnabled:            true,
			SystemPromptTemplate: "Translate literarily with natural narrative cadence, emotional depth, and consistent stylistic register. ACTIVELY DE-CALQUE: Never translate word-by-word; dissolve foreign syntax into natural, idiomatic target language structures. Accurately convey idioms, metaphors, and humor by their figurative meaning (e.g. 'in her birthday suit' means completely naked, NOT literal 'birthday suit'; 'triumphant smile' means a smug/victorious grin). Dialogue must sound like authentic living speech with fitting tone particles and consistent relational hierarchy.",
			ProhibitedRules: []string{
				"Never omit sentences, summarize paragraphs, or inject extraneous translator commentary.",
				"Never use raw machine translation, stiff word-by-word convert phrasing, or foreign syntactic calques.",
				"Never translate foreign idioms, colloquialisms, or metaphors literally into nonsensical phrases (e.g. translating 'birthday suit' literally instead of 'completely naked', or translating CJK idioms literally character-by-character without resolving figurative intent).",
				"Never copy foreign preposition attachments literally (e.g. translating 'dabbed at it with a cloth' with awkward dangling prepositions instead of active verbal descriptions).",
				"Never invert relational hierarchy or arbitrarily alter established pronoun dynamics between characters.",
				"In Latin-script targets (Vietnamese, English, French, Spanish, German): Never capitalize mid-sentence pronouns, common nouns, or cultivation ranks mid-sentence. Only capitalize sentence starts and true proper names.",
				"Never leave raw source scripts or incompatible foreign punctuation marks (such as raw CJK '。', '「」', '『』' in Western text, or Western semicolons ';' in CJK prose).",
				"Never output chatbot conversational chatter, meta-commentary, or explanatory notes (e.g. 'Here is the translation...', 'Sure!'). If a segment contains only images or HTML/SVG without text, output ONLY the preserved markup.",
				"In 1st-person POV narration, third-person pronouns referring to other characters must strictly follow the established relationship in the character graph (avoiding mismatched or overly detached pronouns when an intimate relationship is defined).",
				"Never use raw convertese or hyperbolic anachronistic cliches in modern fiction (e.g. translating modern academic terms like 単位制 into natural target equivalents like 'credit system', never archaic calques).",
				"Never infantilize teenage or adult characters by prepending 'bé' to their proper names (e.g. 'bé Akari', 'bé Nanami' are strictly forbidden; render simply by their proper name or natural relational pronouns like 'em', 'em ấy', 'cô nàng').",
				"Never retain raw Japanese honorific suffixes ('-chan', '-san', '-kun') attached to names in Latin-script targets; render names directly or adapt using natural relational pronouns.",
			},
			FewShots: []FewShotExample{
				{
					Input:  "As I opened my eyes, I was greeted by the sight of Alice, in her birthday suit, flashing a triumphant smile.",
					Output: "Vừa mở mắt định thần, đập vào mắt tôi là cảnh tượng Alice đang trần như nhộng, trên môi nở nụ cười đắc thắng.",
					Note:   "De-calques passive 'was greeted by the sight of', correctly renders idiom 'birthday suit' (trần như nhộng) and 'triumphant smile' (nụ cười đắc thắng)",
				},
				{
					Input:  "「……べ、別にアンタのために作ったわけじゃないんだからね！」彼女は顔を真っ赤にしてそっぽを向いた。",
					Output: "\"I-It’s not like I made this for you or anything, okay?!\" Her face flushed crimson as she abruptly turned away.",
					Note:   "JA->EN: Authentic tsundere role voice, natural contractions, vivid descriptive verbs instead of wooden literalism",
				},
				{
					Input:  "He was greeted by a lavish banquet and surrounded by countless disciples enthusiastically shouting his name.",
					Output: "映入眼帘的是一场极其奢华的盛宴，周围无数群情激昂的宗门弟子正齐声高呼着他的名字。",
					Note:   "EN->ZH: De-Europeanization: converts passive 'was greeted by' into sensory '映入眼帘的是', eliminates redundant '被/的', applies rhythmic Chinese phrasing",
				},
				{
					Input:  "“Boy, you look sleepy! For a second there I thought you were gonna pass out,” she dabbed at his cheek with a napkin.",
					Output: "「もう、すごく眠そう！　一瞬、そのまま気絶しちゃうんじゃないかって心配したんだから」彼女はそう言いながら、ナプキンで彼の頬をそっと拭った。",
					Note:   "EN->JA: Drops redundant subjects, natural role language with light novel rhythm, converts 'for a second there' idiomatically, and uses standard Japanese dialogue punctuation",
				},
				{
					Input:  "夜色如水，月光洒落在他清秀的脸庞上。",
					Output: "Màn đêm phẳng lặng như nước, ánh trăng bàng bạc phủ nhẹ lên gương mặt thanh tú của chàng.",
					Note:   "ZH->VI: Fluent literary Vietnamese with proper lowercase pronoun 'chàng' and balanced Sino-Vietnamese diction",
				},
				{
					Input:  "「どうしたの？　そんな怖い顔で見つめられちゃ、落ち着かないよ」",
					Output: "“왜 그래? 그렇게 무서운 얼굴로 쳐다보면 좌불안석이잖아.”",
					Note:   "JA->KO: Removes passive translationese, natural conversational ending and idiom",
				},
				{
					Input:  "「お兄ちゃんが借金のカタに私をあげるって言ったの。これからよろしくね」彼女はとんでもないことを言い放った。",
					Output: "“Anh trai em bảo em đến đây, làm vật thế chấp cho món nợ. Từ giờ mong anh giúp đỡ ạ.”\nEm ấy buông ra một câu động trời như thế.",
					Note:   "JA->VI: 1st-person POV third-person pronoun matches established 'anh - em' relationship in character graph (e.g. 'Em ấy' instead of detached 'Cô ấy'), and translates 'とんでもないこと' naturally as 'câu động trời' instead of wuxia convertese.",
				},
				{
					Input:  "高校から大学に上がって、一番の変化といえば単位制になったことだろう。",
					Output: "Từ cấp ba bước lên đại học, thay đổi lớn nhất chính là việc chuyển sang hệ thống tín chỉ.",
					Note:   "JA->VI: Translates '単位制' naturally as 'hệ thống tín chỉ' instead of raw Sino-Vietnamese convertese 'học chế tín chỉ'.",
				},
				{
					Input:  "<div><svg xmlns=\"...\"><image xlink:href=\".../cover.jpeg\"/></svg></div>",
					Output: "<div><svg xmlns=\"...\"><image xlink:href=\".../cover.jpeg\"/></svg></div>",
					Note:   "Pure image/cover markup: preserves markup without any chatbot chatter, meta-commentary, or excuses.",
				},
			},
		},
		{
			ID:                   "skill_shadow_critic",
			Name:                 "Shadow Critic",
			Description:          "Inspects drafts in real time to catch word-by-word translationese, idiom errors, pronoun drift, glossary compliance, untranslated CJK, arbitrary capitalization, and illegal punctuation.",
			Category:             "critic",
			EstimatedTokenWeight: 850,
			WeightBadge:          "heavy",
			IsEnabled:            true,
			SystemPromptTemplate: "Evaluate drafts in real time. Strictly issue REVISE if: (1) the draft suffers from word-by-word translation, literal idiom errors, or foreign syntax calques; (2) untranslated source characters or foreign script leak; (3) relational pronouns conflict with character graph; (4) arbitrary capitalization occurs; or (5) a sentence ends with a semicolon (;) or invalid punctuation.",
			ProhibitedRules: []string{
				"Do not PASS any draft that translates foreign idioms or metaphors literally into nonsensical or stiff phrasing.",
				"Do not PASS any draft containing foreign syntax calques or awkward robotic sentence structures that read like machine translation.",
				"Do not PASS any draft containing untranslated raw source script or incompatible punctuation.",
				"Do not PASS any draft containing inverted character pronouns or contradictory address terms.",
				"Do not PASS any draft where a sentence ends with a semicolon (;) instead of a proper terminal mark.",
				"Do not PASS any draft with arbitrary mid-sentence capitalization of common nouns or pronouns.",
				"Do not PASS any draft containing conversational chatbot chatter or meta-talk (such as 'Here is the translation:', 'Sure, below is...', 'Certainly!').",
				"Do not PASS any draft with stiff raw convertese or hyperbolic anachronisms in modern fiction (e.g. archaic calques instead of natural modern terms).",
				"Do not PASS any draft where the 1st-person narrator uses distant pronouns to refer to someone with whom an intimate relationship is established in the character graph.",
				"Do not PASS any draft that infantilizes teenage or adult characters by prepending 'bé' to their names (e.g. 'bé Akari', 'bé Nanami' must be REVISED to 'Akari', 'Nanami', or natural relational pronouns).",
			},
			FewShots: []FewShotExample{
				{
					Input:  "Draft: \"Anh trai em bảo em đến đây, làm vật thế chấp cho món nợ. Từ giờ mong anh giúp đỡ ạ.\"\nCô ấy buông ra một câu kinh thiên động địa như thế.\nSource: 「お兄ちゃんが借金のカタに私をあげるって言ったの。これからよろしくね」彼女はとんでもないことを言い放った。",
					Output: "REVISE -> Reason: Protagonist calls heroine 'em' in character relation, but narration coldly refers to her as 'Cô ấy' (should be 'Em ấy' / 'Cô bé'); also uses archaic wuxia cliche 'kinh thiên động địa' for modern slice-of-life (should be 'câu động trời' / 'chuyện không tưởng'). Correction: “Anh trai em bảo em đến đây, làm vật thế chấp cho món nợ. Từ giờ mong anh giúp đỡ ạ.”\nEm ấy buông ra một câu động trời như thế.",
					Note:   "Enforces 1st-person romance pronoun intimacy and eliminates Sino-Vietnamese convertese",
				},
				{
					Input:  "Draft: Thay đổi lớn nhất chính là việc chuyển sang học chế tín chỉ.\nSource: 一番の変化といえば単位制になったことだろう。",
					Output: "REVISE -> Reason: Sino-Vietnamese convertese 'học chế tín chỉ' is unnatural in modern Vietnamese. Correction: Thay đổi lớn nhất chính là việc chuyển sang hệ thống tín chỉ.",
					Note:   "Catches raw convertese for Japanese educational terms",
				},
				{
					Input:  "Draft: Con bé mãn nguyện trong bộ đồ sinh nhật của mình... và nở nụ cười đỉnh cao thế giới;\nSource: She looked all satisfied in her birthday suit, flashing a triumphant smile.",
					Output: "REVISE -> Reason: Literal translation of idiom 'birthday suit' as 'bộ đồ sinh nhật' (should be 'trần như nhộng'), 'triumphant smile' as 'nụ cười đỉnh cao thế giới' (should be 'nụ cười đắc thắng'), and illegal trailing semicolon. Correction: Con bé tỏ vẻ mãn nguyện trong bộ dạng trần như nhộng, trên môi nở nụ cười đắc thắng.",
					Note:   "Catches word-by-word idiom calques and syntax errors in Vietnamese",
				},
				{
					Input:  "Draft: 他被一个巨大的阴影给笼罩了，在心里感觉到了极度的恐惧。\nSource: A colossal shadow loomed over him, chilling him to the bone.",
					Output: "REVISE -> Reason: Unnatural Europeanized passive '被...给笼罩' and stiff literal phrasing. Correction: 一道庞大可怖的黑影骤然将他笼罩，令他刹那间如坠冰窖。",
					Note:   "Catches awkward Europeanized passive in Chinese and improves literary tension",
				},
				{
					Input:  "Draft: \"It is not like I did it for you, okay?\" She was gotten red in the face;「...」\nSource: 「……べ、別にアンタのために作ったわけじゃないんだからね！」",
					Output: "REVISE -> Reason: Stiff uncontracted dialogue lacking tsundere cadence, awkward passive voice ('was gotten red'), and illegal raw Japanese punctuation and trailing semicolon. Correction: \"I-It's not like I made this for you or anything, okay?!\" Blushing furiously, she averted her eyes.",
					Note:   "Catches stiff robotic English translation and foreign punctuation",
				},
				{
					Input:  "Draft: Hắn nhìn xem đối phương nói: 'Ngươi đi đâu;';\nSource: 他看着对方道：'你去哪？'",
					Output: "REVISE -> Reason: Sentence terminates with illegal semicolon (;), contains redundant word 'xem', and unstandard quotes. Correction: Hắn nhìn đối phương hỏi: “Huynh đi đâu thế?”",
					Note:   "Catches illegal semicolon, quote formatting, and awkward phrasing",
				},
				{
					Input:  "Draft: 「彼は彼女の手を取りました、そして走りました。」\nSource: He took her hand and ran.",
					Output: "REVISE -> Reason: Excessive third-person pronouns ('彼/彼女') and unnatural coordinate conjunction. Correction: 「彼女の手を引いて駆け出した。」",
					Note:   "Catches unnatural pronoun repetition and mechanical translationese in Japanese",
				},
			},
		},
		{
			ID:                   "skill_agentic_researcher",
			Name:                 "Agentic Researcher",
			Description:          "Proactively invokes tools to lookup character relations, world lore, past context, and web references when uncertain.",
			Category:             "research",
			EstimatedTokenWeight: 900,
			WeightBadge:          "heavy",
			IsEnabled:            true,
			SystemPromptTemplate: "Proactively invoke tools when encountering uncertainty: (1) Call 'lookup_character_relation' if unsure about pronoun/address terms between characters; (2) Call 'lookup_world_lore' if encountering unfamiliar setting factions, ranks, or world rules; (3) Call 'search_book_context' for past story facts; (4) Call 'web_lookup' for rare cultural idioms. Resume translating with retrieved context, adapting appropriately to target language norms.",
			ProhibitedRules: []string{
				"Do not guess pronouns or address terms when they can be precisely queried via 'lookup_character_relation'.",
				"Do not guess world factions, magic ranks, or tech tiers without querying 'lookup_world_lore'.",
				"Do not make redundant queries outside the immediate scope of the segment being translated.",
			},
			FewShots: []FewShotExample{
				{
					Input:  "Scenario: Unsure how Character A addresses Character B in conversation.",
					Output: "ToolCall: lookup_character_relation(character_a=\"Speaker\", character_b=\"Listener\") -> Retrieved: Speaker addresses Listener with warm, respectful seniority. -> Target Translation: [VI: “Chị, lần này cảm ơn chị đã ra tay giúp đỡ.” / EN: “Sister, thank you for stepping in to help me.” / JA: “姉さん、助けていただき感謝します。”]",
					Note:   "Dynamic character relation lookup with target-appropriate speech register",
				},
				{
					Input:  "Scenario: Encountering unfamiliar world lore 'Sanctuary of Magic' and rank 'Forbidden Spell'.",
					Output: "ToolCall: lookup_world_lore(query=\"Sanctuary of Magic\") -> Retrieved: Supreme magical order; Forbidden Spell is 10th-tier strategic destruction magic. -> Target Translation: [VI: 'Đại Ma Đạo Sư / Cấm Chú' / EN: 'Archmage / Forbidden Spell' / JA: '大魔導士 / 禁呪']",
					Note:   "World Bible lore lookup ensuring canonical terminology consistency",
				},
				{
					Input:  "Terminology: 'Foundation Establishment Stage' (筑基期)",
					Output: "ToolCall: search_book_context(query=\"筑基\") -> Retrieved: Second cultivation realm following Qi Condensation. Canonical translation: [VI: 'Trúc Cơ Kỳ' / EN: 'Foundation Establishment' / JA: '築基期' / KO: '축기기'].",
					Note:   "FTS5 past context retrieval verifying canonical story glossary",
				},
			},
		},
		{
			ID:                   "skill_foreign_sanitizer",
			Name:                 "Foreign Sanitizer",
			Description:          "Sanitizes residual foreign script, fixes irregular punctuation, corrects capitalization, and enforces typography standards.",
			Category:             "post_process",
			EstimatedTokenWeight: 120,
			WeightBadge:          "light",
			IsEnabled:            true,
			SystemPromptTemplate: "Sanitize residual foreign CJK script, fix irregular punctuation (e.g. semicolons at sentence ends), remove arbitrary mid-sentence capitalization, and standardize quotation marks according to target language typography.",
			ProhibitedRules: []string{
				"Do not alter the narrative meaning or syntactic structure of already approved sentences.",
				"Do not leave stray semicolons (;) or commas (,) at the end of sentences or paragraphs.",
				"Do not leave unclosed quotation marks or duplicate punctuation (;;, .., ??).",
				"Do not use Asian bracket quotes 「」 in Latin prose (use “” or \"\"), and do not use Western ASCII punctuation in CJK prose (use ，。？！「」).",
			},
			FewShots: []FewShotExample{
				{
					Input:  "He said「Hello」and walked in;",
					Output: "He said: “Hello” and walked in. (VI: “Anh ấy nói: “Xin chào” rồi bước vào.”)",
					Note:   "Replaces Japanese quote brackets with standard quotation marks, eliminates illegal trailing semicolon, uses standard period",
				},
				{
					Input:  "「你好!」 he said; then left...",
					Output: "ZH: “你好！”他说完便离开了。 / JA: 「こんにちは！」と言って、彼は立ち去った。",
					Note:   "Replaces half-width western punctuation with standard East Asian full-width punctuation and standardized dialogue marks",
				},
			},
		},
	}
}
