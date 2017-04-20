simplcheck: ./cmd/* ./lib/*
	cd cmd && CGO_ENABLED=0 go build -o ../simplcheck

docker: simplcheck
	docker build -t gabesullice/simplcheck .

push: docker
	docker push gabesullice/simplcheck

.PHONY: docker
