
all: release

release:
	cd cmd; go build -o goisgod -installsuffix .
