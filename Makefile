all:
	cd src/api/src && go build -v -o ../bin/api .
	docker build -t api_refresher_img:latest -f docker/refresher_image ./src
go:
	cd src/api/src && go build -v -o ../bin/api .
deps:
	go get -v go.mongodb.org/mongo-driver/mongo
	go get -v github.com/valyala/fasthttp
img:
	docker build -t api_refresher_img:latest -f docker/refresher_image ./src
clean:
	rm -rvf src/api/bin
