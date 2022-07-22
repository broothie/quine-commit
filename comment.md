# Welcome to the Quine Commit!

This commit represents the output of an attempt to generate a commit where the ***commit message is the short SHA of the commit itself***.

Go ahead: click on the commit message.

## How it went down

The initial inspiration came from the [quine tweet](https://twitter.com/quinetweet/status/1309951041321013248).

### How it was built

You can probably guess what the strategy was here: brute forcing guessing the short SHA. This project has a dependency on how GitHub works as well, since it leverages GitHub's [commit SHA auto-linking](https://docs.github.com/en/get-started/writing-on-github/working-with-advanced-formatting/autolinked-references-and-urls#commit-shas). I did a quick test to figure out what the minimum number of characters required to trigger short SHA auto-linking is: it appears to be 7.

Armed with this knowledge, I wrote a [Ruby script](https://github.com/broothie/quine-commit/commit/9c22ebcc8890757942b05b91a866d4bfb46581c6), trashed it, [re-wrote it in Go](https://github.com/broothie/quine-commit/commit/a9f5169993ef50b8d78c9959ecf3c741f2980795), and iterated on that. As much as I wanted to play around with [Ruby's new-ish async stuff](https://brunosutic.com/blog/async-ruby), I'm much more comfortable with Go's concurrency patterns. Plus, my guess is the little bump in speed would come in handy in the long run.

The final version of the program does the following:
- spawn `n` workers
- each worker gets a unique path on the filesystem
- each worker initializes a repo at their path
- then, in a loop:
    - generate a 7-character hexadecimal string
    - make an empty commit with the hex string as the commit message
    - `git rev-parse --short=7`
        - if it's a match:
            - celebrate! 🎉
            - also, be sure to print it (_and_ write it to a file to be safe)
            - then bail
        - if not:
            - `git reset --hard HEAD~`
    - occasionally, remove the repo and re-`git init`

### How it was run

I didn't want to run the script on my own machine because I didn't want to go through the hassle of **keeping it plugged in** and/or **figuring out how to keep the script running** and/or **confirming that it was actually running all the time**. Plus, I figured I could could get a fancier rig from ***the cloud*** ☁️.

Up on GCP, I went through 4 instances while trying to figure out the right setup. IIRC the first few were either too slow in disk ops, or had too little CPU. Running trials with 50 to 200 workers would either blow out the CPU and have all the goroutines fighting over time, or cause the single-iteration time to be too high e.g. on the order of 10s of seconds.

I eventually settled on GCP's lowest tier compute-optimized instance, `c2-standard-4`, which was heavily based on the fact that it seems like you can only attach local disks to compute- or memory-optimized (or GPU) instances. By this point I was operating under the assumption that I was wasting precious time talking to the network attached disks these instances have by default, so a local SSD seemed necessary.

On this box, I was able to run 300 workers, hitting almost exactly 80% CPU usage, and iterating roughly every 750ms. With 7 hex digits, we have a 1 in `16^7 = 268,435,456` chance of guessing the short SHA. So:
```
> 16**7 * 0.750 / 60 / 60 / 24 / 300
7.76722962962963
```
i.e. I should be expecting a result within ~7¾ days.

How did it actually shake out?
```
$ stat nohup.out
...
Modify: 2022-07-19 19:29:40.561706164 +0000
Change: 2022-07-19 19:29:40.561706164 +0000
 Birth: 2022-07-16 09:03:44.409662694 +0000
```
**3.434675925925926 days!** Pretty good!

## Things I learned and open questions

### `git reset --hard` doesn't completely get rid of your changes

This became clear after two things:
- I started seeing messages about `git` automatically running `git gc`
- I checked the `reflog` and, well, there're at least some remnants of the reset commits in there

I briefly had a version of the program which would run `git gc` after every `reset`, but that didn't seem to help either upon checking the reflog, so I ended up adding an occasional nuke and re-`init` of the repo.

But then, how can you *truly* get rid of your changes without `rm-rf`-ing? (I'm sure there's a way, I just haven't gotten around to Googling it).

### Did I really need 300 workers?

I made a lot of assumptions during this project, and this is *kindof* one of them. I did do a bit of tuning when starting up the program and seeing the average iteration time of each worker. It seemed like iteration time scaled more with number of workers than with (what I imagine is) disk usage.

That would be the other variable here right? With more workers comes more disk operations, and I would guess the bottleneck would then come from workers waiting on disk IO.

### What is the *true* meaning of [df2128c](https://github.com/broothie/quine-commit/commit/df2128c1b3fed98d646d86911adba677a97165ad)?

D? F? 21? 28?? C??!?!1 What is the significance of these numbers and letters? We may never know ¯\_(ツ)_/¯
