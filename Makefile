workers ?= 3
dir ?= ../clones

list:
	@ cat Makefile

quine-commit:
	go build

clean:
	go clean
	rm -rf nohup.out ../clones

tail:
	tail -f nohup.out

run: clean quine-commit
	./quine-commit -w $(workers) -d $(dir)

nohup: clean quine-commit
	nohup ./quine-commit -w $(workers) -d $(dir) &

ps:
	@ ps -ax | grep quine-commit | grep -v grep ||:

gcloud:
	gcloud compute ssh --zone us-central1-a --project andrewb-general instance-4

pi:
	ssh raspberrypi.local
