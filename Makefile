workers ?= 300
dir ?= ../clones

list:
	@ cat Makefile

self-referential-commit:
	go build

clean:
	go clean
	rm -rf nohup.out pid ../clones

tail:
	tail -f nohup.out

run: clean self-referential-commit
	nohup ./self-referential-commit -w $(workers) -d $(dir) &

ps:
	ps ax | grep self-referential-commit

ssh:
	gcloud compute ssh --zone us-central1-a --project andrewb-general instance-4
