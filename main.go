package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

const shortSHALength = 7

var (
	cloneDirectory  = flag.String("d", "clones", "clone directory")
	workers         = flag.Int("w", 3, "number of workers")
	refreshInterval = flag.Int("r", 1000, "refresh every r attempts")
	logInterval     = flag.Int("l", 100, "log every l attempts")
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	start := time.Now()
	flag.Parse()

	// Path
	workingDirectory, err := os.Getwd()
	if err != nil {
		fmt.Println("failed to get working directory", err)
		os.Exit(1)
		return
	}

	clonePath, err := filepath.Abs(*cloneDirectory)
	if err != nil {
		fmt.Println("failed to get absolute path of clone dir", err)
		os.Exit(1)
		return
	}

	clonePath = filepath.Join(clonePath, strconv.Itoa(int(start.Unix())))

	// Async
	doneChan := make(chan struct{})
	errChan := make(chan error)

	// Start workers
	worker := 0
	for ; worker < *workers; worker++ {
		go findLuckySHA(worker, workingDirectory, clonePath, doneChan, errChan)
		time.Sleep(100 * time.Millisecond)
	}

	go func() {
		for err := range errChan {
			fmt.Println("error", err)
			go findLuckySHA(worker, workingDirectory, clonePath, doneChan, errChan)
			worker++
			time.Sleep(100 * time.Millisecond)
		}
	}()

	// Wait and cancel
	<-doneChan
}

func findLuckySHA(worker int, remotePath, clonePath string, doneChan chan struct{}, errChan chan error) {
	repoPath := filepath.Join(clonePath, fmt.Sprintf("%d-quine-commit", worker))

	for attempts := 0; ; attempts++ {
		start := time.Now()
		shouldRefresh := attempts%*refreshInterval == 0
		shouldLog := attempts%*logInterval == 0

		if shouldRefresh {
			if err := os.RemoveAll(repoPath); err != nil {
				errChan <- errors.Wrap(err, "failed to remove repo")
				return
			}

			if err := gitCloneLocal(remotePath, repoPath); err != nil {
				errChan <- errors.Wrapf(err, "failed to git clone %q", repoPath)
				return
			}
		}

		shortSha := randomShortSHA(shortSHALength)
		if err := gitCommit(repoPath, shortSha); err != nil {
			errChan <- err
			return
		}

		outputSha, err := gitRevParse(repoPath, shortSHALength)
		if err != nil {
			errChan <- err
			return
		}

		if shortSha == outputSha {
			message := fmt.Sprintf("success! the lucky sha was %s from worker %d", shortSha, worker)
			fmt.Println(message)
			if err := os.WriteFile(filepath.Join(repoPath, "short.sha"), []byte(message), 0666); err != nil {
				fmt.Println("failed to write file under repo", repoPath, err)
				return
			}

			close(doneChan)
		} else {
			if err := gitReset(repoPath); err != nil {
				errChan <- err
				return
			}
		}

		if shouldLog {
			fmt.Println(shortSha, "!=", outputSha, "worker", worker, "attempt", attempts, "elapsed", time.Since(start))
		}
	}
}

func gitClone(repoPath string) error {
	return gitCloneLocal("https://github.com/broothie/quine-commit.git", repoPath)
}

func gitCloneLocal(remotePath, repoPath string) error {
	output, err := exec.Command("git", "clone", remotePath, repoPath).CombinedOutput()
	fmt.Print(string(output))
	if err != nil {
		return errors.Wrapf(err, "failed to clone git repo at %q", repoPath)
	}

	return nil
}

func gitCommit(repoPath, message string) error {
	output, err := exec.Command("git", "-C", repoPath, "commit", "--allow-empty", "-m", message).CombinedOutput()
	if err != nil {
		fmt.Print(string(output))
		return errors.Wrapf(err, "failed to commit to repo at %q", repoPath)
	}

	return nil
}

func gitRevParse(repoPath string, length int) (string, error) {
	output, err := exec.Command("git", "-C", repoPath, "rev-parse", fmt.Sprintf("--short=%d", length), "HEAD").CombinedOutput()
	if err != nil {
		fmt.Println(string(output))
		return "", errors.Wrapf(err, "failed to parse revision at %q", repoPath)
	}

	return strings.TrimSpace(string(output)), nil
}

func gitReset(repoPath string) error {
	output, err := exec.Command("git", "-C", repoPath, "reset", "--hard", "HEAD~").CombinedOutput()
	if err != nil {
		fmt.Print(string(output))
		return errors.Wrapf(err, "failed to reset repo at %q", repoPath)
	}

	return nil
}

func randomShortSHA(length int) string {
	const hexRunes = "0123456789abcdef"

	runes := make([]rune, length)
	for i := range runes {
		runes[i] = rune(hexRunes[rand.Intn(len(hexRunes))])
	}

	return string(runes)
}
