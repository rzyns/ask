package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/yeasy/ask/internal/cache"
	"github.com/yeasy/ask/internal/config"
	"github.com/yeasy/ask/internal/git"
	"github.com/yeasy/ask/internal/github"
	"github.com/yeasy/ask/internal/repository"
	"github.com/yeasy/ask/internal/skill"
	"github.com/yeasy/ask/internal/skillhub"
)

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:     "install [url...]",
	Aliases: []string{"add", "i"},
	Short:   "Install one or more skills from git repositories",
	Long: `Download and install skills into agent-specific directories. 
You can provide full git URLs or GitHub shorthands (owner/repo).
You can also specify versions: owner/repo@v1.0.0

Use --agent (-a) to specify target agents (claude, cursor, codex, opencode).
Multiple agents can be specified by repeating the flag.
If no agent is specified, skills are installed to .agent/skills/ by default.`,
	Example: `  # Install from GitHub shorthand
  ask skill install browser-use/browser-use
  
  # Install to specific agents
  ask skill install mcp-builder --agent claude --agent cursor
  ask skill install mcp-builder -a claude -a cursor
  
  # Install globally for an agent
  ask skill install mcp-builder --agent claude --global
  
  # Install multiple skills at once
  ask skill install browser-use web-surfer mcp-server
  
  # Install specific version
  ask skill install anthropics/skills@v1.2.0
  
  # Install from subdirectory
  ask skill install anthropics/skills/skills/browser-use
  
  # Install from GitHub browser URL
  ask skill install https://github.com/anthropics/skills/tree/main/skills/mcp-builder
  
  # Install from full URL
  ask skill install https://github.com/browser-use/browser-use.git`,
	Args: cobra.MinimumNArgs(1),
	Run:  runInstall,
}

func runInstall(cmd *cobra.Command, args []string) {
	// Check for offline mode
	if offline, _ := cmd.Flags().GetBool("offline"); offline || github.OfflineMode {
		fmt.Println("Error: Cannot install skills in offline mode.")
		os.Exit(1)
	}

	// Check for global flag
	global, _ := cmd.Flags().GetBool("global")

	// Get agent targets
	agents, _ := cmd.Flags().GetStringSlice("agent")

	// Validate agent names
	for _, agent := range agents {
		if !config.IsValidAgent(agent) {
			fmt.Printf("Error: Unknown agent '%s'. Supported agents: %s\n",
				agent, strings.Join(config.GetSupportedAgentNames(), ", "))
			os.Exit(1)
		}
	}

	// Ensure project is initialized for non-global, non-agent-specific operations
	if !global && len(agents) == 0 {
		if !ensureInitialized() {
			return
		}
	}

	// Track installation results
	var succeeded, failed []string

	// Install each skill
	// Check for repo aliases and expand them
	// Load config to check for repos
	cfg, err := config.LoadConfig()
	if err != nil {
		// Ignore error, might not be initialized
		def := config.DefaultConfig()
		cfg = &def
	}

	var expandedArgs []string
	for _, input := range args {
		// Check if input matches a configured repository name
		var targetRepo *config.Repo
		for i := range cfg.Repos {
			r := &cfg.Repos[i]
			// Debug print to diagnose matching issues (will remove later)
			// fmt.Printf("DEBUG: Checking input '%s' against repo '%s' (URL: %s)\n", input, r.Name, r.URL)

			// Match by name
			if r.Name == input {
				targetRepo = r
				break
			}
			// Match by owner/repo shorthand from URL
			if strings.Contains(r.URL, input) {
				if strings.HasPrefix(r.URL, input) || strings.Contains(r.URL, "/"+input) {
					targetRepo = r
					break
				}
			}
		}

		if targetRepo != nil {
			fmt.Printf("Fetching skills from repo '%s'...\n", input)

			var repos []github.Repository
			var err error

			// Try git-based discovery first for 'dir' type repos (avoids API rate limits)
			if targetRepo.Type == "dir" {
				repos, err = repository.FetchSkillsViaGit(*targetRepo)
			}

			// If git discovery failed or wasn't applicable, fall back to API
			if err != nil || targetRepo.Type != "dir" {
				repos, err = repository.FetchSkills(*targetRepo)
				if err != nil {
					fmt.Printf("Failed to fetch skills from repo '%s': %v\n", input, err)
					failed = append(failed, input)
					continue
				}
			}

			if len(repos) == 0 {
				fmt.Printf("No skills found in repo '%s'\n", input)
				continue
			}

			fmt.Printf("Found %d skills in repo '%s'. Queueing for installation...\n", len(repos), input)
			for _, r := range repos {
				// Construct install URL for each skill
				// Ideally we should use the skill's specific path if possible
				// But for now, we can use the HTML URL or clone URL + path
				// The simplest way given current installSingleSkill logic might be passing the full browser URL
				expandedArgs = append(expandedArgs, r.HTMLURL)
			}
		} else {
			expandedArgs = append(expandedArgs, input)
		}
	}

	// Install each expanded skill
	for _, input := range expandedArgs {
		err := installSingleSkill(input, global, agents, cfg)
		if err != nil {
			failed = append(failed, input)
			fmt.Printf("Failed to install %s: %v\n", input, err)
		} else {
			succeeded = append(succeeded, input)
		}
	}

	// Print summary if multiple skills were requested
	if len(args) > 1 {
		fmt.Println()
		fmt.Println("Installation Summary:")
		if len(succeeded) > 0 {
			fmt.Printf("  ✓ Succeeded: %d (%s)\n", len(succeeded), strings.Join(succeeded, ", "))
		}
		if len(failed) > 0 {
			fmt.Printf("  ✗ Failed: %d (%s)\n", len(failed), strings.Join(failed, ", "))
		}
	}

	// Exit with error if any installation failed
	if len(failed) > 0 {
		os.Exit(1)
	}
}

// parseGitHubBrowserURL parses a GitHub browser URL and extracts components
// Input: https://github.com/owner/repo/tree/branch/path/to/skill
// Returns: repoURL, branch, subDir, skillName, ok
func parseGitHubBrowserURL(url string) (repoURL, branch, subDir, skillName string, ok bool) {
	// Remove trailing slashes
	url = strings.TrimSuffix(url, "/")

	// Check if it contains /tree/ (GitHub browser URL format)
	if !strings.Contains(url, "/tree/") {
		return "", "", "", "", false
	}

	// Pattern: https://github.com/owner/repo/tree/branch/path
	parts := strings.SplitN(url, "/tree/", 2)
	if len(parts) != 2 {
		return "", "", "", "", false
	}

	repoURL = parts[0] + ".git"

	// Split branch and path
	branchAndPath := parts[1]
	pathParts := strings.SplitN(branchAndPath, "/", 2)
	branch = pathParts[0]

	if len(pathParts) > 1 {
		subDir = pathParts[1]
		// Skill name is the last component of the path
		skillName = filepath.Base(subDir)
	} else {
		// No subdir, use repo name from URL
		urlParts := strings.Split(parts[0], "/")
		skillName = urlParts[len(urlParts)-1]
	}

	return repoURL, branch, subDir, skillName, true
}

// installSingleSkill installs a single skill and returns an error if it fails
func installSingleSkill(input string, global bool, agents []string, cfg *config.Config) error {
	// Parse version if specified (skill@version)
	var version string
	originalInput := input
	if idx := strings.LastIndex(input, "@"); idx != -1 && !strings.HasPrefix(input, "git@") {
		version = input[idx+1:]
		input = input[:idx]
	}

	var repoURL, subDir, skillName, branch, localSourcePath string

	// First, check if it's a GitHub browser URL with /tree/
	if parsedURL, parsedBranch, parsedSubDir, parsedName, ok := parseGitHubBrowserURL(input); ok {
		repoURL = parsedURL
		branch = parsedBranch
		subDir = parsedSubDir
		skillName = parsedName
	} else {
		// Check if it's a direct URL or shorthand
		isURL := strings.HasPrefix(input, "http") || strings.HasPrefix(input, "git@")

		if !isURL {
			parts := strings.Split(input, "/")
			if len(parts) > 2 {
				// It's a subdirectory install: owner/repo/path/to/skill
				owner := parts[0]
				repo := parts[1]
				subDir = strings.Join(parts[2:], "/")
				skillName = parts[len(parts)-1]
				repoURL = fmt.Sprintf("https://github.com/%s/%s.git", owner, repo)
			} else {
				// Potential RepoName/SkillName from cache (e.g. "anthropics-skills/browser-use")
				// or Standard install: owner/repo (e.g. "browser-use/browser-use")

				foundInCache := false
				repoName := parts[0]
				skillNamePart := parts[1]

				reposCache, err := cache.NewReposCache()
				if err == nil && reposCache.HasRepo(repoName) {
					// Repo exists in cache, check if skill exists in it
					skills, err := reposCache.ListSkills(repoName)
					if err == nil {
						for _, s := range skills {
							if s.Name == skillNamePart {
								fmt.Printf("Found skill '%s' in cached repo '%s'\n", skillNamePart, repoName)

								// Resolve URL and subDir
								repoInfos, err := reposCache.LoadIndex()
								if err == nil {
									for _, info := range repoInfos {
										if info.Name == s.RepoName {
											repoURL = info.URL

											// If URL is missing in index (bug fix), lookup from config
											if repoURL == "" && cfg != nil {
												for _, r := range cfg.Repos {
													// Calculate derived name as used in sync
													derivedName := r.Name
													if !strings.HasPrefix(r.URL, "http") {
														parts := strings.Split(r.URL, "/")
														if len(parts) >= 2 {
															derivedName = parts[0] + "-" + parts[1]
														}
													} else {
														derivedName = strings.ReplaceAll(r.URL, "/", "-")
													}

													if r.Name == s.RepoName || derivedName == s.RepoName {
														// Construct URL from repo config
														// r.URL might be "anthropics/skills/skills"
														// We need to convert it to git URL
														if !strings.HasPrefix(r.URL, "http") {
															parts := strings.Split(r.URL, "/")
															if len(parts) >= 2 {
																repoURL = fmt.Sprintf("https://github.com/%s/%s.git", parts[0], parts[1])
															}
														} else {
															repoURL = r.URL
														}
														break
													}
												}
											}

											localSourcePath = s.Path
											rel, err := filepath.Rel(info.LocalPath, s.Path)
											if err == nil && rel != "." {
												subDir = rel
											}
											skillName = s.Name
											foundInCache = true
											break
										}
									}
								}
								if foundInCache {
									break
								}
							}
						}
					}
				}

				if !foundInCache {
					// Fallback: Check if it's a known repo in config (e.g. anthropics/skill-creator)
					// This handles the case where cache is missing/deleted but input format matches config
					configMatch := false
					if cfg != nil {
						for _, r := range cfg.Repos {
							if r.Name == parts[0] {
								// Found matching repo config
								if !strings.HasPrefix(r.URL, "http") {
									repoParts := strings.Split(r.URL, "/")
									if len(repoParts) >= 2 {
										repoURL = fmt.Sprintf("https://github.com/%s/%s.git", repoParts[0], repoParts[1])
									}
								} else {
									repoURL = r.URL
								}

								// Use the second part as skill name/subdir
								// Note: We don't know the exact subdir structure inside the repo if not cached.
								// However, most skill repos have skills at root or in /skills/.
								// 'ask' assumes sparse checkout of subDir.
								// Ideally we need the subDir path relative to repo root.
								// r.URL might be "anthropics/skills/skills" -> subDir base is "skills"
								// So full subDir = "skills/" + parts[1]

								baseSubDir := ""
								if !strings.HasPrefix(r.URL, "http") {
									repoParts := strings.Split(r.URL, "/")
									if len(repoParts) > 2 {
										baseSubDir = strings.Join(repoParts[2:], "/")
									}
								}

								if baseSubDir != "" {
									subDir = filepath.Join(baseSubDir, parts[1])
								} else {
									subDir = parts[1]
								}

								skillName = parts[1]
								configMatch = true

								// Trigger background sync for this repo
								go func(name string) {
									// Locate the executable
									exe, err := os.Executable()
									if err != nil {
										return
									}
									// Spawn 'ask repo sync name' detached
									cmd := exec.Command(exe, "repo", "sync", name)
									// Detach process attributes could be OS-specific, but simple Start() usually works for short-lived parent
									// if we don't attach pipes.
									_ = cmd.Start()
									// We don't wait.
								}(r.Name)

								break
							}
						}
					}

					if !configMatch {
						// Standard install: owner/repo
						repoURL = "https://github.com/" + input
						urlParts := strings.Split(strings.TrimSuffix(repoURL, ".git"), "/")
						skillName = urlParts[len(urlParts)-1]
					}
				}
			}
		} else {
			// It's a URL (e.g., https://github.com/xxx.git)
			repoURL = input
			urlParts := strings.Split(strings.TrimSuffix(repoURL, ".git"), "/")
			skillName = urlParts[len(urlParts)-1]
		}
	}

	// SkillHub Slug Resolution
	// If it wasn't a valid GitHub URL and resolved to just a "name" or if we want to support direct slug install:
	// "ask skill install madappgang-claude-code-python"
	// This falls into the "else" of "check if input matches configured repository name" in the caller `Run` loop?
	// The `Run` loop checks configured repos. If not found, it passes `input` directly to `installSingleSkill`.
	// So `installSingleSkill` receives "madappgang-claude-code-python".
	// It goes to `else { Check if it's a direct URL or shorthand }`.
	// `isURL` = false.
	// `parts := strings.Split(input, "/")` -> len=1.
	// So it falls to `else { Standard install: owner/repo }` ? NO.
	// The code assumes input is "owner/repo" if it splits to 2?
	// Wait, let's look at `installSingleSkill` logic again.

	/*
		if !isURL {
			parts := strings.Split(input, "/")
			if len(parts) > 2 {
				// subdir
			} else {
				// owner/repo
				repoURL = "https://github.com/" + input
			}
		}
	*/

	// If input is "slug", `parts` has len 1.
	// The code `else { owner/repo }` logic (lines 262-267) assumes `len(parts) <= 2` handles owner/repo.
	// But if `len(parts) == 1`, `repoURL` becomes `https://github.com/slug`.
	// This is valid for GitHub user profile or org, but not a repo.

	// We need to inject logic: if it looks like a slug (and not owner/repo), try SkillHub resolve.
	// Or try SkillHub resolve if GitHub check fails?
	// Doing it optimistically: if `strings.Contains(input, "/")` is false, it might be a slug.

	if !strings.Contains(input, "/") && !strings.HasPrefix(input, "http") && !strings.HasPrefix(input, "git") {
		// 1. Try resolving from local cache first
		reposCache, err := cache.NewReposCache()
		if err == nil {
			skills, _ := reposCache.SearchSkills(input)

			// Find all exact matches
			var exactMatches []cache.SkillEntry
			for _, s := range skills {
				if s.Name == input {
					exactMatches = append(exactMatches, s)
				}
			}

			if len(exactMatches) > 1 {
				// Ambiguous!
				fmt.Printf("Error: Multiple skills named '%s' found:\n", input)
				for _, m := range exactMatches {
					fmt.Printf("  - %s/%s\n", m.RepoName, m.Name)
				}
				return fmt.Errorf("ambiguous skill name '%s'. Please specify the repository like 'RepoName/SkillName'", input)
			} else if len(exactMatches) == 1 {
				// Single match
				s := exactMatches[0]
				fmt.Printf("found skill '%s' in local cache (repo: %s)\n", input, s.RepoName)

				// We need to resolve the repo URL.
				repoInfos, err := reposCache.LoadIndex()
				if err == nil {
					for _, info := range repoInfos {
						if info.Name == s.RepoName {
							repoURL = info.URL

							// If URL is missing in index (bug fix), lookup from config
							if repoURL == "" && cfg != nil {
								for _, r := range cfg.Repos {
									// Calculate derived name as used in sync
									derivedName := r.Name
									if !strings.HasPrefix(r.URL, "http") {
										parts := strings.Split(r.URL, "/")
										if len(parts) >= 2 {
											derivedName = parts[0] + "-" + parts[1]
										}
									} else {
										derivedName = strings.ReplaceAll(r.URL, "/", "-")
									}

									if r.Name == s.RepoName || derivedName == s.RepoName {
										if !strings.HasPrefix(r.URL, "http") {
											parts := strings.Split(r.URL, "/")
											if len(parts) >= 2 {
												repoURL = fmt.Sprintf("https://github.com/%s/%s.git", parts[0], parts[1])
											}
										} else {
											repoURL = r.URL
										}
										break
									}
								}
							}

							// Calculate relative path (subdir)
							localSourcePath = s.Path
							rel, err := filepath.Rel(info.LocalPath, s.Path)
							if err == nil && rel != "." {
								subDir = rel
							}
							skillName = s.Name
							break
						}
					}
				}
			}
		}

		// 2. If not found in cache, try resolving as SkillHub slug
		if repoURL == "" {
			client := skillhub.NewClient()
			if resolved, err := client.Resolve(input); err == nil {
				fmt.Printf("Resolved SkillHub slug '%s' to '%s'\n", input, resolved)
				return installSingleSkill(resolved, global, agents, cfg)
			}
		}
	}

	// Use branch from version if not set from URL parsing
	if branch == "" && version != "" {
		branch = version
	}

	// Determine target directories based on agents
	var targetDirs []string
	var scopeLabel string

	if len(agents) > 0 {
		// Install to specific agent directories
		for _, agentName := range agents {
			agentType, _ := config.ResolveAgentType(agentName)
			dir, err := config.GetAgentSkillsDir(agentType, global)
			if err != nil {
				return fmt.Errorf("failed to get skills dir for agent %s: %w", agentName, err)
			}
			targetDirs = append(targetDirs, dir)
		}
		scopeLabel = strings.Join(agents, ", ")
		if global {
			scopeLabel += " (global)"
		}
	} else {
		if global {
			targetDirs = []string{config.GetSkillsDirByScope(true)}
			scopeLabel = "global"
		} else {
			// Try to load config to get active/detected directories
			cfg, err := config.LoadConfig()
			if err == nil {
				wd, _ := os.Getwd()
				targetDirs = cfg.GetActiveSkillsDirs(wd)
				scopeLabel = "detected targets"
			} else {
				// Fallback to default if config load fails
				targetDirs = []string{config.GetSkillsDirByScope(false)}
				scopeLabel = "project"
			}
		}
	}

	fmt.Printf("Installing %s to %s...\n", skillName, scopeLabel)

	// Check if already installed in all targets
	allExist := true
	for _, dir := range targetDirs {
		destPath := filepath.Join(dir, skillName)
		if _, err := os.Stat(destPath); os.IsNotExist(err) {
			allExist = false
		}
	}
	if allExist {
		fmt.Printf("Skill %s is already installed in all target directories\n", skillName)
		return nil
	}

	// Clone to a temporary directory first, then copy to each target
	tempDir, err := os.MkdirTemp("", "ask-install-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	tempSkillPath := filepath.Join(tempDir, skillName)

	if localSourcePath != "" {
		if err := copyDir(localSourcePath, tempSkillPath); err != nil {
			return fmt.Errorf("failed to copy local skill: %w", err)
		}
	} else {
		if subDir != "" {
			err = git.InstallSubdir(repoURL, branch, subDir, tempSkillPath)
		} else {
			err = git.Clone(repoURL, tempSkillPath)
		}

		if err != nil {
			return fmt.Errorf("git operation failed: %w", err)
		}
	}

	// Checkout specific version if specified
	if version != "" && subDir == "" {
		fmt.Printf("Checking out version %s...\n", version)
		if err := git.Checkout(tempSkillPath, version); err != nil {
			fmt.Printf("Warning: Failed to checkout version %s: %v\n", version, err)
		}
	}

	// Get current commit for lock file
	var commitHash string
	if subDir == "" {
		commitHash, _ = git.GetCurrentCommit(tempSkillPath)
	}

	// Get skill metadata from SKILL.md
	var skillDescription string
	if skill.FindSkillMD(tempSkillPath) {
		meta, err := skill.ParseSkillMD(tempSkillPath)
		if err == nil && meta != nil && meta.Description != "" {
			skillDescription = meta.Description
		}
	}
	if skillDescription == "" {
		skillDescription = "Skill installed from " + originalInput
	}

	// Central storage location: always store source in .agent/skills/
	centralDir := config.DefaultSkillsDir
	if global {
		centralDir = config.GetSkillsDirByScope(true)
	}
	centralPath := filepath.Join(centralDir, skillName)

	// Check if source already exists in central storage
	sourceExists := false
	if _, err := os.Stat(centralPath); !os.IsNotExist(err) {
		sourceExists = true
	}

	// Copy source to central storage if not exists
	if !sourceExists {
		if err := os.MkdirAll(centralDir, 0755); err != nil {
			return fmt.Errorf("failed to create central skills directory %s: %w", centralDir, err)
		}
		if err := copyDir(tempSkillPath, centralPath); err != nil {
			return fmt.Errorf("failed to copy skill to central storage %s: %w", centralPath, err)
		}
	}

	// Create symlinks (or copy as fallback) to each target directory
	for _, dir := range targetDirs {
		destPath := filepath.Join(dir, skillName)

		// Skip central storage itself (no self-link needed)
		if destPath == centralPath {
			continue
		}

		// Skip if already exists
		if _, err := os.Stat(destPath); !os.IsNotExist(err) {
			if isSymlink(destPath) {
				fmt.Printf("  → Already linked in %s\n", destPath)
			} else {
				fmt.Printf("  → Already installed in %s\n", destPath)
			}
			continue
		}

		// Ensure target directory exists
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create skills directory %s: %w", dir, err)
		}

		// Create symlink to central storage (with copy fallback for Windows)
		if err := createSymlinkOrCopy(centralPath, destPath); err != nil {
			return fmt.Errorf("failed to link skill to %s: %w", destPath, err)
		}

		if isSymlink(destPath) {
			fmt.Printf("  → Linked to %s\n", destPath)
		} else {
			fmt.Printf("  → Copied to %s (symlink not supported)\n", destPath)
		}
	}

	// Update config
	updatedCfg, err := config.LoadConfigByScope(global)
	if err == nil {
		skillInfo := config.SkillInfo{
			Name:        skillName,
			Description: skillDescription,
			URL:         repoURL,
		}
		if subDir != "" {
			// Avoid duplicating https://github.com/ prefix
			if strings.HasPrefix(input, "https://") || strings.HasPrefix(input, "http://") {
				skillInfo.URL = input
			} else {
				skillInfo.URL = fmt.Sprintf("https://github.com/%s", input)
			}
		}

		updatedCfg.AddSkillInfo(skillInfo)
		err = updatedCfg.SaveByScope(global)
		if err != nil {
			configFile := "ask.yaml"
			if global {
				configFile = "~/.ask/config.yaml"
			}
			fmt.Printf("Warning: Failed to update %s: %v\n", configFile, err)
		}

		// Update lock file
		lockFile, _ := config.LoadLockFileByScope(global)
		lockEntry := config.LockEntry{
			Name:        skillName,
			URL:         skillInfo.URL,
			Commit:      commitHash,
			Version:     version,
			InstalledAt: time.Now(),
		}
		lockFile.AddEntry(lockEntry)
		if err := lockFile.SaveByScope(global); err != nil {
			lockFileName := "ask.lock"
			if global {
				lockFileName = "~/.ask/ask.lock"
			}
			fmt.Printf("Warning: Failed to update %s: %v\n", lockFileName, err)
		}
	} else if !global {
		// If config doesn't exist for project-level, we might be in a non-init project
		fmt.Println("Warning: ask.yaml not found. Run 'ask init' to track dependencies.")
	}

	fmt.Printf("Successfully installed %s!\n", skillName)
	return nil
}

// copyDir recursively copies a directory from src to dst
func copyDir(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}
	return nil
}

// copyFile copies a single file from src to dst
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = srcFile.Close() }()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return err
	}
	defer func() { _ = dstFile.Close() }()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// createSymlinkOrCopy creates a symlink from target to source, or falls back to copy on failure.
// Uses relative paths for portability. Works on Linux, macOS, and Windows (with fallback).
func createSymlinkOrCopy(source, target string) error {
	// Calculate relative path from target's directory to source
	targetDir := filepath.Dir(target)
	relPath, err := filepath.Rel(targetDir, source)
	if err != nil {
		// If relative path fails, fall back to copy
		return copyDir(source, target)
	}

	// Try creating symlink
	if err := os.Symlink(relPath, target); err != nil {
		// Fallback to copy on Windows or permission errors
		return copyDir(source, target)
	}
	return nil
}

// isSymlink checks if the given path is a symbolic link
func isSymlink(path string) bool {
	info, err := os.Lstat(path)
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeSymlink != 0
}

func registerInstallFlags(cmd *cobra.Command) {
	cmd.Flags().StringSliceP("agent", "a", []string{}, "target agent(s) (claude->.claude/skills, cursor->.cursor/skills, etc.)")
	cmd.Flags().BoolP("global", "g", false, "install globally (user-level)")
}

func init() {
	skillCmd.AddCommand(installCmd)
	registerInstallFlags(installCmd)
}
