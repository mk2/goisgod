
GO = go

all: release

release:
	cd cmd; $(GO) build -o goisgod -installsuffix .
