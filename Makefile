workers ?= 3
dir ?= ../clones

list:
	@ cat Makefile

self-referential-commit:
	go build

clean:
	go clean
	rm -rf nohup.out ../clones

tail:
	tail -f nohup.out

run: clean self-referential-commit
	./self-referential-commit -w $(workers) -d $(dir)

nohup: clean self-referential-commit
	nohup ./self-referential-commit -w $(workers) -d $(dir) &
	echo $$! > pid

kill:
	kill $$(cat pid)
	rm pid

ps:
	ps -ax | grep self-referential-commit | grep -v grep

ssh:
	gcloud compute ssh --zone us-central1-a --project andrewb-general instance-4
