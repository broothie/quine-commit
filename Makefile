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
	echo $! > pid

kill:
	bash -c 'kill $(cat pid)'

ps:
	ps ax | grep self-referential-commit

ssh:
	gcloud compute ssh --zone us-central1-a --project andrewb-general instance-4
