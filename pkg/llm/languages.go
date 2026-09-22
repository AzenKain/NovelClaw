package llm

import (
	"fmt"
	"strings"
)

// NormalizeLanguage maps language codes or common names to canonical English language names.
func NormalizeLanguage(lang string) string {
	cleaned := strings.TrimSpace(strings.ToLower(lang))
	switch cleaned {
	case "vi", "vie", "vietnamese", "tiếng việt", "tieng viet":
		return "Vietnamese"
	case "en", "eng", "english", "tiếng anh", "tieng anh":
		return "English"
	case "ja", "jpn", "japanese", "tiếng nhật", "tieng nhat", "nihongo":
		return "Japanese"
	case "zh", "zho", "chi", "chinese", "tiếng trung", "tieng trung", "zhongwen":
		return "Chinese"
	case "ko", "kor", "korean", "tiếng hàn", "tieng han":
		return "Korean"
	case "fr", "fra", "fre", "french", "tiếng pháp":
		return "French"
	case "de", "deu", "ger", "german", "tiếng đức":
		return "German"
	case "es", "spa", "spanish", "tiếng tây ban nha":
		return "Spanish"
	case "ru", "rus", "russian", "tiếng nga":
		return "Russian"
	default:
		if lang == "" {
			return "Vietnamese"
		}
		return lang
	}
}

// GetLanguageScriptRule generates strict transliteration, capitalization, and typography rules for a target language.
func GetLanguageScriptRule(sourceLang, targetLang string) string {
	tgtNorm := NormalizeLanguage(targetLang)
	srcNorm := NormalizeLanguage(sourceLang)

	switch tgtNorm {
	case "Vietnamese":
		return "CRITICAL SCRIPT, CAPITALIZATION & TYPOGRAPHY RULES FOR VIETNAMESE:\n" +
			"1. SCRIPT & TRANSLITERATION:\n" +
			"   - Output MUST be 100% in natural Vietnamese. ABSOLUTELY ZERO raw Japanese Kanji, Hiragana, Katakana, or Chinese Hanzi characters may remain in the translated text!\n" +
			"   - Japanese Proper Names & Terms: Transcribe all Kanji names into customary Romaji (following standard Hepburn romanization or the injected [SELECTIVE CONTEXT] glossary).\n" +
			"   - Honorifics & Relational Address Terms (Strict Systemic Consistency):\n" +
			"     * Adapt Japanese honorifics naturally into Vietnamese relational terms ('tiền bối', 'anh', 'em', 'chị', 'cô', 'bạn') conforming strictly to character relationships.\n" +
			"     * STRICT CONSISTENCY (NO HYBRID / MIXED ADDRESS TERMS):\n" +
			"       - When the translation establishes Vietnamese relational terms ('tiền bối', 'anh', 'em', 'hậu bối'), MAINTAIN 100% consistency throughout the entire work!\n" +
			"       - ABSOLUTELY DO NOT mix untranslated Romaji such as 'senpai', 'sempai', 'kouhai', 'sensei' into dialogue or narrative when characters address each other as 'tiền bối' / 'anh'.\n" +
			"       - Adhere strictly to the interpersonal relationships and address terms established in [SELECTIVE CONTEXT] or [CHARACTER RELATIONS].\n" +
			"     * STRICT BAN ON INFANTILIZING PREFIXES & RAW SUFFIXES (NO 'BÉ + [NAME]', NO '-CHAN'/'-SAN'):\n" +
			"       - NEVER translate Japanese diminutive suffix '-chan' by prepending 'bé' to teenage or adult character names (e.g. 'bé Akari', 'bé Nanami', 'bé Yua' are STRICTLY FORBIDDEN and unnatural/cringe).\n" +
			"       - NEVER retain raw Japanese honorific suffixes ('-chan', '-san', '-kun') attached to names in Romaji (e.g. 'Akari-chan' -> simply 'Akari' or 'em ấy'; 'Nanami-san' -> 'Nanami' or 'bạn Nanami'). Render character names directly or use natural relational pronouns ('em', 'em ấy', 'cô nàng') as appropriate to context.\n\n" +
			"2. STRICT CAPITALIZATION RULES (PREVENT ARBITRARY CAPITALIZATION):\n" +
			"   - ONLY capitalize: (a) the first letter of a sentence (after '.', '!', '?', '...', or dialogue dash '— '); and (b) true proper nouns (specific personal names, specific geographical place names, and specific official organizations/sects/factions).\n" +
			"   - NEVER capitalize mid-sentence pronouns: 'hắn', 'nàng', 'y', 'ngươi', 'ta', 'chàng', 'thiếp', 'chúng ta', 'bọn họ' MUST remain in lowercase unless starting a sentence. NEVER write 'Hắn nhìn thấy Nàng' or 'Ngươi đi cùng Ta'.\n" +
			"   - NEVER capitalize mid-sentence common nouns, kinship terms, cultivation realms, martial arts, or item categories: 'trúc cơ', 'kim đan', 'đấu tông', 'trưởng lão', 'tông chủ', 'sư phụ', 'đệ tử', 'pháp bảo', 'công pháp', 'đan dược' MUST remain lowercase. NEVER write 'Kim Đan Kỳ' or 'Đại Trưởng Lão' mid-sentence.\n" +
			"   - NEVER apply English-style Title Case (capitalizing every word) to Vietnamese sentences, headings, skills, or dialogue.\n\n" +
			"3. STRICT PUNCTUATION & TYPOGRAPHY RULES (FORBIDDEN WEIRD PUNCTUATION):\n" +
			"   - Every complete narrative sentence MUST terminate with a period (.), exclamation mark (!), question mark (?), or ellipsis (...).\n" +
			"   - ABSOLUTELY FORBIDDEN: NEVER terminate a sentence or paragraph with a semicolon (;) or comma (,). Semicolons (;) are strictly reserved for separating closely related independent clauses within a compound sentence; NEVER use them at the end of narrative sentences or paragraphs.\n" +
			"   - Convert all raw CJK punctuation marks into standard target typography:\n" +
			"     * Replace Chinese full stop '。' with standard period '.'\n" +
			"     * Replace Chinese commas '，' and enumeration commas '、' with standard comma ','\n" +
			"     * Replace Chinese colons '：' with ':'\n" +
			"     * Replace Chinese semicolons '；' with ',' (or separate into two sentences with '.'). NEVER leave a semicolon at sentence end.\n" +
			"     * Replace Chinese quotation brackets '「...」' or '『...』' with standard quotation marks '“...”' or narrative dialogue dashes ('— ').\n" +
			"     * Replace Chinese brackets '《...》' with italics or quotation marks for book/manual/technique titles.\n" +
			"     * Replace Chinese brackets '【...】' with standard brackets '[...]' for system/game notifications or plain text.\n" +
			"   - Standard Spacing: NO space before punctuation marks (',', '.', ';', ':', '!', '?'); exactly ONE space after punctuation marks (unless at the end of a paragraph).\n" +
			"   - Ellipsis: Use standard 3 dots '...' or Unicode '…'. NEVER output 2 dots '..' or 4+ dots '.....'.\n" +
			"   - No AI artifacts: NEVER output raw code formatting, bullet point semicolons, or weird decorative symbols inside narrative prose.\n\n" +
			"4. LITERARY FLUENCY, IDIOMATIC EXPRESSION & ANTI-TRANSLATIONESE:\n" +
			"   - DE-CALQUING & SYNTACTIC NATURALIZATION: Do NOT mirror source language clause order, preposition chains, or passive voice slavishly.\n" +
			"     * Convert awkward passive/nominal patterns into active sensory descriptions: 'I was greeted by the sight of Alice' -> 'Đập vào mắt tôi là cảnh tượng Alice...' or 'Trước mắt tôi là Alice...', NOT 'Khung cảnh đón chào tôi là Alice'.\n" +
			"     * Dissolve literal preposition attachments: 'dabbed at mouth with it' -> 'dùng nó chấm nhẹ lên khóe miệng', NOT 'chấm lên miệng bằng nó'; 'as much as possible, he liked' -> 'bất cứ khi nào có thể, anh đều thích...', NOT 'càng nhiều càng tốt, anh thích'.\n" +
			"     * Eliminate unnatural abstract pronouns for objects/body parts: NEVER end sentences with robotic 'kích thích chúng' (exciting them) or 'bằng nó' (with it).\n" +
			"   - FIGURATIVE SENSE & IDIOMS: NEVER translate idioms, metaphors, or colloquial phrases word-by-word into nonsensical literal terms!\n" +
			"     * Foreign idioms must be rendered by their contextual meaning and cultural equivalents (e.g. English 'in her birthday suit' means completely naked -> translate as 'trần như nhộng' / 'không một mảnh vải che thân', ABSOLUTELY NEVER translate as 'bộ đồ sinh nhật'!).\n" +
			"     * English 'triumphant smile' -> 'nụ cười đắc thắng' / 'cười đắc ý', NEVER translate as 'nụ cười đỉnh cao thế giới' or 'đỉnh cao vô địch'!\n" +
			"     * English 'for a second there' -> 'suýt nữa thì' / 'có lúc tôi tưởng', NEVER translate as 'trong một giây đấy'!\n" +
			"     * English 'nobody could possibly lodge any complaint against her' -> 'chẳng ai có thể chê vào đâu được' / 'đẹp đến mức không thể bắt bẻ lấy một lời'.\n" +
			"   - AUTHENTIC DIALOGUE & LIVING CADENCE: Characters must speak naturally in living Vietnamese dialogue with fitting conversational particles ('á', 'nè', 'hả', 'đấy', 'chứ', 'nhé', 'kia mà') and lively tone. Relational honorifics must be consistent throughout the scene (e.g. if Character A is 'anh' to Character B's 'em', do not mix calling 'cháu' with addressing 'anh').\n" +
			"   - THIRD-PERSON PRONOUN COHERENCE WITH CHARACTER RELATIONS:\n" +
			"     * In 1st-person POV narration and internal thoughts, all third-person pronouns referring to other characters MUST strictly align with the established relationship in [CHARACTER RELATIONS] / [SELECTIVE CONTEXT]:\n" +
			"       - Do not mechanically translate '彼女' / '彼' into detached 'cô ấy' / 'anh ấy' if the defined relationship between narrator and target is intimate (e.g. if the relation is 'anh - em', narrative should use matching relational pronouns like 'em ấy', or use character name according to tone, avoiding distant 'cô ấy').\n" +
			"       - Conversely, if the relationship is classmates, peers, or strangers, use appropriate neutral pronouns (such as 'cô ấy', 'cậu ấy', 'bạn ấy').\n" +
			"       - Core principle: Narration and internal thoughts must always remain faithful to the direct interpersonal relationship defined in the Character Graph, without arbitrarily altering relational intimacy.\n" +
			"   - ANTI-SINO-VIETNAMESE OVERUSE & CONVERTESE:\n" +
			"     * Strictly prohibit importing raw Sino-Japanese / Sino-Vietnamese calques into modern slice-of-life / school life settings:\n" +
			"       - '単位制' (tan'i-sei) -> MUST BE TRANSLATED AS 'hệ thống tín chỉ' / 'học theo tín chỉ' (NEVER 'học chế tín chỉ' or 'chế độ đơn vị')!\n" +
			"       - '履修' -> 'đăng ký môn học' / 'học phần'.\n" +
			"       - '講義' -> 'buổi học' / 'tiết học' / 'bài giảng'.\n" +
			"       - 'サークル' -> 'câu lạc bộ' (CLB), 'バイト' -> 'làm thêm' / 'việc làm thêm'.\n" +
			"     * Prohibit hyperbolic wuxia/xianxia cliches ('kinh thiên động địa', 'sát na', 'kinh hãi') in modern romantic/school novels:\n" +
			"       - 'とんでもないこと' -> translate as 'lời lẽ động trời' / 'chuyện không tưởng' / 'lời lẽ quá quắt', ABSOLUTELY NEVER as 'kinh thiên động địa'!\n" +
			"   - LOGICAL COHERENCE & MODAL NEGATION INTEGRITY (STRICTLY PREVENT LOGIC INVERSIONS):\n" +
			"     * Exercise extreme vigilance with modal auxiliaries and negative structures (especially Japanese '〜なかったはず', '〜ないはず', '〜はずがない', '〜ていいはず', '〜わけがない'):\n" +
			"       - 'アイツはこんなこと一言も言ってなかったはずなのに' expresses certainty of recollection or counterfactual indignation: translate into natural Vietnamese as 'rõ ràng là cái gã hiểu em ấy hơn tôi kia đâu có hé nửa lời về chuyện này cơ mà!' OR 'đáng lẽ một chuyện như vậy thì cái gã hiểu rõ em ấy hơn tôi kia ít ra cũng phải hé trước nửa lời mới phải chứ!'.\n" +
			"       - ABSOLUTELY NEVER combine words mechanically into absurd logic-inverted monstrosities like: 'lẽ ra phải không hé nửa lời mới đúng chứ' (which erroneously turns certainty of non-occurrence into an obligation to remain silent, contradicting the protagonist's astonishment)!\n" +
			"       - Always preserve the genuine psychological reality of the character: the character is shocked because his friend did NOT give a heads-up, NOT demanding that his friend remain silent.\n" +
			"     * STRICT BAN ON DIALOGUE PRONOUN LEAKAGE INTO NARRATION:\n" +
			"       - Casual/vulgar conversational pronouns ('tao', 'mày') are STRICTLY RESTRICTED to direct spoken dialogue between close friends.\n" +
			"       - ABSOLUTELY NEVER let 'tao', 'mày' leak into 1st-person POV narration or internal monologues (narration must strictly use 'tôi' or 'mình').\n" +
			"     * DIRECT DIALOGUE ADDRESS CONSISTENCY:\n" +
			"       - In direct 1-on-1 dialogue, once an established address pair is used (e.g. 'anh - em' between protagonist and friend's younger sister / kouhai), maintain it consistently across the entire conversation; do NOT arbitrarily switch to 'tôi - em' mid-scene.\n" +
			"   - ZERO CHATBOT META-TALK:\n" +
			"     * ABSOLUTELY NEVER output conversational chatter, meta-commentary, or chatbot preamble/postamble (e.g. 'Dưới đây là bản dịch:', 'Here is the translation:', 'Sure!').\n" +
			"     * When encountering a chapter or segment containing only images or markup (cover, illustration, <img>, <svg>), output ONLY the preserved HTML/SVG markup without any commentary!"

	case "English":
		return fmt.Sprintf("CRITICAL SCRIPT, CAPITALIZATION, TYPOGRAPHY & LITERARY FLUENCY RULES FOR ENGLISH:\n"+
			"1. SCRIPT & TRANSLITERATION:\n"+
			"   - Output MUST be 100%% in natural, publishable English (Latin alphabet). Zero untranslated Kanji, Hanzi, Kana, or Hangul characters.\n"+
			"   - Proper Names: Transcribe Japanese names to standard Romaji (Hepburn), Chinese names to standard Pinyin, Korean names to Revised Romanization according to standard transliteration or the injected glossary.\n\n"+
			"2. CAPITALIZATION & GRAMMAR:\n"+
			"   - Follow standard English publishing orthography. Capitalize proper nouns, official titles, and sentence beginnings.\n"+
			"   - Do NOT arbitrarily capitalize common nouns, cultivation ranks, general roles, or pronouns mid-sentence unless designated as proper names in the world bible.\n\n"+
			"3. PUNCTUATION & TYPOGRAPHY:\n"+
			"   - Sentences MUST terminate with '.', '!', '?', or '...'. NEVER terminate sentences or paragraphs with a semicolon (;) or comma (,).\n"+
			"   - Convert all CJK punctuation marks ('。', '，', '、', '；', '「」', '『』', '《》', '【】') into standard English typography ('\"...\"', italics for book/technique titles, brackets '[...]' for system messages).\n"+
			"   - Use standard typographic quotation marks (\"...\") and clean em-dashes (—) without extraneous spaces.\n\n"+
			"4. LITERARY FLUENCY, IDIOMATIC EXPRESSION & ANTI-TRANSLATIONESE (STRICTLY NO WORD-BY-WORD):\n"+
			"   - DE-CALQUING & SYNTACTIC NATURALIZATION: Do NOT mirror source East Asian topic-prominent, passive, or repetitive clause structures.\n"+
			"     * Dissolve repetitive passive/causative constructions: (e.g. JA '怒られた' -> 'she scolded me' / 'I caught an earful', NOT stiff 'I was gotten angry at').\n"+
			"     * Eliminate sensory filter bloat: Prefer strong active verbs ('The door creaked open' instead of 'He felt the sight of the door opening').\n"+
			"     * Resolve Asian onomatopoeia/sound symbolism (擬音語・擬態語) into vivid descriptive verbs or natural English sounds (e.g. 'doki-doki' -> 'her heart fluttered / pounded in her chest', NOT literal 'doki-doki'; 'niko-niko' -> 'beaming / grinning from ear to ear').\n"+
			"   - IDIOMS & FIGURATIVE TRANSLATION: Translate idioms by their contextual meaning, never word-by-word.\n"+
			"     * Chinese idioms (成语) & martial tropes: Render vividly (e.g. '自寻死路' -> 'you have a death wish' / 'courting disaster', NOT wooden literalism; '扮猪吃老虎' -> 'playing the fool to catch the tiger' / 'hiding one's true strength').\n"+
			"     * Japanese colloquial expressions: Render into natural modern English vernacular without losing comedic or emotional nuance.\n"+
			"   - AUTHENTIC LIVING DIALOGUE: Characters must speak natural contemporary English with proper contractions (don't, can't, it's, I'm), witty banter, and distinct character voices (e.g. tsundere, deadpan, formal butler, energetic junior).")

	case "Chinese":
		return fmt.Sprintf("CRITICAL SCRIPT, TYPOGRAPHY & LITERARY FLUENCY RULES FOR CHINESE (中文文学出版与网文翻译标准):\n"+
			"1. SCRIPT & TRANSLITERATION:\n"+
			"   - Output MUST be 100%% in natural, fluent Simplified Chinese characters (简体中文) without raw Latin, Kana, or Hangul.\n"+
			"   - Foreign/Western proper names from '%s' must be translated or transcribed into standard Chinese equivalents (e.g. Alice -> 爱丽丝).\n\n"+
			"2. PUNCTUATION & PUBLISHING TYPOGRAPHY:\n"+
			"   - Follow standard Chinese publishing punctuation: full stops '。', commas '，', enumeration commas '、', dialogue quotation marks '“……”', book/technique titles '《……》', system prompts '【……】'.\n"+
			"   - Strictly FORBIDDEN: NEVER terminate narrative sentences with English semicolons (;) or commas (,).\n"+
			"   - Ellipsis: Use standard Chinese six-dot ellipsis '……' (two characters). NEVER output English '...' or irregular dots.\n\n"+
			"3. LITERARY FLUENCY, ANTI-TRANSLATIONESE & DE-CALQUING (彻底杜绝欧化句式与生硬直译):\n"+
			"   - 彻底摒弃欧化句式 (De-Europeanization): 严禁生硬套用印欧语系的长定语从句和介词短语。\n"+
			"     * 杜绝滥用“被”字句: 中文“被”字带有消极色彩。将英文被动语态（如 'he was welcomed by...' / 'I was greeted by...'）转换为主动句或自然描述句（如“一睁开眼，映入眼帘的便是……”或“众人纷纷热情相迎”，严禁机械翻译为“我被……欢迎”）。\n"+
			"     * 杜绝多重“的”字堆叠: 严禁出现“一个穿着漂亮的衣服的年轻的女孩”式的翻译腔；提炼为四字成语或精炼主谓结构（如“身着华服的妙龄少女”）。\n"+
			"     * 杜绝机械化代词泛滥: 英文频繁使用 'it/they'，在中文语境中应自然省略指代或换用具体名词，严禁满篇“它/它们/通过它”。\n"+
			"   - 成语与四字格律化 (Idiomatic & Rhythmic Prose): 善用精练的中文四字词与成语（如“目瞪口呆”、“心惊胆战”、“波澜不惊”、“风驰电掣”），使文风具备网文与轻小说的节奏感与审美意境。\n"+
			"   - 生动活泼的角色对话 (Authentic Dialogue): 对话须符合人设口吻（如傲娇、毒舌、天然呆、老者威严），恰当运用语气助词（“啦”、“嘛”、“呀”、“呢”、“才不是……呢”），严禁出现机器人式的生硬对话。", srcNorm)

	case "Japanese":
		return fmt.Sprintf("CRITICAL SCRIPT, TYPOGRAPHY & LITERARY FLUENCY RULES FOR JAPANESE (文芸・ライトノベル出版基準):\n"+
			"1. SCRIPT & TRANSLITERATION:\n"+
			"   - Output MUST be 100%% in natural, fluent Japanese prose with balanced Kanji, Hiragana, and Katakana.\n"+
			"   - Foreign/Western names and loanwords from '%s' must be transcribed into standard Katakana (e.g. Alice -> アリス) or according to standard transliteration conventions and the injected glossary.\n\n"+
			"2. PUBLISHING PUNCTUATION & TYPOGRAPHY:\n"+
			"   - Follow standard Japanese publishing punctuation: sentences end with '。', dialogue enclosed in '「……」', thoughts or sub-quotes in '『……』'.\n"+
			"   - Ellipsis: Use standard Japanese double 3-dots '……' (two glyphs). NEVER use raw English '...' or semicolons (;).\n"+
			"   - Exclamation & Question marks: Follow publishing spacing (leave a full-width space '！　' or '？　' after marks mid-sentence unless closed by quote).\n\n"+
			"3. LITERARY FLUENCY, ANTI-TRANSLATIONESE & DE-CALQUING (脱・直訳調・翻訳調の徹底):\n"+
			"   - 主語の自然な省略 (Natural Subject Dropping): 英語の 'I', 'he', 'she', 'they' を直訳して「彼」「彼女」「私」を毎文繰り返すのは厳禁。文脈から自明な主語は自然に省略すること。\n"+
			"   - 役割語とキャラクター口調の徹底 (Distinct Role Language & Speech Styles): キャラクターの年齢・性別・立場に応じた生きた口調・語尾を厳格に再現すること。\n"+
			"     * 主人公（少年/青年）: 「〜だよ」「〜だろ」「〜じゃねえか」「〜ってわけか」\n"+
			"     * 少女/妹キャラ: 「〜だよ！」「〜なの？」「〜してよね！」\n"+
			"     * ツンデレ: 「〜じゃない！」「別に〜なんかじゃないんだからね！」\n"+
			"     * 威厳ある師匠/老人: 「〜じゃ」「〜のう」「わしの……」\n"+
			"     * 敬語/丁寧語: 「〜です」「〜ます」「〜でございます」\n"+
			"     * 全キャラクターを一律の「です・ます」で無機質に訳すことは厳禁。\n"+
			"   - 慣用句・比喩の自然な意訳 (Figurative Idiom Rendering):\n"+
			"     * 英語 'in her birthday suit' -> 「一糸まとわぬ姿」「生まれたままの姿」（絶対に「誕生日の服」などと誤訳しないこと）。\n"+
			"     * 英語 'triumphant smile' -> 「勝ち誇った笑み」「ドヤ顔」。\n"+
			"     * 英語 'for a second there' -> 「一瞬、〜かと思った」。\n"+
			"   - テンポとリズム (Cadence & Pacing): 日本語ライトノベル特有の軽快なテンポ、心情描写の余韻を大切にし、生きた日本語として読める文章に仕上げること。", srcNorm)

	case "Korean":
		return fmt.Sprintf("CRITICAL SCRIPT, TYPOGRAPHY & LITERARY FLUENCY RULES FOR KOREAN (웹소설・출판 번역 기준):\n"+
			"1. SCRIPT & TRANSLITERATION:\n"+
			"   - Output MUST be 100%% in natural Korean prose using Hangul.\n"+
			"   - Foreign/Western names and terms from '%s' must be transcribed using standard National Institute of Korean Language (국립국어원) loanword orthography (e.g. Alice -> 앨리스) or according to standard transliteration conventions and the injected glossary.\n\n"+
			"2. PUNCTUATION & TYPOGRAPHY:\n"+
			"   - Follow standard Korean spacing (띄어쓰기) and publishing punctuation.\n"+
			"   - Dialogue enclosed in '“……”', thoughts in '‘……’'. Ellipsis uses standard '……' (6 dots). Never terminate sentences with semicolons (;).\n\n"+
			"3. LITERARY FLUENCY, ANTI-TRANSLATIONESE & DE-CALQUING (번역투 문장 완전 척결):\n"+
			"   - 번역투 표현 지양: 외국어 문법의 직역투를 완벽히 배제할 것.\n"+
			"     * 불필요한 대명사 제거: 영어의 'he/she/it', 일본어의 '彼/彼女'를 '그', '그녀', '그것'으로 매번 직역하지 말고 생략하거나 인물명/호칭으로 자연스럽게 전환.\n"+
			"     * 이중 피동 및 어색한 피동 지양: '-되어지다', '-하여지다', '-에 의해' 남발을 피하고 능동형('밝혀졌다', '풀었다', '맞이했다')으로 자연스럽게 재구성.\n"+
			"     * 일본어투/외래어투 조사 표현 지양: '-에 있어서', '-에 관하여', '-적(的)' 대신 자연스러운 한국어 조사 및 형용사 사용.\n"+
			"   - 관용구 및 비유의 생생한 의역: 외국어 관용구를 단어 그대로 번역하지 말고 한국어의 등가 표현으로 자연스럽게 치환.\n"+
			"     * 'in her birthday suit' -> '실오라기 하나 걸치지 않은 알몸으로' / '생일 정장(X)'.\n"+
			"     * 'triumphant smile' -> '득의양양한 미소' / '의기양양한 표정'.\n"+
			"   - 생동감 넘치는 캐릭터 대사: 하십시오체, 해요체, 해체, 해라체를 캐릭터 성격과 상호 관계에 맞게 철저히 구사하고 어미(-거든, -잖아, -단 말이야)를 살려 자연스러운 호흡을 완성할 것.", srcNorm)

	case "French":
		return fmt.Sprintf("CRITICAL SCRIPT, TYPOGRAPHY & LITERARY FLUENCY RULES FOR FRENCH (NORMES D'ÉDITION LITTÉRAIRE):\n"+
			"1. SCRIPT & ORTHOGRAPHY: 100%% natural literary French with correct diacritics (é, è, ê, à, ç, etc.).\n"+
			"2. PUNCTUATION & TYPOGRAPHY: Dialogue uses typographic guillemets « ... » with non-breaking spaces or em-dash (—) dialogue styling. Sentences end with '.', '!', '?', or '...'. NEVER end with a semicolon (;).\n"+
			"3. ANTI-TRANSLATIONESE & LITERARY VOICE:\n"+
			"   - De-calque foreign syntax; employ appropriate literary narrative tenses (Passé simple / Imparfait for narration, Présent for dialogue).\n"+
			"   - Render foreign idioms and metaphors by their natural French cultural equivalents (e.g. 'in her birthday suit' -> 'dans le plus simple appareil' / 'complètement nue', NEVER 'costume d'anniversaire').\n"+
			"   - Dialogue must reflect natural spoken French flow with appropriate colloquial register and distinct character personalities.")

	case "German":
		return fmt.Sprintf("CRITICAL SCRIPT, TYPOGRAPHY & LITERARY FLUENCY RULES FOR GERMAN (LITERATUR- UND VERLAGSNORMEN):\n"+
			"1. SCRIPT & ORTHOGRAPHY: 100%% natural literary German with correct umlauts (ä, ö, ü, ß) and noun capitalization.\n"+
			"2. PUNCTUATION & TYPOGRAPHY: Use standard German quotation marks („...“) or guillemets (»...«). Sentences end with '.', '!', '?', or '...'. NEVER end with a semicolon (;).\n"+
			"3. ANTI-TRANSLATIONESE & LITERARY VOICE:\n"+
			"   - De-calque passive and foreign relative clauses into natural German literary syntax (Präteritum for narration).\n"+
			"   - Render foreign idioms by natural German equivalents (e.g. 'in her birthday suit' -> 'im Evaskostüm' / 'splitterfasernackt', NEVER 'Geburtstagsanzug').\n"+
			"   - Dialogue must sound lively, natural, and character-authentic.")

	case "Spanish":
		return fmt.Sprintf("CRITICAL SCRIPT, TYPOGRAPHY & LITERARY FLUENCY RULES FOR SPANISH (NORMAS DE EDICIÓN LITERARIA):\n"+
			"1. SCRIPT & ORTHOGRAPHY: 100%% natural literary Spanish with correct accents (á, é, í, ó, ú, ñ).\n"+
			"2. PUNCTUATION & TYPOGRAPHY: Opening inverted marks (¿?, ¡!) and em-dash dialog (—) or quotes («...»). Sentences end with '.', '!', '?', or '...'. NEVER end with a semicolon (;).\n"+
			"3. ANTI-TRANSLATIONESE & LITERARY VOICE:\n"+
			"   - De-calque foreign syntax into natural Spanish cadence (Pretérito indefinido / imperfecto for narrative).\n"+
			"   - Translate idioms by cultural equivalents (e.g. 'in her birthday suit' -> 'en traje de Eva' / 'completamente desnuda', NEVER 'traje de cumpleaños').\n"+
			"   - Dialogue must sound fluid, expressive, and authentically human.")

	default:
		return fmt.Sprintf("CRITICAL SCRIPT, TYPOGRAPHY & LITERARY FLUENCY RULES FOR %s:\n"+
			"1. SCRIPT & INTEGRITY: Output MUST be 100%% in natural, fluent %s. Zero untranslated source characters or meta commentary.\n"+
			"2. PUNCTUATION: Sentences MUST terminate with '.', '!', '?', or '...'. NEVER terminate sentences or paragraphs with a semicolon (;) or comma (,).\n"+
			"3. ANTI-TRANSLATIONESE & DE-CALQUING: Never translate word-by-word. Deconstruct foreign grammatical structures and reconstruct them into native, idiomatic %s prose. Convert passive voice and nominal abstractions into natural active descriptions.\n"+
			"4. FIGURATIVE SENSE & IDIOMS: Render idioms, metaphors, and slang by their true cultural and figurative equivalents in %s, never by literal word substitution.\n"+
			"5. AUTHENTIC LIVING DIALOGUE: Ensure all character speech reflects natural spoken cadence, distinct personalities, and proper relational hierarchy.", strings.ToUpper(tgtNorm), tgtNorm, tgtNorm, tgtNorm)
	}
}
