run:
	docker run osniantonio/stresstest:latest --url=http://google.com --requests=1000 --concurrency=10 --insecure