package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"vex-backend/config"
	"vex-backend/git"
	vectormgr "vex-backend/vector/manager"
)

type WebhookPayload struct {
	RepoURL string `json:"repo_url"`
}

// isOnlyWikiLinks returns true when the content (after removing frontmatter,
// comments and common link syntaxes) contains no letters or digits — i.e. only
// wiki links and punctuation/whitespace remain.
func isOnlyWikiLinks(content string) bool {
	// Remove YAML frontmatter: --- ... --- at start of file
	reFront := regexp.MustCompile(`(?s)\A---.*?---\s*`)
	content = reFront.ReplaceAllString(content, "")

	// Remove HTML comments
	reComments := regexp.MustCompile(`(?s)<!--.*?-->`)
	content = reComments.ReplaceAllString(content, "")

	// Remove markdown inline links like [text](url)
	reMDLinks := regexp.MustCompile(`\[[^\]]+\]\([^)]+\)`)
	content = reMDLinks.ReplaceAllString(content, "")

	// Remove wiki links [[...]]
	reWiki := regexp.MustCompile(`\[\[[^\]]+\]\]`)
	content = reWiki.ReplaceAllString(content, "")

	// If anything letter/number remains, it's not only links.
	reAlphaNum := regexp.MustCompile(`\p{L}|\p{N}`)
	return !reAlphaNum.MatchString(content)
}

// GitWebhookHandler returns an http.HandlerFunc that pulls the repo, deletes any existing
// vectors for markdown files and re-embeds them. It uses the provided Manager instance.
func GitWebhookHandler(m vectormgr.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("[GitWebhook] invoked at %v from %s", start, r.RemoteAddr)

		// Ensure repo is up to date (clone or pull)
		repo := config.Config.NotesRepo
		log.Printf("[GitWebhook] ensuring notes repo is up-to-date: %s", repo)
		files, err := git.GetChangedFiles(repo)
		if err != nil {
			log.Printf("[GitWebhook] git.GetFiles error: %v", err)
			http.Error(w, "git error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		log.Printf("[GitWebhook] found %d changed files", len(files))

		// If no files changed, return early
		if len(files) == 0 {
			duration := time.Since(start)
			resp := map[string]any{
				"status":          "success",
				"processed_count": 0,
				"skipped_count":   0,
				"processed":       []string{},
				"skipped":         []string{},
				"duration_ms":     duration.Milliseconds(),
				"message":         "no files changed",
			}

			respBytes, err := json.Marshal(resp)
			if err != nil {
				log.Printf("[GitWebhook] failed to marshal response: %v", err)
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}

			log.Printf("[GitWebhook] completed: no changes detected, duration=%s", duration)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(respBytes)
			return
		}

		basePath := filepath.Join(config.Config.CloneFolder, filepath.Base(repo))

		processed := make([]string, 0, len(files))
		skipped := make([]string, 0, len(files))

		// Process only changed markdown files:
		// delete any existing vectors for the file (by metadata) then re-embed it.
		for _, rel := range files {
			// only process markdown files
			if strings.ToLower(filepath.Ext(rel)) != ".md" {
				skipped = append(skipped, rel)
				log.Printf("[GitWebhook] skipping non-markdown file: %s", rel)
				continue
			}

			fullpath := filepath.Join(basePath, rel)
			log.Printf("[GitWebhook] processing markdown file: %s", fullpath)

			// Try to read the file to decide whether to embed
			data, err := os.ReadFile(fullpath)
			if err != nil {
				// If we can't read it, log and skip (don't fail the whole webhook).
				log.Printf("[GitWebhook] warning: failed to read %s: %v", fullpath, err)
				skipped = append(skipped, rel)
				continue
			}
			content := string(data)

			// If the file contains only wiki-links (like [[a]] [[b]]), skip embedding.
			if isOnlyWikiLinks(content) {
				// Optionally delete existing vectors for this file so stale embeddings are removed.
				if err := m.DeleteVectorsWithMetaData(r.Context(), "filepath", fullpath); err != nil {
					log.Printf("[GitWebhook] warning: failed to delete existing vectors for %s: %v", fullpath, err)
				} else {
					log.Printf("[GitWebhook] deleted existing vectors for %s (file is link-only)", fullpath)
				}

				skipped = append(skipped, rel)
				log.Printf("[GitWebhook] skipping link-only file: %s", rel)
				continue
			}

			// delete any existing vectors that have metadata filepath = fullpath
			if err := m.DeleteVectorsWithMetaData(r.Context(), "filepath", fullpath); err != nil {
				// don't fail the whole webhook on delete errors; log and continue
				log.Printf("[GitWebhook] warning: failed to delete existing vectors for %s: %v", fullpath, err)
			} else {
				log.Printf("[GitWebhook] deleted existing vectors for %s", fullpath)
			}

			// store (embed) the file into the vector DB
			if err := m.StoreFileAsVectorsInDB(r.Context(), fullpath); err != nil {
				log.Printf("[GitWebhook] failed to store vectors for %s: %v", fullpath, err)
				http.Error(w, "embed error: "+err.Error(), http.StatusInternalServerError)
				return
			}
			log.Printf("[GitWebhook] embedded %s", fullpath)
			processed = append(processed, rel)
		}

		duration := time.Since(start)
		resp := map[string]any{
			"status":          "success",
			"processed_count": len(processed),
			"skipped_count":   len(skipped),
			"processed":       processed,
			"skipped":         skipped,
			"duration_ms":     duration.Milliseconds(),
		}

		respBytes, err := json.Marshal(resp)
		if err != nil {
			log.Printf("[GitWebhook] failed to marshal response: %v", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		log.Printf("[GitWebhook] completed: processed=%d skipped=%d duration=%s", len(processed), len(skipped), duration)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(respBytes)
	}
}
