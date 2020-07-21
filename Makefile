include ./docker/.env

deploy:
ifeq ($(API_PASSWD),)
	$(error "env var API_PASSWD is empty")
endif

ifneq ($(wildcard $(PWD)/src/api/bin/api),)
	@echo "Founded api bin"
else
	$(error "api bin file doesn't exists")
endif
	@echo "Checking docker-compose config"
	@cd ./docker && docker-compose config >> /dev/null

	@echo "Checking for needed docker images"
	@docker inspect --type=image api_refresher_img >> /dev/null

	@echo "Starting ..."
	@cd ./docker && docker-compose up -d
all:
	cd src/api/src && go build -v -o ../bin/api .
	docker build -t api_refresher_img:latest -f docker/img/api_refresher ./src
go:
	cd src/api/src && go build -v -o ../bin/api .
deps:
	go get -v go.mongodb.org/mongo-driver/mongo
	go get -v github.com/valyala/fasthttp
	go get -v github.com/AubSs/fasthttplogger
img:
	docker build -t api_refresher_img:latest -f docker/img/api_refresher ./src/refresher
distclean:
	rm -rvf src/api/bin
clean:
	@docker rmi api_refresher_img:latest
	@rm -rvf src/api/bin
