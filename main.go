package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

var (
	workers  = flag.Int("w", 1, "number of workers")
	logEvery = flag.Int("l", 10, "log every l attempts")
)

func init() {
	rand.Seed(time.Now().UnixNano())
	flag.Parse()
}

func main() {
	start := time.Now()
	if err := os.RemoveAll("clones"); err != nil {
		panic(err)
	}

	if err := os.RemoveAll("short.sha"); err != nil {
		panic(err)
	}

	// Async
	ctx, cancel := context.WithCancel(context.Background())
	group, ctx := errgroup.WithContext(ctx)
	resultChan := make(chan string)

	// Start workers
	for worker := 0; worker < *workers; worker++ {
		group.Go(findLuckySHA(ctx, start, worker, resultChan))
	}

	// Wait and cancel
	fmt.Println(<-resultChan)
	cancel()
	if err := group.Wait(); err != nil {
		panic(err)
	}
}

func findLuckySHA(ctx context.Context, start time.Time, worker int, resultChan chan string) func() error {
	return func() error {
		repoPath := path.Join("clones", strconv.Itoa(int(start.Unix())), fmt.Sprintf("%d-self-referential-commit", worker))
		if err := gitClone(repoPath); err != nil {
			return errors.Wrapf(err, "failed to git init %q", repoPath)
		}

		attempts := 0
		for {
			select {
			case <-ctx.Done():
				return nil
			default:
				attemptStart := time.Now()

				shortSha := randomShortSHA()
				message := fmt.Sprintf("short sha: %s", shortSha)
				output, err := gitCommit(repoPath, message)
				if err != nil {
					return err
				}

				if strings.TrimSpace(output) == fmt.Sprintf("[main %s] %s", shortSha, message) {
					resultChan <- fmt.Sprintf("success! the lucky sha was %s from worker %d", shortSha, worker)
					close(resultChan)

					if err := os.WriteFile("short.sha", []byte(shortSha), 0666); err != nil {
						return errors.Wrapf(err, "failed to write file under repo %q", repoPath)
					}
				} else {
					if err := gitReset(repoPath); err != nil {
						return err
					}

					if err := gitGC(repoPath); err != nil {
						return err
					}
				}

				if attempts%*logEvery == 0 {
					fmt.Println("worker", worker, "attempt", attempts, "elapsed", time.Since(attemptStart))
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

func gitCommit(repoPath, message string) (string, error) {
	output, err := exec.Command("git", "-C", repoPath, "commit", "--allow-empty", "-m", message).CombinedOutput()
	if err != nil {
		fmt.Print(string(output))
		return "", errors.Wrapf(err, "failed to commit to repo at %q", repoPath)
	}

	return string(output), nil
}

func gitReset(repoPath string) error {
	output, err := exec.Command("git", "-C", repoPath, "reset", "--hard", "HEAD~").CombinedOutput()
	if err != nil {
		fmt.Print(string(output))
		return errors.Wrapf(err, "failed to reset repo at %q", repoPath)
	}

	return nil
}

func gitGC(repoPath string) error {
	output, err := exec.Command("git", "-C", repoPath, "gc").CombinedOutput()
	if err != nil {
		fmt.Print(string(output))
		return errors.Wrapf(err, "failed to gc repo at %q", repoPath)
	}

	return nil
}

func randomShortSHA() string {
	const hexRunes = "0123456789abcdef"

	runes := make([]rune, 7)
	for i := range runes {
		runes[i] = rune(hexRunes[rand.Intn(len(hexRunes))])
	}

	return string(runes)
}
