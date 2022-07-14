package main

import (
	"context"
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
	"golang.org/x/sync/errgroup"
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
	wd, err := os.Getwd()
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

	// Async
	ctx, cancel := context.WithCancel(context.Background())
	group, ctx := errgroup.WithContext(ctx)
	resultChan := make(chan string)

	// Start workers
	for worker := 0; worker < *workers; worker++ {
		group.Go(findLuckySHA(ctx, worker, wd, filepath.Join(clonePath, strconv.Itoa(int(start.Unix()))), resultChan))
		time.Sleep(100 * time.Millisecond)
	}

	// Wait and cancel
	fmt.Println(<-resultChan)
	cancel()
	if err := group.Wait(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func findLuckySHA(ctx context.Context, worker int, remotePath, clonePath string, resultChan chan string) func() error {
	return func() error {
		repoPath := filepath.Join(clonePath, fmt.Sprintf("%d-self-referential-commit", worker))

		attempts := 0
		for {
			shouldRefresh := attempts%*refreshInterval == 0
			shouldLog := attempts%*logInterval == 0

			select {
			case <-ctx.Done():
				return nil
			default:
				attemptStart := time.Now()

				if shouldRefresh {
					if err := os.RemoveAll(repoPath); err != nil {
						return errors.Wrap(err, "failed to remove repo")
					}

					if err := gitCloneLocal(remotePath, repoPath); err != nil {
						return errors.Wrapf(err, "failed to git clone %q", repoPath)
					}
				}

				shortSha := randomShortSHA(shortSHALength)
				if err := gitCommit(repoPath, shortSha); err != nil {
					return err
				}

				outputSha, err := gitRevParse(repoPath, shortSHALength)
				if err != nil {
					return err
				}

				if shortSha == outputSha {
					message := fmt.Sprintf("success! the lucky sha was %s from worker %d", shortSha, worker)
					if err := os.WriteFile(filepath.Join(repoPath, "short.sha"), []byte(message), 0666); err != nil {
						return errors.Wrapf(err, "failed to write file under repo %q", repoPath)
					}

					resultChan <- message
					close(resultChan)
				} else {
					if err := gitReset(repoPath); err != nil {
						return err
					}
				}

				if shouldLog {
					fmt.Println(shortSha, "!=", outputSha, "worker", worker, "attempt", attempts, "elapsed", time.Since(attemptStart))
				}
			}

			attempts += 1
		}
	}
}

func gitClone(repoPath string) error {
	output, err := exec.Command("git", "clone", "https://github.com/broothie/self-referential-commit.git", repoPath).CombinedOutput()
	fmt.Print(string(output))
	if err != nil {
		return errors.Wrapf(err, "failed to init git repo at %q", repoPath)
	}

	return nil
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
