autodeploy:
	CGO_ENABLED=0 go build -o autodeploy main.go

clean:
	rm -f autodeploy

run-server: clean autodeploy
	go install github.com/joho/godotenv/cmd/godotenv@latest 
	godotenv -f autodeploy.env ./autodeploy \
		--port 8090 \
		--dev-logging