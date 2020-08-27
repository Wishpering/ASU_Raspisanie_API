include ./docker/.env

deploy:
ifeq ($(API_PASSWD),)
	$(error "env var API_PASSWD is empty")
endif

	@echo "Checking docker-compose config"
	@cd ./docker && docker-compose config >> /dev/null

	@echo "Checking for needed docker images"
	@docker inspect --type=image api_refresher_img >> /dev/null
	@docker inspect --type=image api_img >> /dev/null	

	@echo "Starting ..."
	@cd ./docker && docker-compose up -d
img:
	docker build -t api_refresher_img:latest -f docker/img/api_refresher_img ./src/refresher
	docker build -t api_img:latest -f docker/img/api_img ./src/api
clean:
	docker rmi api_img:latest
	docker rmi api_refresher_img:latest
	@rm -rvf src/api/bin
