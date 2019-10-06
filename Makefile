build:
	docker build -t keybase-bot-weather .

run:
	docker run -it -e KEYBASE_USERNAME="$(KEYBASE_USERNAME)" -e KEYBASE_PAPERKEY="$(KEYBASE_PAPERKEY)" keybase-bot-weather

deploy-image:
	gcloud --project=keybasebots builds submit --tag gcr.io/keybasebots/weather
