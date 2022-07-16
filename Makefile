workers ?= 300
dir ?= ../clones

list:
	@ cat Makefile

self-referential-commit:
	go build

i.clean:
	go clean
	rm -rf nohup.out pid ../clones

i.tail:
	tail -f nohup.out

i.run: i.clean self-referential-commit
	nohup ./self-referential-commit -w $(workers) -d $(dir) &

i.ps:
	ps ax | grep self-referential-commit
