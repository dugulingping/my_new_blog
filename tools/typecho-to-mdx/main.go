package main

// Usage:
//   cd tools/typecho-to-mdx
//   go run . -password 'your-password' -output ../../tmp/typecho-import
//
// Typical production run:
//   go run . -host 192.168.58.1 -port 3306 -user root -password 'mysql_QbESS5' \
//     -database blog -prefix typecho -output ../../src/content/blog/typecho-import

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	_ "github.com/go-sql-driver/mysql"
)

type config struct {
	host      string
	port      int
	user      string
	password  string
	database  string
	prefix    string
	outputDir string
	status    string
	postType  string
	limit     int
	overwrite bool
}

type post struct {
	CID        int64
	Title      string
	Slug       string
	Body       string
	Created    int64
	Modified   int64
	Categories []string
	Tags       []string
}

var (
	validPrefixPattern = regexp.MustCompile(`^[A-Za-z0-9_]+$`)
	typechoMarkerRegex = regexp.MustCompile(`(?i)<!--\s*(markdown|html)\s*-->`)
	metingBlockRegex   = regexp.MustCompile(`(?is)\[Meting\].*?\[/Meting\]`)
	wrappedHeadingRegex = regexp.MustCompile(`(?m)^(#{1,6}\s+.*?)(?:\s*#+\s*)$`)
	htmlTagRegex       = regexp.MustCompile(`</?[^>]+>`)
	spaceRegex         = regexp.MustCompile(`\s+`)
)

func main() {
	cfg := config{}
	flag.StringVar(&cfg.host, "host", "192.168.58.1", "MySQL host")
	flag.IntVar(&cfg.port, "port", 3306, "MySQL port")
	flag.StringVar(&cfg.user, "user", "root", "MySQL user")
	flag.StringVar(&cfg.password, "password", "", "MySQL password")
	flag.StringVar(&cfg.database, "database", "blog", "database name")
	flag.StringVar(&cfg.prefix, "prefix", "typecho", "Typecho table prefix, with or without trailing underscore")
	flag.StringVar(&cfg.outputDir, "output", filepath.Join("src", "content", "blog", "typecho-import"), "output directory")
	flag.StringVar(&cfg.status, "status", "publish", "Typecho status to export")
	flag.StringVar(&cfg.postType, "type", "post", "Typecho content type to export")
	flag.IntVar(&cfg.limit, "limit", 0, "maximum number of posts to export, 0 means all")
	flag.BoolVar(&cfg.overwrite, "overwrite", false, "overwrite existing files")
	flag.Parse()

	if cfg.password == "" {
		log.Fatal("missing required flag: -password")
	}

	cfg.prefix = normalizePrefix(cfg.prefix)
	if !validPrefixPattern.MatchString(cfg.prefix) {
		log.Fatalf("invalid prefix %q: only letters, digits and underscores are allowed", cfg.prefix)
	}

	db, err := openDatabase(cfg)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("ping database: %v", err)
	}

	posts, err := fetchPosts(db, cfg)
	if err != nil {
		log.Fatalf("fetch posts: %v", err)
	}

	if len(posts) == 0 {
		log.Printf("no posts matched type=%q status=%q", cfg.postType, cfg.status)
		return
	}

	if err := os.MkdirAll(cfg.outputDir, 0o755); err != nil {
		log.Fatalf("create output directory: %v", err)
	}

	exported, skipped, err := exportPosts(posts, cfg)
	if err != nil {
		log.Fatalf("export posts: %v", err)
	}

	log.Printf("done: exported=%d skipped=%d output=%s", exported, skipped, cfg.outputDir)
}

func normalizePrefix(prefix string) string {
	prefix = strings.TrimSpace(prefix)
	if prefix == "" {
		return prefix
	}
	if !strings.HasSuffix(prefix, "_") {
		return prefix + "_"
	}
	return prefix
}

func openDatabase(cfg config) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true&loc=Local",
		cfg.user,
		cfg.password,
		cfg.host,
		cfg.port,
		cfg.database,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	db.SetConnMaxLifetime(3 * time.Minute)
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(5)
	return db, nil
}

func fetchPosts(db *sql.DB, cfg config) ([]post, error) {
	contentsTable := cfg.prefix + "contents"
	query := fmt.Sprintf(`
SELECT cid, title, slug, text, created, modified
FROM %s
WHERE type = ? AND status = ? AND parent = 0
ORDER BY created ASC
`, contentsTable)

	var rows *sql.Rows
	var err error
	if cfg.limit > 0 {
		query += " LIMIT ?"
		rows, err = db.Query(query, cfg.postType, cfg.status, cfg.limit)
	} else {
		rows, err = db.Query(query, cfg.postType, cfg.status)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []post
	for rows.Next() {
		var p post
		if err := rows.Scan(&p.CID, &p.Title, &p.Slug, &p.Body, &p.Created, &p.Modified); err != nil {
			return nil, err
		}

		categories, tags, err := fetchTerms(db, cfg.prefix, p.CID)
		if err != nil {
			return nil, fmt.Errorf("fetch terms for cid=%d: %w", p.CID, err)
		}
		p.Categories = categories
		p.Tags = tags
		posts = append(posts, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}

func fetchTerms(db *sql.DB, prefix string, cid int64) ([]string, []string, error) {
	relationshipsTable := prefix + "relationships"
	metasTable := prefix + "metas"
	query := fmt.Sprintf(`
SELECT m.type, m.name, m.slug
FROM %s AS r
JOIN %s AS m ON m.mid = r.mid
WHERE r.cid = ?
ORDER BY m.type, m.order, m.mid
`, relationshipsTable, metasTable)

	rows, err := db.Query(query, cid)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var categories []string
	var tags []string

	for rows.Next() {
		var termType sql.NullString
		var name sql.NullString
		var slug sql.NullString
		if err := rows.Scan(&termType, &name, &slug); err != nil {
			return nil, nil, err
		}

		value := strings.TrimSpace(slug.String)
		if value == "" {
			value = strings.TrimSpace(name.String)
		}
		if value == "" {
			continue
		}

		switch termType.String {
		case "category":
			categories = append(categories, value)
		case "tag":
			tags = append(tags, value)
		}
	}

	return dedupe(categories), dedupe(tags), rows.Err()
}

func exportPosts(posts []post, cfg config) (int, int, error) {
	exported := 0
	skipped := 0
	usedNames := map[string]int{}

	for _, p := range posts {
		filename := buildFilename(p, usedNames)
		targetPath := filepath.Join(cfg.outputDir, filename+".mdx")

		if _, err := os.Stat(targetPath); err == nil && !cfg.overwrite {
			skipped++
			log.Printf("skip existing: cid=%d path=%s", p.CID, targetPath)
			continue
		}

		document, err := buildDocument(p)
		if err != nil {
			return exported, skipped, fmt.Errorf("build document for cid=%d: %w", p.CID, err)
		}

		if err := os.WriteFile(targetPath, []byte(document), 0o644); err != nil {
			return exported, skipped, fmt.Errorf("write %s: %w", targetPath, err)
		}

		exported++
		log.Printf("exported: cid=%d slug=%q path=%s", p.CID, p.Slug, targetPath)
	}

	return exported, skipped, nil
}

func buildFilename(p post, used map[string]int) string {
	base := sanitizeSlug(p.Slug)
	if isWeakFilename(base) {
		base = sanitizeSlug(p.Title)
	}
	if base == "" {
		base = fmt.Sprintf("post-%d", p.CID)
	}

	if count := used[base]; count > 0 {
		used[base] = count + 1
		return fmt.Sprintf("%s-%d", base, count+1)
	}
	used[base] = 1
	return base
}

func buildDocument(p post) (string, error) {
	body := normalizeBody(p.Body)
	description := buildDescription(body)
	allTags := mergeTags(p.Categories, p.Tags)

	pubDate := unixToTime(p.Created)
	modifiedDate := unixToTime(p.Modified)

	title := strings.TrimSpace(p.Title)
	pubDateText := formatFrontmatterDate(pubDate)
	updatedDateText := ""
	createdAtText := formatRawTimestamp(pubDate)
	updatedAtText := ""
	if !modifiedDate.IsZero() && modifiedDate.Unix() != pubDate.Unix() {
		updatedDateText = formatFrontmatterDate(modifiedDate)
		updatedAtText = formatRawTimestamp(modifiedDate)
	}

	var commentParts []string
	commentParts = append(commentParts, fmt.Sprintf("cid=%d", p.CID))
	if p.Slug != "" {
		commentParts = append(commentParts, fmt.Sprintf("slug=%s", p.Slug))
	}
	if len(p.Categories) > 0 {
		commentParts = append(commentParts, fmt.Sprintf("source-categories=%s", strings.Join(p.Categories, ",")))
	}
	if len(p.Tags) > 0 {
		commentParts = append(commentParts, fmt.Sprintf("source-tags=%s", strings.Join(p.Tags, ",")))
	}
	if len(allTags) > 0 {
		commentParts = append(commentParts, fmt.Sprintf("merged-tags=%s", strings.Join(allTags, ",")))
	}

	var builder strings.Builder
	builder.WriteString("---\n")
	builder.WriteString("title: ")
	builder.WriteString(quoteYAMLString(title))
	builder.WriteString("\n")
	builder.WriteString("description: ")
	builder.WriteString(quoteYAMLString(description))
	builder.WriteString("\n")
	builder.WriteString("pubDate: ")
	builder.WriteString(quoteYAMLString(pubDateText))
	builder.WriteString("\n")
	if createdAtText != "" {
		builder.WriteString("createdAt: ")
		builder.WriteString(quoteYAMLString(createdAtText))
		builder.WriteString("\n")
	}
	if len(allTags) > 0 {
		builder.WriteString("tags:\n")
		for _, tag := range allTags {
			builder.WriteString("  - ")
			builder.WriteString(quoteYAMLString(tag))
			builder.WriteString("\n")
		}
	}
	if updatedDateText != "" {
		builder.WriteString("updatedDate: ")
		builder.WriteString(quoteYAMLString(updatedDateText))
		builder.WriteString("\n")
	}
	if updatedAtText != "" {
		builder.WriteString("updatedAt: ")
		builder.WriteString(quoteYAMLString(updatedAtText))
		builder.WriteString("\n")
	}
	builder.WriteString("---\n\n")
	if len(commentParts) > 0 {
		builder.WriteString("{/* migrated from Typecho: ")
		builder.WriteString(strings.Join(commentParts, " | "))
		builder.WriteString(" */}\n\n")
	}
	builder.WriteString(strings.TrimSpace(body))
	builder.WriteString("\n")
	return builder.String(), nil
}

func quoteYAMLString(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `"`, `\"`)
	value = strings.ReplaceAll(value, "\n", "\\n")
	return `"` + value + `"`
}

func normalizeBody(body string) string {
	body = strings.ReplaceAll(body, "\r\n", "\n")
	body = strings.ReplaceAll(body, "\r", "\n")
	body = strings.TrimPrefix(body, "\uFEFF")
	body = typechoMarkerRegex.ReplaceAllString(body, "")
	body = metingBlockRegex.ReplaceAllString(body, "")
	body = strings.ReplaceAll(body, "<!--more-->", "\n\n")
	body = strings.ReplaceAll(body, "<!-- more -->", "\n\n")
	body = wrappedHeadingRegex.ReplaceAllString(body, "$1")
	body = strings.TrimSpace(body)
	return body
}

func buildDescription(body string) string {
	summarySource := body
	if idx := strings.Index(summarySource, "<!--more-->"); idx >= 0 {
		summarySource = summarySource[:idx]
	}

	summarySource = htmlTagRegex.ReplaceAllString(summarySource, " ")
	summarySource = strings.NewReplacer(
		"`", " ",
		"#", " ",
		"*", " ",
		"_", " ",
		">", " ",
		"[", " ",
		"]", " ",
		"(", " ",
		")", " ",
	).Replace(summarySource)
	summarySource = spaceRegex.ReplaceAllString(summarySource, " ")
	summarySource = strings.TrimSpace(summarySource)
	if summarySource == "" {
		return "从 Typecho 迁移的文章。"
	}

	return truncateRunes(summarySource, 120)
}

func truncateRunes(s string, limit int) string {
	if limit <= 0 || utf8.RuneCountInString(s) <= limit {
		return s
	}

	runes := []rune(s)
	return strings.TrimSpace(string(runes[:limit])) + "..."
}

func unixToTime(ts int64) time.Time {
	switch {
	case ts <= 0:
		return time.Time{}
	case ts > 1_000_000_000_000:
		return time.UnixMilli(ts).Local()
	default:
		return time.Unix(ts, 0).Local()
	}
}

func formatFrontmatterDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02")
}

func formatRawTimestamp(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
}

func sanitizeSlug(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" {
		return ""
	}

	var builder strings.Builder
	lastDash := false
	for _, r := range s {
		switch {
		case unicode.IsLetter(r), unicode.IsNumber(r):
			builder.WriteRune(r)
			lastDash = false
		case r == '-', r == '_':
			if !lastDash && builder.Len() > 0 {
				builder.WriteByte('-')
				lastDash = true
			}
		default:
			if !lastDash && builder.Len() > 0 {
				builder.WriteByte('-')
				lastDash = true
			}
		}
	}

	result := strings.Trim(builder.String(), "-")
	return result
}

func isWeakFilename(name string) bool {
	if name == "" {
		return true
	}

	for _, r := range name {
		if !unicode.IsNumber(r) {
			return false
		}
	}
	return true
}

func dedupe(values []string) []string {
	if len(values) == 0 {
		return values
	}

	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func mergeTags(groups ...[]string) []string {
	var merged []string
	for _, group := range groups {
		merged = append(merged, group...)
	}
	return dedupe(merged)
}
