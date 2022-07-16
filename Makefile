
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
	nohup ./self-referential-commit -w 300 -d ../clones &

i.ps:
	ps ax | grep self-referential-commit
