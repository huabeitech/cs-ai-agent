package chunk

import (
	"crypto/sha256"
	"cs-agent/internal/pkg/enums"
	"encoding/hex"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	defaultTargetTokens  = 300
	defaultMaxTokens     = 400
	defaultOverlapTokens = 40
)

func normalizeOptions(opts ChunkOptions) ChunkOptions {
	if opts.TargetTokens <= 0 {
		opts.TargetTokens = defaultTargetTokens
	}
	if opts.MaxTokens <= 0 {
		opts.MaxTokens = defaultMaxTokens
	}
	if opts.MaxTokens < opts.TargetTokens {
		opts.MaxTokens = opts.TargetTokens
	}
	if opts.OverlapTokens < 0 {
		opts.OverlapTokens = 0
	}
	if opts.OverlapTokens == 0 {
		opts.OverlapTokens = defaultOverlapTokens
	}
	if opts.Provider == "" {
		opts.Provider = string(enums.KnowledgeChunkProviderStructured)
	}
	return opts
}

func normalizeText(text string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(text)), " ")
}

func estimateTokenCount(text string) int {
	text = strings.TrimSpace(text)
	if text == "" {
		return 0
	}
	count := 0
	inWord := false
	for _, r := range text {
		switch {
		case unicode.IsSpace(r):
			inWord = false
		case unicode.Is(unicode.Han, r):
			count++
			inWord = false
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			if !inWord {
				count++
				inWord = true
			}
		default:
			count++
			inWord = false
		}
	}
	if count == 0 {
		return utf8.RuneCountInString(text)
	}
	return count
}

func contentHash(text string) string {
	sum := sha256.Sum256([]byte(text))
	return hex.EncodeToString(sum[:])
}

func splitSentences(text string) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	var sentences []string
	var builder strings.Builder
	for _, r := range text {
		builder.WriteRune(r)
		switch r {
		case '\n', '。', '！', '？', '!', '?', ';', '；':
			sentence := normalizeText(builder.String())
			if sentence != "" {
				sentences = append(sentences, sentence)
			}
			builder.Reset()
		}
	}
	if builder.Len() > 0 {
		sentence := normalizeText(builder.String())
		if sentence != "" {
			sentences = append(sentences, sentence)
		}
	}
	if len(sentences) == 0 {
		return []string{normalizeText(text)}
	}
	return sentences
}

func tailTextByTokens(text string, tokenLimit int) string {
	if tokenLimit <= 0 {
		return ""
	}
	sentences := splitSentences(text)
	if len(sentences) == 0 {
		return ""
	}
	var selected []string
	total := 0
	for i := len(sentences) - 1; i >= 0; i-- {
		sentence := sentences[i]
		tokens := estimateTokenCount(sentence)
		if total > 0 && total+tokens > tokenLimit {
			break
		}
		selected = append([]string{sentence}, selected...)
		total += tokens
	}
	return strings.TrimSpace(strings.Join(selected, " "))
}

func splitPlainText(text string, opts ChunkOptions) []string {
	text = normalizeText(text)
	if text == "" {
		return nil
	}
	opts = normalizeOptions(opts)
	sentences := splitSentences(text)
	if len(sentences) == 0 {
		return nil
	}

	chunks := make([]string, 0)
	current := make([]string, 0)
	currentTokens := 0

	flush := func() {
		if len(current) == 0 {
			return
		}
		chunks = append(chunks, strings.Join(current, " "))
	}

	for _, sentence := range sentences {
		sentenceTokens := estimateTokenCount(sentence)
		if sentenceTokens > opts.MaxTokens {
			if len(current) > 0 {
				flush()
				overlap := tailTextByTokens(strings.Join(current, " "), opts.OverlapTokens)
				current = nil
				currentTokens = 0
				if overlap != "" {
					current = append(current, overlap)
					currentTokens = estimateTokenCount(overlap)
				}
			}
			for _, piece := range splitLongSentence(sentence, opts.MaxTokens) {
				piece = normalizeText(piece)
				if piece != "" {
					chunks = append(chunks, piece)
				}
			}
			continue
		}

		if currentTokens > 0 && currentTokens+sentenceTokens > opts.MaxTokens {
			flush()
			overlap := tailTextByTokens(strings.Join(current, " "), opts.OverlapTokens)
			current = nil
			currentTokens = 0
			if overlap != "" {
				current = append(current, overlap)
				currentTokens = estimateTokenCount(overlap)
			}
		}

		current = append(current, sentence)
		currentTokens += sentenceTokens
	}

	flush()
	return chunks
}

func splitLongSentence(text string, maxTokens int) []string {
	runes := []rune(strings.TrimSpace(text))
	if len(runes) == 0 {
		return nil
	}
	if maxTokens <= 0 {
		return []string{text}
	}
	window := maxTokens * 2
	if window < 50 {
		window = 50
	}
	var result []string
	for start := 0; start < len(runes); start += window {
		end := start + window
		if end > len(runes) {
			end = len(runes)
		}
		part := normalizeText(string(runes[start:end]))
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}
