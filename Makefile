
list:
	@ cat Makefile

i.clean: i.reset
	go clean

self-referential-commit:
	go build

i.tail:
	tail -f nohup.out

i.run: self-referential-commit
	nohup ./self-referential-commit -w 1000 -d ../clones &

i.reset:
	rm -rf nohup.out pid ../clones

i.ps:
	ps ax | grep self-referential-commit
