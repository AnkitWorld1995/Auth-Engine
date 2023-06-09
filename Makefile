
run:
	go run main.go

build:
	set GOOS=linux
	set GOARCH=amd64
	set CGO_ENABLED=0
	go build -o main main.go

zip:
	zip -jrm build/Authenication-SVC.zip build/Authenication-SVC