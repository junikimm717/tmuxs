UNAME := $(shell uname)
ifeq ($(UNAME), Darwin)
	FLAGS=-ldflags '-s -extldflags "-sectcreate __TEXT __info_plist Info.plist"'
else
	FLAGS=-tags netgo -ldflags '-extldflags "-static"'
endif

tmuxs:
	go build $(FLAGS) -o tmuxs .

clean:
	rm -rf tmuxs

all: tmuxs
make: tmuxs
