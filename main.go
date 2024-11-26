package main

import (
	"fmt"
	"net/http"
	"os"
	"time"
	"math"
	"sync"
	"sync/atomic"
)

// Result structure for processed combinations
type Result struct {
	Combination string
}

// Progress tracks processing progress
type Progress struct {
	current   int64
	total     int64
	startTime time.Time
}

func (p *Progress) increment() {
	atomic.AddInt64(&p.current, 1)
}

func (p *Progress) display() {
	current := atomic.LoadInt64(&p.current)
	percentage := float64(current) * 100 / float64(p.total)
	elapsed := time.Since(p.startTime)
	speed := float64(current) / elapsed.Seconds()
	remaining := time.Duration(float64(p.total-current)/speed) * time.Second

	fmt.Printf("\rProgress: %.2f%% (%d/%d) Speed: %.1f/s ETA: %v", 
		percentage, current, p.total, speed, remaining.Round(time.Second))
}

// Function to check username availability on different platforms
func checkUsernameAvailability(username string) bool {
	// Check GitHub
	if !checkGitHub(username) {
		return false
	}
	// Check Twitter (X)
	if !checkTwitter(username) {
		return false
	}
	// Check LinkedIn
	if !checkLinkedIn(username) {
		return false
	}
	// Check Instagram
	if !checkInstagram(username) {
		return false
	}
	return true
}

// GitHub username check
func checkGitHub(username string) bool {
	url := fmt.Sprintf("https://github.com/%s", username)
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != 404 {
		return false
	}
	return true
}

// Twitter (X) username check
func checkTwitter(username string) bool {
	url := fmt.Sprintf("https://twitter.com/%s", username)
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != 404 {
		return false
	}
	return true
}

// LinkedIn username check
func checkLinkedIn(username string) bool {
	url := fmt.Sprintf("https://www.linkedin.com/in/%s", username)
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != 404 {
		return false
	}
	return true
}

// Instagram username check
func checkInstagram(username string) bool {
	url := fmt.Sprintf("https://www.instagram.com/%s/", username)
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != 404 {
		return false
	}
	return true
}

// Generate 3-character combinations and check availability (CPU)
func generateCombinationsCPU(results chan<- string, progress *Progress) {
	const alphabet = "abcdefghijklmnopqrstuvwxyz"
	alphabetLen := len(alphabet)

	// Calculate total combinations (3-character combinations)
	total := int(math.Pow(float64(alphabetLen), float64(3)))

	// Generate 3-character combinations and check each username
	for i := 0; i < total; i++ {
		combination := ""
		// Generate the combination based on the index
		temp := i
		for j := 0; j < 3; j++ {
			combination = string(alphabet[temp%alphabetLen]) + combination
			temp /= alphabetLen
		}

		// Check availability on all platforms
		if checkUsernameAvailability(combination) {
			// Send the combination to the channel if it's available on all platforms
			results <- combination
		}

		// Update progress
		progress.increment()
	}
	close(results)
}

func main() {
	// Configuration
	const (
		workerCount = 1000
		resultsFile = "common_usernames.txt"
	)

	// Initialize progress tracking
	progress := &Progress{
		total:     int64(math.Pow(26, 3)), // Only 3-character combinations
		startTime: time.Now(),
	}

	// Start progress display
	done := make(chan bool)
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				progress.display()
			case <-done:
				progress.display()
				fmt.Println()
				return
			}
		}
	}()

	// Initialize channels
	resultChan := make(chan Result, workerCount*2)
	var results []Result

	// Process combinations for length 3 only
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		combinations := make(chan string, 1000)

		// Generate combinations using CPU
		go generateCombinationsCPU(combinations, progress)

		// Process generated combinations
		for combination := range combinations {
			resultChan <- Result{
				Combination: combination,
			}
		}
	}()

	// Collect results
	var resultWg sync.WaitGroup
	resultWg.Add(1)
	go func() {
		defer resultWg.Done()
		for result := range resultChan {
			results = append(results, result)
		}
	}()

	// Wait for all processing to complete
	wg.Wait()
	close(resultChan)
	resultWg.Wait()
	done <- true

	// Save results to a .txt file
	file, err := os.Create(resultsFile)
	if err != nil {
		fmt.Printf("Error creating results file: %v\n", err)
		return
	}
	defer file.Close()

	// Write common usernames to the file (one per line)
	for _, result := range results {
		_, err := file.WriteString(result.Combination + "\n")
		if err != nil {
			fmt.Printf("Error writing to file: %v\n", err)
			return
		}
	}

	fmt.Printf("\nProcessed %d combinations. Results saved to %s\n", 
		len(results), resultsFile)
}
