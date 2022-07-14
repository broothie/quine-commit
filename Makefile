
list:
	@ cat Makefile

self-referential-commit:
	go build

i.run: self-referential-commit
	nohup ./self-referential-commit -w 500 -d ../clones &

i.reset:
	rm -rf nohup.out pid ../clones

i.stop:
	kill $(cat pid)
