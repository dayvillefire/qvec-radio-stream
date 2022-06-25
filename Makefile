BINARY := qvec-radio-stream

all: binary

clean:
	go clean -v

binary: clean
	GOARCH=386 GOOS=linux go build -v

